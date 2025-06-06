/*
 * Patient Management Api
 *
 * Patient and Medical Records management for Web-In-Cloud system
 *
 * API version: 1.0.0
 * Contact: your-email@example.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package mdm

type Medication struct {

	// Name of the medication
	Name string `json:"name,omitempty"`

	// Dosage amount
	Dosage string `json:"dosage,omitempty"`

	// How often to take the medication
	Frequency string `json:"frequency,omitempty"`

	// Duration of treatment
	Duration string `json:"duration,omitempty"`
}
