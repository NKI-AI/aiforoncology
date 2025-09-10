package importer

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NKI-AI/aiforoncology/src/golang/research/dcmw/internal/models"
	"github.com/NKI-AI/aiforoncology/src/golang/research/dcmw/internal/processor"
	"gorm.io/gorm"
)

type DICOMFilesImporter struct {
	DicomDir  string
	DB        *gorm.DB
	Threads   int
	BatchSize int
	Debug     bool
	Depth     int

	fileCh chan string
	dataCh chan *processor.ProcessedData

	// Add caches to reduce database queries
	patientCache  sync.Map // map[string]*models.Patient
	studyCache    sync.Map // map[string]*models.Study
	seriesCache   sync.Map // map[string]*models.Series
	instanceCache sync.Map // map[string]bool - just track existence
}

func NewDICOMFilesImporter(dicomDir string, db *gorm.DB, threads int, batchSize int, debug bool, depth int) *DICOMFilesImporter {
	if threads <= 0 {
		threads = runtime.NumCPU()
	}

	return &DICOMFilesImporter{
		DicomDir:  dicomDir,
		DB:        db,
		Threads:   threads,
		BatchSize: batchSize,
		Debug:     debug,
		Depth:     depth,
		fileCh:    make(chan string, 10000),
		dataCh:    make(chan *processor.ProcessedData, 10000),
	}
}

func DisplayProgressBar(processedCount *int64, done chan struct{}) {
	symbols := []string{"|", "/", "-", "\\"} // Rotating symbols for visual effect
	index := 0
	startTime := time.Now()

	for {
		select {
		case <-done:
			fmt.Println() // Move to the next line when done
			return
		default:
			processed := atomic.LoadInt64(processedCount)
			duration := time.Since(startTime).Seconds()
			filesPerSecond := float64(processed) / duration

			// Print progress bar without interfering with logging
			fmt.Printf("\r%s Processed %d files (%.2f files/second)", symbols[index], processed, filesPerSecond)
			fmt.Print(" ") // Add space to clear any leftover characters

			// Update symbol index
			index = (index + 1) % len(symbols)

			time.Sleep(500 * time.Millisecond) // Refresh rate
		}
	}
}

func (importer *DICOMFilesImporter) ProcessFolder() {
	startTime := time.Now()

	// Counter to track processed files
	var processedCount int64

	done := make(chan struct{})
	go DisplayProgressBar(&processedCount, done)

	// Start the cache cleanup routine
	go importer.cleanupCaches()

	go importer.walkFiles()

	// Start the writer goroutine
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		importer.writer()
	}()

	// Start worker goroutines
	var workerWg sync.WaitGroup
	for i := 0; i < importer.Threads; i++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			importer.worker(&processedCount)
		}()
	}

	// Wait for all workers to finish
	workerWg.Wait()
	close(importer.dataCh) // Close dataCh after workers finish

	// Wait for writer to finish
	writerWg.Wait()
	close(done) // Signal progress bar to finish

	log.Printf("Completed processing directory %s in %v with a total of %d files.\n",
		importer.DicomDir, time.Since(startTime), atomic.LoadInt64(&processedCount))
}

func (importer *DICOMFilesImporter) walkFiles() {
	defer close(importer.fileCh)
	err := filepath.Walk(importer.DicomDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check depth
		if importer.Depth >= 0 {
			relativePath, err := filepath.Rel(importer.DicomDir, path)
			if err != nil {
				return err
			}

			if strings.Count(relativePath, string(filepath.Separator)) > importer.Depth {
				return nil // Skip deeper files
			}
		}

		if !info.IsDir() && info.Mode().IsRegular() {
			importer.fileCh <- path
		}

		return nil
	})
	if err != nil {
		log.Printf("Error walking the path %q: %v\n", importer.DicomDir, err)
	}
}

func (importer *DICOMFilesImporter) worker(processedCount *int64) {
	for filePath := range importer.fileCh {
		processedData, err := processor.ProcessOneDICOMFile(filePath, importer.DB, importer.Debug)
		if err != nil {
			if strings.Contains(err.Error(), "DICOM dataset is nil") {
				log.Printf("Warning: File not readable as DICOM: %s", filePath)
			} else {
				log.Printf("Error processing file %s: %v", filePath, err)
			}
			continue
		}

		// Additional validation
		if processedData == nil {
			log.Printf("Warning: No data extracted from file: %s", filePath)
			continue
		}

		// Validate required fields are present
		if processedData.Patient == nil || processedData.Patient.PatientMRN == "" {
			log.Printf("Warning: Missing patient information in file: %s", filePath)
			continue
		}
		if processedData.Study == nil || processedData.Study.StudyInstanceUID == "" {
			log.Printf("Warning: Missing study information in file: %s", filePath)
			continue
		}
		if processedData.Series == nil || processedData.Series.SeriesInstanceUID == "" {
			log.Printf("Warning: Missing series information in file: %s", filePath)
			continue
		}
		if processedData.Instance == nil || processedData.Instance.SOPInstanceUID == "" {
			log.Printf("Warning: Missing instance information in file: %s", filePath)
			continue
		}

		importer.dataCh <- processedData
		atomic.AddInt64(processedCount, 1)
	}
}

func (importer *DICOMFilesImporter) writer() {
	batch := make([]*processor.ProcessedData, 0, importer.BatchSize)
	timeout := time.NewTimer(1 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case data, ok := <-importer.dataCh:
			if !ok {
				if len(batch) > 0 {
					importer.writeBatch(batch)
				}
				return
			}
			batch = append(batch, data)
			if len(batch) >= importer.BatchSize {
				importer.writeBatch(batch)
				batch = make([]*processor.ProcessedData, 0, importer.BatchSize)
				timeout.Reset(1 * time.Second)
			}
		case <-timeout.C:
			if len(batch) > 0 {
				importer.writeBatch(batch)
				batch = make([]*processor.ProcessedData, 0, importer.BatchSize)
			}
			timeout.Reset(1 * time.Second)
		}
	}
}

func (importer *DICOMFilesImporter) writeBatch(batch []*processor.ProcessedData) {
	tx := importer.DB.Begin()

	// Prepare statements for better performance
	patientStmt := tx.Session(&gorm.Session{PrepareStmt: true})
	studyStmt := tx.Session(&gorm.Session{PrepareStmt: true})
	seriesStmt := tx.Session(&gorm.Session{PrepareStmt: true})
	instanceStmt := tx.Session(&gorm.Session{PrepareStmt: true})
	mriStmt := tx.Session(&gorm.Session{PrepareStmt: true})

	for _, data := range batch {
		// Try to get patient from cache first
		if cachedPatient, ok := importer.patientCache.Load(data.Patient.PatientMRN); ok {
			data.Patient = cachedPatient.(*models.Patient)
		} else {
			var patient models.Patient
			err := patientStmt.Where("patient_mrn = ?", data.Patient.PatientMRN).First(&patient).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = patientStmt.Create(data.Patient).Error
				if err != nil {
					tx.Rollback()
					log.Printf("Error creating patient: %v", err)
					return
				}
				importer.patientCache.Store(data.Patient.PatientMRN, data.Patient)
			} else if err != nil {
				tx.Rollback()
				log.Printf("Error querying patient: %v", err)
				return
			} else {
				data.Patient = &patient
				importer.patientCache.Store(data.Patient.PatientMRN, &patient)
			}
		}

		// Similar pattern for Study
		studyKey := data.Study.StudyInstanceUID
		if cachedStudy, ok := importer.studyCache.Load(studyKey); ok {
			data.Study = cachedStudy.(*models.Study)
		} else {
			var study models.Study
			err := studyStmt.Where("study_instance_uid = ?", studyKey).First(&study).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				data.Study.PatientID = data.Patient.ID
				err = studyStmt.Create(data.Study).Error
				if err != nil {
					tx.Rollback()
					log.Printf("Error creating study: %v", err)
					return
				}
				importer.studyCache.Store(studyKey, data.Study)
			} else if err != nil {
				tx.Rollback()
				log.Printf("Error querying study: %v", err)
				return
			} else {
				data.Study = &study
				importer.studyCache.Store(studyKey, &study)
			}
		}

		// Similar pattern for Series
		seriesKey := data.Series.SeriesInstanceUID
		if cachedSeries, ok := importer.seriesCache.Load(seriesKey); ok {
			data.Series = cachedSeries.(*models.Series)
		} else {
			var series models.Series
			err := seriesStmt.Where("series_instance_uid = ?", seriesKey).First(&series).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				data.Series.StudyID = data.Study.ID
				err = seriesStmt.Create(data.Series).Error
				if err != nil {
					tx.Rollback()
					log.Printf("Error creating series: %v", err)
					return
				}

				// Create MRISpecifics only for new series and only if needed
				if data.MRISpecifics != nil && (data.Series.Modality == "MR" || data.Series.Modality == "MROT") {
					data.MRISpecifics.SeriesID = data.Series.ID
					err = mriStmt.Create(data.MRISpecifics).Error
					if err != nil {
						tx.Rollback()
						log.Printf("Error creating MRI specifics: %v", err)
						return
					}
				}

				importer.seriesCache.Store(seriesKey, data.Series)
			} else if err != nil {
				tx.Rollback()
				log.Printf("Error querying series: %v", err)
				return
			} else {
				data.Series = &series
				importer.seriesCache.Store(seriesKey, &series)
			}
		}

		// Check for existing instance before creating
		instanceKey := data.Instance.SOPInstanceUID
		if _, ok := importer.instanceCache.Load(instanceKey); !ok {
			var existingInstance models.Instance
			err := instanceStmt.Where("sop_instance_uid = ?", instanceKey).First(&existingInstance).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				data.Instance.SeriesID = data.Series.ID
				err = instanceStmt.Create(data.Instance).Error
				if err != nil {
					if !strings.Contains(err.Error(), "UNIQUE constraint") {
						tx.Rollback()
						log.Printf("Error creating instance: %v", err)
						return
					}
				} else {
					importer.instanceCache.Store(instanceKey, true)
				}
			} else if err != nil {
				tx.Rollback()
				log.Printf("Error querying instance: %v", err)
				return
			} else {
				importer.instanceCache.Store(instanceKey, true)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing batch: %v", err)
	} else if importer.Debug {
		log.Printf("Successfully wrote batch of %d records\n", len(batch))
	}
}

// Periodically clean up caches to prevent memory growth
func (importer *DICOMFilesImporter) cleanupCaches() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		importer.patientCache = sync.Map{}
		importer.studyCache = sync.Map{}
		importer.seriesCache = sync.Map{}
		importer.instanceCache = sync.Map{}
	}
}
