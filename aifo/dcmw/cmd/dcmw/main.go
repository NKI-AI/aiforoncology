package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NKI-AI/aiforoncology/src/golang/research/dcmw/internal/importer"
	"github.com/NKI-AI/aiforoncology/src/golang/research/dcmw/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	var dirs string
	var dbName string
	var threads int
	var batchSize int
	var debug bool

	flag.StringVar(&dirs, "dirs", "", "Comma-separated list of directories containing DICOM files.")
	flag.StringVar(&dbName, "db_name", "", "Name of the database to use.")
	flag.IntVar(&threads, "threads", 4, "Number of threads to use for processing.")
	flag.IntVar(&batchSize, "batch_size", 100, "Batch size for database commits.")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode.")

	flag.Parse()

	if dbName == "" {
		log.Fatal("No database name was set.")
	}

	if dirs == "" {
		log.Fatal("No directories were specified.")
	}

	dirList := strings.Split(dirs, ",")

	// Initialize the database
	dbConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent, Error, Warn, Info (default)
	}
	db, err := gorm.Open(sqlite.Open(dbName), dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Enable connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetMaxOpenConns(5)            // Max open connections, can be tuned based on performance needs
	sqlDB.SetMaxIdleConns(5)            // Max idle connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Max connection lifetime

	// Migrate the schema
	err = db.AutoMigrate(
		&models.Patient{},
		&models.Study{},
		&models.Series{},
		&models.Instance{},
		&models.MRISpecifics{},
		// Add other models as needed...
	)

	// Configure SQLite for better performance
	db.Exec("PRAGMA journal_mode = WAL;")
	db.Exec("PRAGMA synchronous = NORMAL")
	db.Exec("PRAGMA cache_size = -2000") // 2GB cache
	db.Exec("PRAGMA temp_store = MEMORY")
	db.Exec("PRAGMA mmap_size = 30000000000") // 30GB memory map
	db.Exec("PRAGMA page_size = 4096")

	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	for _, dir := range dirList {
		log.Printf("Starting to process directory: %s\n", dir)
		importer := importer.NewDICOMFilesImporter(dir, db, threads, batchSize, debug, -1)
		importer.ProcessFolder()
	}

	fmt.Printf("Done processing all DICOM files into database %s.\n", dbName)
}
