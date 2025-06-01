package mdm

import (
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
	var patient Patient

	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",  
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

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

	now := time.Now()
	patient.CreatedAt = now
	patient.UpdatedAt = now

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

	if err := db.CreateDocument(c, patient.Id, &patient); err != nil {
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

	c.JSON(http.StatusCreated, patient)
}

func (o implPatientsAPI) GetAllPatients(c *gin.Context) {
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

	patients, err := db.FindAllDocuments(c)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "Bad Gateway",
			"message": "Failed to retrieve patients",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, patients)
}

func (o implPatientsAPI) GetPatient(c *gin.Context) {
	patientId := c.Param("patientId")
	
	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",
		})
		return
	}

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

	patient, err := db.FindDocument(c, patientId)
	switch err {
	case nil:
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

	if updatedPatient.Id != "" && updatedPatient.Id != patientId {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "Forbidden",
			"message": "Patient ID in path and request body do not match",
		})
		return
	}

	updatedPatient.Id = patientId
	updatedPatient.UpdatedAt = time.Now()

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

	c.JSON(http.StatusOK, updatedPatient)
}

func (o implPatientsAPI) DeletePatient(c *gin.Context) {
	patientId := c.Param("patientId")
	
	if patientId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "Patient ID is required",  
		})
		return
	}

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

	c.Status(http.StatusNoContent)
}