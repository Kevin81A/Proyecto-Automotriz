package domain

import (
	"fmt"
	"strings"
	"time"
)

// MinimumModelYear is the earliest model year the workshop accepts.
const MinimumModelYear = 1901

// Vehicle is the automobile under service, owned by exactly one customer.
type Vehicle struct {
	ID         string
	CustomerID string
	Plate      string
	VIN        string
	Brand      string
	Model      string
	ModelYear  int
	CreatedAt  time.Time
}

// NewVehicle builds a vehicle after validating its identifying data. The plate
// and the VIN are stored uppercase so a lookup never depends on how the
// reception desk typed them.
func NewVehicle(id, customerID, plate, vin, brand, model string, modelYear int, createdAt time.Time) (Vehicle, error) {
	plate = strings.ToUpper(strings.TrimSpace(plate))
	vin = strings.ToUpper(strings.TrimSpace(vin))
	brand = strings.TrimSpace(brand)
	model = strings.TrimSpace(model)
	if id == "" {
		return Vehicle{}, fmt.Errorf("%w: vehicle identifier is required", ErrInvalidInput)
	}
	if customerID == "" {
		return Vehicle{}, fmt.Errorf("%w: the vehicle owner is required", ErrInvalidInput)
	}
	if plate == "" {
		return Vehicle{}, fmt.Errorf("%w: vehicle plate is required", ErrInvalidInput)
	}
	for _, ch := range plate {
		if !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') && ch != '-' {
			return Vehicle{}, fmt.Errorf("%w: vehicle plate contains invalid characters", ErrInvalidInput)
		}
	}
	if len(plate) < 3 || len(plate) > 10 {
		return Vehicle{}, fmt.Errorf("%w: vehicle plate must be between 3 and 10 characters", ErrInvalidInput)
	}
	if vin == "" {
		return Vehicle{}, fmt.Errorf("%w: vehicle VIN is required", ErrInvalidInput)
	}
	for _, ch := range vin {
		if !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') {
			return Vehicle{}, fmt.Errorf("%w: vehicle VIN contains invalid characters", ErrInvalidInput)
		}
	}
	if len(vin) < 4 || len(vin) > 17 {
		return Vehicle{}, fmt.Errorf("%w: vehicle VIN must be between 4 and 17 characters", ErrInvalidInput)
	}
	if brand == "" || model == "" {
		return Vehicle{}, fmt.Errorf("%w: vehicle brand and model are required", ErrInvalidInput)
	}
	if strings.ContainsAny(brand, "<>;\"'--") || strings.ContainsAny(model, "<>;\"'--") {
		return Vehicle{}, fmt.Errorf("%w: vehicle brand or model contains invalid characters", ErrInvalidInput)
	}
	maxYear := time.Now().Year() + 2
	if modelYear < MinimumModelYear || modelYear > maxYear {
		return Vehicle{}, fmt.Errorf("%w: vehicle model year %d is not possible", ErrInvalidInput, modelYear)
	}
	return Vehicle{
		ID:         id,
		CustomerID: customerID,
		Plate:      plate,
		VIN:        vin,
		Brand:      brand,
		Model:      model,
		ModelYear:  modelYear,
		CreatedAt:  createdAt,
	}, nil
}
