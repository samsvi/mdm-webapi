package mdm

import (
	"net/http"
	"time"

	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type implPatientsAPI struct {
}

func NewPatientsAPI() PatientsAPI {
	return &implPatientsAPI{}
}

func (o implPatientsAPI) CreatePatient(c *gin.Context) {
	updatePatientCollectionFunc(c, func(c *gin.Context, patients *[]Patient) (*[]Patient, interface{}, int) {
		var patient Patient

		if err := c.ShouldBindJSON(&patient); err != nil {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Invalid request body",
				"error":   err.Error(),
			}, http.StatusBadRequest
		}

		// Validate required fields
		if patient.FirstName == "" || patient.LastName == "" || 
		   patient.DateOfBirth == "" || patient.Gender == "" || 
		   patient.InsuranceNumber == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Missing required fields",
			}, http.StatusBadRequest
		}

		if patient.Id == "" || patient.Id == "@new" {
			patient.Id = uuid.NewString()
		}

		// Check if patient already exists
		conflictIndx := slices.IndexFunc(*patients, func(p Patient) bool {
			return patient.Id == p.Id || patient.InsuranceNumber == p.InsuranceNumber
		})

		if conflictIndx >= 0 {
			return nil, gin.H{
				"status":  http.StatusConflict,
				"message": "Patient already exists",
			}, http.StatusConflict
		}

		// Set timestamps
		now := time.Now()
		patient.CreatedAt = now
		patient.UpdatedAt = now

		*patients = append(*patients, patient)
		
		// Return the created patient
		patientIndx := slices.IndexFunc(*patients, func(p Patient) bool {
			return patient.Id == p.Id
		})
		
		if patientIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to save patient",
			}, http.StatusInternalServerError
		}

		return patients, (*patients)[patientIndx], http.StatusCreated
	})
}

func (o implPatientsAPI) DeletePatient(c *gin.Context) {
	updatePatientCollectionFunc(c, func(c *gin.Context, patients *[]Patient) (*[]Patient, interface{}, int) {
		patientId := c.Param("patientId")

		if patientId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID is required",
			}, http.StatusBadRequest
		}

		patientIndx := slices.IndexFunc(*patients, func(p Patient) bool {
			return patientId == p.Id
		})

		if patientIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Patient not found",
			}, http.StatusNotFound
		}

		*patients = append((*patients)[:patientIndx], (*patients)[patientIndx+1:]...)
		return patients, nil, http.StatusNoContent
	})
}

func (o implPatientsAPI) GetAllPatients(c *gin.Context) {
	updatePatientCollectionFunc(c, func(c *gin.Context, patients *[]Patient) (*[]Patient, interface{}, int) {
		result := *patients
		if result == nil {
			result = []Patient{}
		}
		// return nil patients - no need to update it in db
		return nil, result, http.StatusOK
	})
}

func (o implPatientsAPI) GetPatient(c *gin.Context) {
	updatePatientCollectionFunc(c, func(c *gin.Context, patients *[]Patient) (*[]Patient, interface{}, int) {
		patientId := c.Param("patientId")

		if patientId == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Patient ID is required",
			}, http.StatusBadRequest
		}

		patientIndx := slices.IndexFunc(*patients, func(p Patient) bool {
			return patientId == p.Id
		})

		if patientIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Patient not found",
			}, http.StatusNotFound
		}

		// return nil patients - no need to update it in db
		return nil, (*patients)[patientIndx], http.StatusOK
	})
}

func (o implPatientsAPI) UpdatePatient(c *gin.Context) {
	updatePatientCollectionFunc(c, func(c *gin.Context, patients *[]Patient) (*[]Patient, interface{}, int) {
		var updatedPatient Patient

		if err := c.ShouldBindJSON(&updatedPatient); err != nil {
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

		patientIndx := slices.IndexFunc(*patients, func(p Patient) bool {
			return patientId == p.Id
		})

		if patientIndx < 0 {
			return nil, gin.H{
				"status":  http.StatusNotFound,
				"message": "Patient not found",
			}, http.StatusNotFound
		}

		// Verify that the ID in the path matches the ID in the body (if provided)
		if updatedPatient.Id != "" && updatedPatient.Id != patientId {
			return nil, gin.H{
				"status":  http.StatusForbidden,
				"message": "Patient ID in path and request body do not match",
			}, http.StatusForbidden
		}

		// Validate required fields
		if updatedPatient.FirstName == "" || updatedPatient.LastName == "" || 
		   updatedPatient.DateOfBirth == "" || updatedPatient.Gender == "" || 
		   updatedPatient.InsuranceNumber == "" {
			return nil, gin.H{
				"status":  http.StatusBadRequest,
				"message": "Missing required fields",
			}, http.StatusBadRequest
		}

		// Update fields
		existingPatient := &(*patients)[patientIndx]
		existingPatient.FirstName = updatedPatient.FirstName
		existingPatient.LastName = updatedPatient.LastName
		existingPatient.DateOfBirth = updatedPatient.DateOfBirth
		existingPatient.Gender = updatedPatient.Gender
		existingPatient.InsuranceNumber = updatedPatient.InsuranceNumber
		
		// Update optional fields if provided
		if updatedPatient.BloodType != "" {
			existingPatient.BloodType = updatedPatient.BloodType
		}
		
		if updatedPatient.Status != "" {
			existingPatient.Status = updatedPatient.Status
		}
		
		if updatedPatient.Allergies != "" {
			existingPatient.Allergies = updatedPatient.Allergies
		}
		
		if updatedPatient.MedicalNotes != "" {
			existingPatient.MedicalNotes = updatedPatient.MedicalNotes
		}

		// Update timestamp
		existingPatient.UpdatedAt = time.Now()

		return patients, (*patients)[patientIndx], http.StatusOK
	})
}