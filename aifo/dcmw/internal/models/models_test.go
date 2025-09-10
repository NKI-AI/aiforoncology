package models

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory SQLite database and migrates the models.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to in-memory database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&Patient{},
		&Study{},
		&Series{},
		&Instance{},
		&MRIInstance{},
		&CTInstance{},
		&MRIInstanceVendor{},
		&CTInstanceVendor{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate database schema: %v", err)
	}

	return db
}

// assertEqual is a helper function to check equality and report errors.
func assertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("Assertion Failed: %s. Expected '%v', got '%v'", message, expected, actual)
	}
}

// assertNoError is a helper function to check for no error and report if an error exists.
func assertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf("Assertion Failed: %s. Unexpected error: %v", message, err)
	}
}

// assertNotNil is a helper function to check if a value is not nil.
func assertNotNil(t *testing.T, value interface{}, message string) {
	if value == nil {
		t.Errorf("Assertion Failed: %s. Value is nil", message)
	}
}

// TestAssociations tests complex associations, including MRIInstance, CTInstance, and their vendors.
func TestAssociations(t *testing.T) {
	db := setupTestDB(t)

	// Create Patient
	patient := Patient{
		PatientID:        "P12345",
		PatientName:      "Jane Doe",
		PatientBirthDate: "19900101",
		PatientSex:       "F",
		PatientWeight:    "60",
	}
	err := db.Create(&patient).Error
	assertNoError(t, err, "Failed to create patient")

	// Create Study
	study := Study{
		PatientID:        patient.ID,
		StudyInstanceUID: "SUID12345",
		StudyID:          "ST12345",
		AccessionNumber:  "ACC12345",
	}
	err = db.Create(&study).Error
	assertNoError(t, err, "Failed to create study")

	// Create Series
	series := Series{
		StudyInstanceUID:  study.StudyInstanceUID,
		SeriesInstanceUID: "SERIESUID12345",
		SeriesNumber:      "1",
		Modality:          "MRI",
		InstanceCount:     50,
		SeriesDate:        "20240202",
		SeriesTime:        "130000",
		SliceThickness:    "2.0",
	}
	err = db.Create(&series).Error
	assertNoError(t, err, "Failed to create series")

	// Create MRIInstance
	mriInstance := MRIInstance{
		InstanceID: 1, // Assuming Instance ID will be 1
	}
	err = db.Create(&mriInstance).Error
	assertNoError(t, err, "Failed to create MRIInstance")

	// Create CTInstance
	ctInstance := CTInstance{
		InstanceID: 2, // Assuming Instance ID will be 2
	}
	err = db.Create(&ctInstance).Error
	assertNoError(t, err, "Failed to create CTInstance")

	// Create MRIInstanceVendor
	mriVendor := MRIInstanceVendor{
		VendorName:    "VendorA",
		MRIInstanceID: mriInstance.ID,
	}
	err = db.Create(&mriVendor).Error
	assertNoError(t, err, "Failed to create MRIInstanceVendor")

	// Create CTInstanceVendor
	ctVendor := CTInstanceVendor{
		VendorName:   "VendorB",
		CTInstanceID: ctInstance.ID,
	}
	err = db.Create(&ctVendor).Error
	assertNoError(t, err, "Failed to create CTInstanceVendor")

	// Verify Associations
	var retrievedPatient Patient
	err = db.Preload("Studies.Series.Images.MRIInstance.Vendors").
		First(&retrievedPatient, "patient_id = ?", "P12345").Error
	assertNoError(t, err, "Failed to retrieve patient with associations")
	if len(retrievedPatient.Studies) != 1 {
		t.Errorf("Expected 1 study, got %d", len(retrievedPatient.Studies))
	}
	if len(retrievedPatient.Studies[0].Series) != 1 {
		t.Errorf("Expected 1 series, got %d", len(retrievedPatient.Studies[0].Series))
	}
	// Depending on your model setup, you might need to create instances associated with series
	// For this test, ensure that images are correctly associated if any
}

// TestTimestamps tests the handling of datetime fields by combining separate date and time strings into a time.Time object.
func TestTimestamps(t *testing.T) {
	db := setupTestDB(t)

	// Create Patient
	patient := Patient{
		PatientID:        "P67890",
		PatientName:      "Alice Smith",
		PatientBirthDate: "19750505",
		PatientSex:       "F",
		PatientWeight:    "65",
	}
	err := db.Create(&patient).Error
	assertNoError(t, err, "Failed to create patient")

	// Create Study with StudyDateTime
	studyDate, err := time.Parse("20060102", "20230315")
	if err != nil {
		t.Fatalf("Failed to parse StudyDate: %v", err)
	}
	_, err = time.Parse("150405", "093000")
	if err != nil {
		t.Fatalf("Failed to parse StudyTime: %v", err)
	}
	studyDateTime := studyDate.Add(time.Hour*9 + time.Minute*30) // 09:30

	study := Study{
		PatientID:        patient.ID,
		StudyInstanceUID: "SUID67890",
		StudyID:          "ST67890",
		AccessionNumber:  "ACC67890",
		StudyDate:        "20230315",
		StudyTime:        "093000",
		StudyDateTime:    &studyDateTime,
	}

	err = db.Create(&study).Error
	assertNoError(t, err, "Failed to create study with datetime")

	// Retrieve and verify StudyDateTime
	var retrievedStudy Study
	err = db.First(&retrievedStudy, "study_instance_uid = ?", "SUID67890").Error
	assertNoError(t, err, "Failed to retrieve study")
	if retrievedStudy.StudyDateTime == nil {
		t.Errorf("StudyDateTime should not be nil")
	} else {
		assertEqual(t, studyDateTime.Unix(), retrievedStudy.StudyDateTime.Unix(), "StudyDateTime mismatch")
	}
}
