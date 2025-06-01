package mdm

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samsvi/mdm-webapi/internal/db_service"
)

type implPatientsAPI struct {
}

func NewPatientsAPI() PatientsAPI {
	return &implPatientsAPI{}
}

func (o implPatientsAPI) CreatePatient(c *gin.Context) {
	log.Printf("🚀 CreatePatient handler called!")
	
	var patient Patient

	if err := c.ShouldBindJSON(&patient); err != nil {
		log.Printf("❌ JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",  
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("✅ Parsed patient: %+v", patient)

	// Validate required fields
	if patient.FirstName == "" || patient.LastName == "" || 
	   patient.DateOfBirth == "" || patient.Gender == "" || 
	   patient.InsuranceNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Missing required fields",
		})
		return
	}

	if patient.Id == "" || patient.Id == "@new" {
		patient.Id = uuid.NewString()
	}

	// Set timestamps
	now := time.Now()
	patient.CreatedAt = now
	patient.UpdatedAt = now

	log.Printf("💾 Saving patient with ID: %s", patient.Id)

	// Get db service
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[Patient])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error", 
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Create patient document
	if err := db.CreateDocument(c, patient.Id, &patient); err != nil {
		log.Printf("❌ Database error: %v", err)
		switch err {
		case db_service.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{
				"status":  "Conflict",
				"message": "Patient already exists",
			})
		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to create patient",
				"error":   err.Error(),
			})
		}
		return
	}

	log.Printf("✅ Patient created successfully!")
	c.JSON(http.StatusCreated, patient)
}

func (o implPatientsAPI) GetAllPatients(c *gin.Context) {
	log.Printf("🔍 GetAllPatients called")
	
	// Get db service
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found",
		})
		return
	}

	db, ok := value.(db_service.DbService[Patient])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Get all patients using db service
	patients, err := db.FindAllDocuments(c)
	if err != nil {
		log.Printf("❌ Error finding patients: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to retrieve patients",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("✅ Found %d patients", len(patients))
	c.JSON(http.StatusOK, patients)
}

func (o implPatientsAPI) GetPatient(c *gin.Context) {
	patientId := c.Param("patientId")
	log.Printf("🔍 GetPatient called for ID: %s", patientId)
	
	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",
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

	db, ok := value.(db_service.DbService[Patient])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Find patient
	patient, err := db.FindDocument(c, patientId)
	switch err {
	case nil:
		log.Printf("✅ Found patient: %s", patient.Id)
		c.JSON(http.StatusOK, *patient)
	case db_service.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "Not Found", 
			"message": "Patient not found",
		})
	default:
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to find patient",
			"error":   err.Error(),
		})
	}
}

func (o implPatientsAPI) UpdatePatient(c *gin.Context) {
	patientId := c.Param("patientId")
	log.Printf("🔄 UpdatePatient called for ID: %s", patientId)
	
	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",
		})
		return
	}

	var updatedPatient Patient
	if err := c.ShouldBindJSON(&updatedPatient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	// Verify ID match
	if updatedPatient.Id != "" && updatedPatient.Id != patientId {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "Forbidden",
			"message": "Patient ID in path and request body do not match",
		})
		return
	}

	updatedPatient.Id = patientId
	updatedPatient.UpdatedAt = time.Now()

	// Get db service
	value, exists := c.Get("db_service")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service not found", 
		})
		return
	}

	db, ok := value.(db_service.DbService[Patient])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Update patient
	if err := db.UpdateDocument(c, patientId, &updatedPatient); err != nil {
		switch err {
		case db_service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "Not Found",
				"message": "Patient not found",
			})
		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "Bad Gateway", 
				"message": "Failed to update patient",
				"error":   err.Error(),
			})
		}
		return
	}

	log.Printf("✅ Patient updated successfully!")
	c.JSON(http.StatusOK, updatedPatient)
}

func (o implPatientsAPI) DeletePatient(c *gin.Context) {
	patientId := c.Param("patientId")
	log.Printf("🗑️ DeletePatient called for ID: %s", patientId)
	
	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",  
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

	db, ok := value.(db_service.DbService[Patient])
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "db_service context is not of correct type",
		})
		return
	}

	// Delete patient
	if err := db.DeleteDocument(c, patientId); err != nil {
		switch err {
		case db_service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "Not Found",
				"message": "Patient not found",
			})
		default:
			c.JSON(http.StatusBadGateway, gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to delete patient", 
				"error":   err.Error(),
			})
		}
		return
	}

	log.Printf("✅ Patient deleted successfully!")
	c.Status(http.StatusNoContent)
}