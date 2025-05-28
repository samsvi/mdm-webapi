package mdm

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samsvi/mdm-webapi/internal/db_service"
)

type patientCollectionUpdater = func(
	ctx *gin.Context,
	patients *[]Patient,
) (updatedPatients *[]Patient, responseContent interface{}, status int)

type medicalRecordsCollectionUpdater = func(
	ctx *gin.Context,
	medicalRecords *[]MedicalRecord,
) (updatedMedicalRecords *[]MedicalRecord, responseContent interface{}, status int)

func updatePatientCollectionFunc(ctx *gin.Context, updater patientCollectionUpdater) {
	value, exists := ctx.Get("db_service")
	if !exists {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service not found",
				"error":   "db_service not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[[]Patient])
	if !ok {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	// For patients collection, we use a fixed collection ID
	collectionId := "patients"

	patients, err := db.FindDocument(ctx, collectionId)

	switch err {
	case nil:
		// continue
	case db_service.ErrNotFound:
		// If collection doesn't exist, create empty collection
		emptyPatients := []Patient{}
		patients = &emptyPatients
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load patients from database",
				"error":   err.Error(),
			})
		return
	}

	updatedPatients, responseObject, status := updater(ctx, patients)

	if updatedPatients != nil {
		err = db.UpdateDocument(ctx, collectionId, updatedPatients)
	} else {
		err = nil // redundant but for clarity
	}

	switch err {
	case nil:
		if responseObject != nil {
			ctx.JSON(status, responseObject)
		} else {
			ctx.AbortWithStatus(status)
		}
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Patients collection was deleted while processing the request",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update patients in database",
				"error":   err.Error(),
			})
	}
}

func updateMedicalRecordsCollectionFunc(ctx *gin.Context, updater medicalRecordsCollectionUpdater) {
	value, exists := ctx.Get("db_service")
	if !exists {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service not found",
				"error":   "db_service not found",
			})
		return
	}

	db, ok := value.(db_service.DbService[[]MedicalRecord])
	if !ok {
		ctx.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status":  "Internal Server Error",
				"message": "db_service context is not of type db_service.DbService",
				"error":   "cannot cast db_service context to db_service.DbService",
			})
		return
	}

	// For medical records collection, we use a fixed collection ID
	collectionId := "medical-records"

	medicalRecords, err := db.FindDocument(ctx, collectionId)

	switch err {
	case nil:
		// continue
	case db_service.ErrNotFound:
		// If collection doesn't exist, create empty collection
		emptyRecords := []MedicalRecord{}
		medicalRecords = &emptyRecords
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to load medical records from database",
				"error":   err.Error(),
			})
		return
	}

	updatedMedicalRecords, responseObject, status := updater(ctx, medicalRecords)

	if updatedMedicalRecords != nil {
		err = db.UpdateDocument(ctx, collectionId, updatedMedicalRecords)
	} else {
		err = nil // redundant but for clarity
	}

	switch err {
	case nil:
		if responseObject != nil {
			ctx.JSON(status, responseObject)
		} else {
			ctx.AbortWithStatus(status)
		}
	case db_service.ErrNotFound:
		ctx.JSON(
			http.StatusNotFound,
			gin.H{
				"status":  "Not Found",
				"message": "Medical records collection was deleted while processing the request",
				"error":   err.Error(),
			},
		)
	default:
		ctx.JSON(
			http.StatusBadGateway,
			gin.H{
				"status":  "Bad Gateway",
				"message": "Failed to update medical records in database",
				"error":   err.Error(),
			})
	}
}