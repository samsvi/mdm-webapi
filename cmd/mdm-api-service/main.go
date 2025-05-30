package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samsvi/mdm-webapi/api"
	"github.com/samsvi/mdm-webapi/internal/db_service"
	"github.com/samsvi/mdm-webapi/internal/mdm"
)

func main() {
    log.Printf("Server started")
    port := os.Getenv("MDM_API_PORT")
    if port == "" {
        port = "8080"
    }
    environment := os.Getenv("MDM_API_ENVIRONMENT")
    if !strings.EqualFold(environment, "production") { // case insensitive comparison
        gin.SetMode(gin.DebugMode)
    }
    
    engine := gin.New()
    engine.Use(gin.Recovery())
    
    corsMiddleware := cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        ExposeHeaders:    []string{""},
        AllowCredentials: false,
        MaxAge: 12 * time.Hour,
    })
    engine.Use(corsMiddleware)

    // Setup database services for individual documents
    patientsDbService := db_service.NewMongoService[mdm.Patient](db_service.MongoServiceConfig{
        Collection: "patients",
    })
    defer patientsDbService.Disconnect(context.Background())

    medicalRecordsDbService := db_service.NewMongoService[mdm.MedicalRecord](db_service.MongoServiceConfig{
        Collection: "medical-records", 
    })
    defer medicalRecordsDbService.Disconnect(context.Background())

    // Setup context middleware to set appropriate db_service
    engine.Use(func(ctx *gin.Context) {
        path := ctx.Request.URL.Path
        if strings.Contains(path, "/medical-records") {
            ctx.Set("db_service", medicalRecordsDbService)
        } else if strings.Contains(path, "/patients") {
            ctx.Set("db_service", patientsDbService)
        }
        ctx.Next()
    })

    // Create API implementations
    patientsAPI := mdm.NewPatientsAPI()
    medicalRecordsAPI := mdm.NewMedicalRecordsAPI()

    // Request routings
    engine.GET("/openapi", api.HandleOpenApi)
    
    // Patients routes
    engine.GET("/api/patients", patientsAPI.GetAllPatients)
    engine.POST("/api/patients", patientsAPI.CreatePatient)
    engine.GET("/api/patients/:patientId", patientsAPI.GetPatient)
    engine.PUT("/api/patients/:patientId", patientsAPI.UpdatePatient)
    engine.DELETE("/api/patients/:patientId", patientsAPI.DeletePatient)
    
    // Medical records routes
    engine.GET("/api/patients/:patientId/medical-records", medicalRecordsAPI.GetPatientMedicalRecords)
    engine.POST("/api/patients/:patientId/medical-records", medicalRecordsAPI.CreateMedicalRecord)
    engine.PUT("/api/patients/:patientId/medical-records/:recordId", medicalRecordsAPI.UpdateMedicalRecord)
    engine.DELETE("/api/patients/:patientId/medical-records/:recordId", medicalRecordsAPI.DeleteMedicalRecord)

    engine.Run(":" + port)
}