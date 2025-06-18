//go:generate go run ../../cmd/dicommapper/main.go
package processor

import (
	"errors"
	"fmt"
	"github.com/NKI-AI/aiforoncology/src/golang/research/dcmw/internal/dicomreader"
	"github.com/NKI-AI/aiforoncology/src/golang/research/dcmw/internal/models"
	"gorm.io/gorm"
	"log"
	"time"
)

type ProcessedData struct {
	Patient      *models.Patient
	Study        *models.Study
	Series       *models.Series
	Instance     *models.Instance
	MRISpecifics *models.MRISpecifics
	// Add other modality-specific fields as needed
}

func ProcessOneDICOMFile(filePath string, db *gorm.DB, debug bool) (*ProcessedData, error) {
	dataset, err := dicomreader.ReadDICOMMetadata(filePath)
	if dataset == nil {
		return nil, fmt.Errorf("DICOM dataset is nil for file %s", filePath)
	}

	if err != nil {
		if debug {
			log.Printf("Error reading DICOM file %s: %v\n", filePath, err)
		}
		return nil, err
	}

	// Process Patient
	patient, err := mapDatasetToPatient(dataset)
	if err != nil {
		return nil, fmt.Errorf("error mapping dataset to patient: %v", err)
	}

	// Process Study
	study, err := mapDatasetToStudy(dataset)
	if err != nil {
		return nil, fmt.Errorf("error mapping dataset to study: %v", err)
	}

	// Process Series
	series, err := mapDatasetToSeries(dataset)
	if err != nil {
		return nil, fmt.Errorf("error mapping dataset to series: %v", err)
	}

	// Process Instance
	instance, err := mapDatasetToInstance(dataset)
	if err != nil {
		return nil, fmt.Errorf("error mapping dataset to image: %v", err)
	}
	instance.DicomFilePath = filePath

	// Process modality-specific data
	// We need to check series.Modality. If it is MRI or MROT, we need to process MRI-specific data
	var existingMRISpecifics models.MRISpecifics
	var mriSpecifics *models.MRISpecifics

	if series.Modality == "MR" || series.Modality == "MROT" {
		err := db.Where("series_id = ?", series.ID).First(&existingMRISpecifics).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			mriSpecifics, err = mapDatasetToMRISpecifics(dataset)
			if err != nil {
				return nil, fmt.Errorf("error mapping dataset to MRI image: %v", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("error checking for existing MRI specifics: %v", err)
		}
	}

	processedData := &ProcessedData{
		Patient:      patient,
		Study:        study,
		Series:       series,
		Instance:     instance,
		MRISpecifics: mriSpecifics,
	}

	return processedData, nil
}

// Helper function to parse DICOM date and time
func parseDICOMDateTime(dateStr, timeStr *string) (*time.Time, error) {
	if dateStr == nil || timeStr == nil {
		return nil, fmt.Errorf("date or time is nil")
	}

	// Parse the date in YYYYMMDD format
	date, err := time.Parse("20060102", *dateStr) // YYYYMMDD format
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %v", err)
	}

	// Handle DICOM time format with possible trailing periods or missing components
	if len(*timeStr) > 0 && (*timeStr)[len(*timeStr)-1] == '.' {
		*timeStr = (*timeStr)[:len(*timeStr)-1] // Remove trailing period
	}

	// Handle cases where the time is only in HHMM format (length 4)
	if len(*timeStr) == 4 {
		*timeStr += "00" // Append "00" for missing seconds
	}

	// Define possible formats to handle standard and varying fractional seconds
	formats := []string{
		"150405",       // HHMMSS
		"150405.0",     // HHMMSS.s
		"150405.00",    // HHMMSS.ss
		"150405.000",   // HHMMSS.sss
		"150405.0000",  // HHMMSS.ssss
		"150405.00000", // HHMMSS.sssss
	}

	// Attempt to parse using each format until one succeeds
	var parsedTime time.Time
	for _, format := range formats {
		parsedTime, err = time.Parse(format, *timeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %v", err)
	}

	// Combine date and time
	dateTime := time.Date(date.Year(), date.Month(), date.Day(),
		parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), parsedTime.Nanosecond(), time.UTC)

	return &dateTime, nil
}
