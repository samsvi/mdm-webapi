package mdm

import (
	"net/http"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implMedicalRecordsAPI struct {
}

func NewMedicalRecordsAPI() MedicalRecordsAPI {
	return &implMedicalRecordsAPI{}
}

func (o implMedicalRecordsAPI) CreateMedicalRecord(c *gin.Context) {
	updateMedicalRecordsCollectionFunc(c, func(c *gin.Context, medicalRecords *[]MedicalRecord) (*[]MedicalRecord, interface{}, int) {
		var record MedicalRecord

		if err := c.ShouldBindJSON(&record); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		patientId := c.Param("patientId")
		if patientId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID is required",
			}, http.StatusBadRequest
		}

		// Validate required fields
		if record.Diagnosis == "" || record.DateOfVisit.IsZero() {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Missing required fields (diagnosis, dateOfVisit)",
			}, http.StatusBadRequest
		}

		if record.Id == "" || record.Id == "@new" {
			record.Id = uuid.NewString()
		}

		// Set patient ID from URL parameter
		record.PatientId = patientId

		// Check if medical record already exists
		conflictIndx := slices.IndexFunc(*medicalRecords, func(mr MedicalRecord) bool {
			return record.Id == mr.Id
		})

		if conflictIndx >= 0 {
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Medical record already exists",
			}, http.StatusConflict
		}

		// Set timestamps
		now := time.Now()
		record.CreatedAt = now
		record.UpdatedAt = now

		*medicalRecords = append(*medicalRecords, record)
		
		// Return the created record
		recordIndx := slices.IndexFunc(*medicalRecords, func(mr MedicalRecord) bool {
			return record.Id == mr.Id
		})
		
		if recordIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to save medical record",
			}, http.StatusInternalServerError
		}

		return medicalRecords, (*medicalRecords)[recordIndx], http.StatusCreated
	})
}

func (o implMedicalRecordsAPI) DeleteMedicalRecord(c *gin.Context) {
	updateMedicalRecordsCollectionFunc(c, func(c *gin.Context, medicalRecords *[]MedicalRecord) (*[]MedicalRecord, interface{}, int) {
		patientId := c.Param("patientId")
		recordId := c.Param("recordId")

		if patientId == "" || recordId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID and Record ID are required",
			}, http.StatusBadRequest
		}

		recordIndx := slices.IndexFunc(*medicalRecords, func(mr MedicalRecord) bool {
			return recordId == mr.Id && patientId == mr.PatientId
		})

		if recordIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Patient or Medical record not found",
			}, http.StatusNotFound
		}

		*medicalRecords = append((*medicalRecords)[:recordIndx], (*medicalRecords)[recordIndx+1:]...)
		return medicalRecords, nil, http.StatusNoContent
	})
}

func (o implMedicalRecordsAPI) GetPatientMedicalRecords(c *gin.Context) {
	updateMedicalRecordsCollectionFunc(c, func(c *gin.Context, medicalRecords *[]MedicalRecord) (*[]MedicalRecord, interface{}, int) {
		patientId := c.Param("patientId")

		if patientId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID is required",
			}, http.StatusBadRequest
		}

		// Filter records for specific patient
		var patientRecords []MedicalRecord
		for _, record := range *medicalRecords {
			if record.PatientId == patientId {
				patientRecords = append(patientRecords, record)
			}
		}

		if patientRecords == nil {
			patientRecords = []MedicalRecord{}
		}

		// return nil medicalRecords - no need to update it in db
		return nil, patientRecords, http.StatusOK
	})
}

func (o implMedicalRecordsAPI) UpdateMedicalRecord(c *gin.Context) {
	updateMedicalRecordsCollectionFunc(c, func(c *gin.Context, medicalRecords *[]MedicalRecord) (*[]MedicalRecord, interface{}, int) {
		var updatedRecord MedicalRecord

		if err := c.ShouldBindJSON(&updatedRecord); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		patientId := c.Param("patientId")
		recordId := c.Param("recordId")

		if patientId == "" || recordId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID and Record ID are required",
			}, http.StatusBadRequest
		}

		recordIndx := slices.IndexFunc(*medicalRecords, func(mr MedicalRecord) bool {
			return recordId == mr.Id && patientId == mr.PatientId
		})

		if recordIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Patient or Medical record not found",
			}, http.StatusNotFound
		}

		// Verify that the ID in the path matches the ID in the body (if provided)
		if updatedRecord.Id != "" && updatedRecord.Id != recordId {
			return nil, gin.H{
				"status":  http.StatusForbidden,
				"message": "Record ID in path and request body do not match",
			}, http.StatusForbidden
		}

		// Validate required fields
		if updatedRecord.Diagnosis == "" || updatedRecord.DateOfVisit.IsZero() {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Missing required fields (diagnosis, dateOfVisit)",
			}, http.StatusBadRequest
		}

		// Update fields
		existingRecord := &(*medicalRecords)[recordIndx]
		existingRecord.Diagnosis = updatedRecord.Diagnosis
		existingRecord.DateOfVisit = updatedRecord.DateOfVisit
		
		// Update optional fields if provided
		if updatedRecord.Symptoms != nil {
			existingRecord.Symptoms = updatedRecord.Symptoms
		}
		
		if updatedRecord.Treatment != "" {
			existingRecord.Treatment = updatedRecord.Treatment
		}
		
		if updatedRecord.Medications != nil {
			existingRecord.Medications = updatedRecord.Medications
		}
		
		if updatedRecord.DoctorName != "" {
			existingRecord.DoctorName = updatedRecord.DoctorName
		}
		
		if updatedRecord.Notes != "" {
			existingRecord.Notes = updatedRecord.Notes
		}
		
		if updatedRecord.FollowUpDate != "" {
			existingRecord.FollowUpDate = updatedRecord.FollowUpDate
		}

		// Update timestamp
		existingRecord.UpdatedAt = time.Now()

		return medicalRecords, (*medicalRecords)[recordIndx], http.StatusOK
	})
}