package mdm

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samsvi/mdm-webapi/internal/db_service"
)

type implMedicalRecordsAPI struct {
}

func NewMedicalRecordsAPI() MedicalRecordsAPI {
	return &implMedicalRecordsAPI{}
}

func (o implMedicalRecordsAPI) CreateMedicalRecord(c *gin.Context) {
	log.Printf("🏥 CreateMedicalRecord called")
	
	patientId := c.Param("patientId")
	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",
		})
		return
	}

	var record MedicalRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Validate required fields
	if record.Diagnosis == "" || record.DateOfVisit.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Missing required fields (diagnosis, dateOfVisit)",
		})
		return
	}

	if record.Id == "" || record.Id == "@new" {
		record.Id = uuid.NewString()
	}

	// Set patient ID from URL parameter
	record.PatientId = patientId

	// Set timestamps
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	log.Printf("💾 Saving medical record with ID: %s for patient: %s", record.Id, patientId)

	// Get db service
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[MedicalRecord])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Create medical record document
	if err := db.CreateDocument(c, record.Id, &record); err != nil {
		log.Printf("❌ Database error: %v", err)
		switch err {
		case db_service.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{
				"status":  "Conflict",
				"message": "Medical record already exists",
			})
		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create medical record",
				"error":   err.Error(),
			})
		}
		return
	}

	log.Printf("✅ Medical record created successfully!")
	c.JSON(http.StatusCreated, record)
}

func (o implMedicalRecordsAPI) GetPatientMedicalRecords(c *gin.Context) {
	patientId := c.Param("patientId")
	log.Printf("🔍 GetPatientMedicalRecords called for patient: %s", patientId)

	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",
		})
		return
	}

	// Pre jednoduchosť vraciame prázdne pole
	// V skutočnosti by sme potrebovali scan kolekcie a filtrovanie podľa patientId
	records := []MedicalRecord{}
	c.JSON(http.StatusOK, records)
}

func (o implMedicalRecordsAPI) UpdateMedicalRecord(c *gin.Context) {
	patientId := c.Param("patientId")
	recordId := c.Param("recordId")
	log.Printf("🔄 UpdateMedicalRecord called for record: %s, patient: %s", recordId, patientId)

	if patientId == "" || recordId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID and Record ID are required",
		})
		return
	}

	var updatedRecord MedicalRecord
	if err := c.ShouldBindJSON(&updatedRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Verify IDs match
	if updatedRecord.Id != "" && updatedRecord.Id != recordId {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "Forbidden",
			"message": "Record ID in path and request body do not match",
		})
		return
	}

	updatedRecord.Id = recordId
	updatedRecord.PatientId = patientId
	updatedRecord.UpdatedAt = time.Now()

	// Get db service
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[MedicalRecord])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Update medical record
	if err := db.UpdateDocument(c, recordId, &updatedRecord); err != nil {
		switch err {
		case db_service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "Not Found",
				"message": "Patient or Medical record not found",
			})
		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update medical record",
				"error":   err.Error(),
			})
		}
		return
	}

	log.Printf("✅ Medical record updated successfully!")
	c.JSON(http.StatusOK, updatedRecord)
}

func (o implMedicalRecordsAPI) DeleteMedicalRecord(c *gin.Context) {
	patientId := c.Param("patientId")
	recordId := c.Param("recordId")
	log.Printf("🗑️ DeleteMedicalRecord called for record: %s, patient: %s", recordId, patientId)

	if patientId == "" || recordId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID and Record ID are required",
		})
		return
	}

	// Get db service
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[MedicalRecord])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Delete medical record
	if err := db.DeleteDocument(c, recordId); err != nil {
		switch err {
		case db_service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "Not Found",
				"message": "Patient or Medical record not found",
			})
		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete medical record",
				"error":   err.Error(),
			})
		}
		return
	}

	log.Printf("✅ Medical record deleted successfully!")
	c.Status(http.StatusNoContent)
}