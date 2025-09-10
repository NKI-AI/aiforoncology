package models

import (
	"time"

	"gorm.io/gorm"
)

// TODO: Can we have required fields?
// Patient represents a patient in the database.
type Patient struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	PatientMRN       string `gorm:"uniqueIndex:idx_patient_mrn" dicom:"0010,0020"` // In DICOM this would be PatientID, but this clashes with GORM
	PatientName      string `gorm:"index" dicom:"0010,0010"`
	PatientBirthDate string `gorm:"index" dicom:"0010,0030"` // Maybe we should make this a datetime?

	Studies []Study `gorm:"foreignKey:PatientID;references:ID"`
}

// Study represents a study associated with a patient.
type Study struct {
	gorm.Model
	ID               uint       `gorm:"primaryKey"`
	PatientID        uint       `gorm:"index:idx_patient_study"` // Foreign key to Patient.ID (uint)
	PatientSex       string     `gorm:"index" dicom:"0010,0040"` // This can be variable across studies
	PatientWeight    string     `gorm:"index" dicom:"0010,1030"` // This can be variable across studies
	StudyInstanceUID string     `gorm:"uniqueIndex:idx_study_uid" dicom:"0020,000D"`
	StudyDescription string     `dicom:"0008,1030"`
	StudyID          string     `dicom:"0020,0010"`
	StudyDate        string     `gorm:"-" dicom:"0008,0020"`
	StudyTime        string     `gorm:"-" dicom:"0008,0030"`
	StudyDateTime    *time.Time `datetime:"StudyDate-StudyTime"`
	AccessionNumber  string     `dicom:"0008,0050"`

	Patient *Patient `gorm:"foreignKey:PatientID;references:ID"`
	Series  []Series `gorm:"foreignKey:StudyInstanceUID;references:StudyInstanceUID"`
}

// Series represents a series within a study.
type Series struct {
	gorm.Model
	ID                    uint       `gorm:"primaryKey"`
	StudyID               uint       `gorm:"index:idx_study_series"` // Foreign key to Study.ID (uint)
	SeriesInstanceUID     string     `gorm:"uniqueIndex:idx_series_uid" dicom:"0020,000E"`
	SeriesNumber          string     `dicom:"0020,0011"`
	SeriesDescription     string     `dicom:"0008,103E"`
	Modality              string     `gorm:"index" dicom:"0008,0060"`
	ProtocolName          string     `dicom:"0018,1030"`
	BodyPartExamined      string     `dicom:"0018,0015"`
	ContrastBolusAgent    string     `dicom:"0018,0010"`
	FrameOfReferenceUID   string     `dicom:"0020,0052"`
	PlanarConfiguration   string     `dicom:"0028,0006"`
	PatientPosition       string     `dicom:"0018,5100"`
	InstitutionName       string     `dicom:"0008,0080"`
	InstanceCount         int        `dicom:"0020,1209"`
	AcquisitionDuration   string     `dicom:"0018,9073"`
	Manufacturer          string     `dicom:"0008,0070"`
	ManufacturerModelName string     `dicom:"0008,1090"`
	SeriesDate            string     `gorm:"-" dicom:"0008,0021"`       // Store as string in format YYYYMMDD
	SeriesTime            string     `gorm:"-" dicom:"0008,0031"`       // Store as string in format HHMMSS
	SeriesDateTime        *time.Time `datetime:"SeriesDate-SeriesTime"` // Combine SeriesDate and SeriesTime here
	SliceThickness        string     `dicom:"0018,0050"`

	StudyInstanceUID string        `gorm:"index" dicom:"0020,000D"`
	Study            *Study        `gorm:"foreignKey:StudyInstanceUID;references:StudyInstanceUID"`
	Images           []Instance    `gorm:"foreignKey:SeriesInstanceUID;references:SeriesInstanceUID"` // TODO Rename to Instances
	MRISpecifics     *MRISpecifics `gorm:"foreignKey:SeriesID;references:ID"`
}

// Instance represents an instance (e.g., image) within a series.
type Instance struct {
	gorm.Model
	ID                       uint       `gorm:"primaryKey"`
	Columns                  int        `dicom:"0028,0011"`
	ContentDate              string     `gorm:"-" dicom:"0008,0023"`
	ContentTime              string     `gorm:"-" dicom:"0008,0033"`
	ContentDateTime          *time.Time `datetime:"ContentDate-ContentTime"`
	SeriesID                 uint       `gorm:"index:idx_series_instance"` // Foreign key to Series.ID (uint)
	SOPInstanceUID           string     `gorm:"uniqueIndex:idx_instance_uid" dicom:"0008,0018"`
	SOPClassUID              string     `dicom:"0008,0016"`
	AcquisitionNumber        int        `dicom:"0020,0012"`
	InstanceNumber           int        `dicom:"0020,0013"`
	Rows                     int        `dicom:"0028,0010"`
	SliceLocation            string     `dicom:"0020,1041"`
	DicomFilePath            string
	AcquistionDate           string     `gorm:"-" dicom:"0008,0022"`
	AcquisitionTime          string     `gorm:"-" dicom:"0008,0032"`
	AcquisitionDateTime      *time.Time `datetime:"AcquistionDate-AcquisitionTime"`
	InstanceCreationDate     string     `gorm:"-" dicom:"0008,0012"`
	InstanceCreationTime     string     `gorm:"-" dicom:"0008,0013"`
	InstanceCreationDateTime *time.Time `datetime:"InstanceCreationDate-InstanceCreationTime"`
	Modality                 string     `gorm:"index" dicom:"0008,0060"`

	SeriesInstanceUID string  `gorm:"index" dicom:"0020,000E"`
	Series            *Series `gorm:"foreignKey:SeriesInstanceUID;references:SeriesInstanceUID"`
}

// MRISpecifics represents MRI-specific image data.
type MRISpecifics struct {
	gorm.Model
	ID                        uint    `gorm:"primaryKey"`
	SeriesID                  uint    `gorm:"uniqueIndex:idx_mri_series"` // Foreign key to Series.ID (uint)
	ReceiveCoilName           string  `dicom:"0018,1250"`
	NumberOfFrames            int     `dicom:"0028,0008"`
	SamplesPerPixel           int     `dicom:"0028,0002"`
	PhotometricInterpretation string  `dicom:"0028,0004"`
	BitsStored                int     `dicom:"0028,0101"`
	ImageType                 string  `dicom:"0008,0008"`
	AcquisitionMatrix         string  `dicom:"0018,1310"`
	AcquisitionDuration       string  `dicom:"0018,9073"`
	AngioFlag                 string  `dicom:"0018,0025"`
	BeatRejectionFlag         string  `dicom:"0018,1080"`
	BitsAllocated             int     `dicom:"0028,0100"`
	DBDt                      string  `dicom:"0018,1318"`
	EchoNumber                int     `dicom:"0018,0086"`
	EchoPlanerPulseSequence   string  `dicom:"0018,9018"`
	EchoTime                  float64 `dicom:"0018,0081"`
	EchoTrainLength           int     `dicom:"0018,0091"`
	FlipAngle                 string  `dicom:"0018,1314"`
	HighBit                   string  `dicom:"0028,0102"`
	HighRRValue               string  `dicom:"0018,1082"`
	ImagedNucleus             string  `dicom:"0018,0085"`
	// ImageOrientation              string  `dicom:"0020,0037"` // Instance fields?
	// ImagePosition                 string  `dicom:"0020,0032"` // Instance fields?
	ImagesInAcquisition           int     `dicom:"0020,1002"`
	ImagingFrequency              string  `dicom:"0018,0084"`
	InPlanePhaseEncodingDirection string  `dicom:"0018,1312"`
	IntervalsAcquired             string  `dicom:"0018,1083"`
	IntervalsRejected             string  `dicom:"0018,1084"`
	InversionTime                 float64 `dicom:"0018,9079"`
	LowRRValue                    string  `dicom:"0018,1081"`
	MagneticFieldStrength         string  `dicom:"0018,0087"`
	MultiPlanarExcitation         string  `dicom:"0018,9012"`
	MultipleSpinEcho              string  `dicom:"0018,9011"`
	NominalInterval               string  `dicom:"0018,1062"`
	NumberOfAverages              float64 `dicom:"0018,0083"`
	NumberOfPhaseEncodingSteps    int     `dicom:"0018,0089"`
	NumberOfTemporalPositions     int     `dicom:"0020,0105"`
	PercentPhaseFieldOfView       string  `dicom:"0018,0094"`
	PercentSampling               string  `dicom:"0018,0093"`
	PixelAspectRatio              string  `dicom:"0028,0034"`
	PixelBandwidth                string  `dicom:"0018,0095"`
	PixelSpacing                  string  `dicom:"0028,0030"`
	PlanarConfiguration           string  `dicom:"0028,0006"`
	PulseSequenceName             string  `dicom:"0018,9005"`
	ReconstructionDiameter        string  `dicom:"0018,1100"`
	RepetitionTime                float64 `dicom:"0018,0080"`
	SaturationRecovery            string  `dicom:"0018,9024"`
	ScanningSequence              string  `dicom:"0018,0020"`
	ScanOptions                   string  `dicom:"0018,0022"`
	SequenceName                  string  `dicom:"0018,0024"`
	SequenceVariant               string  `dicom:"0018,0021"`
	SliceThickness                float64 `dicom:"0018,0050"`
	SpacingBetweenSlices          float64 `dicom:"0018,0088"`
	SteadyStatePulseSequence      string  `dicom:"0018,9017"`
	TemporalPositionIdentifier    string  `dicom:"0020,0100"`
	TemporalResolution            float64 `dicom:"0020,0110"`
	TransmitCoilName              string  `dicom:"0018,1251"`
	// TriggerTime                   float64 `dicom:"0018,1060"`
	VariableFlipAngleFlag string  `dicom:"0018,1315"`
	PulseSequence         int     `dicom:"0027,1032"`
	AcquisitionContrast   string  `dicom:"0008,9209"`
	WaterFatShift         float64 `dicom:"2001,1022"`

	Series *Series `gorm:"foreignKey:SeriesID;references:ID"`
}
