package domain

import (
	"fmt"
	"strings"
	"time"
)

// Customer is the owner of the vehicles the workshop services.
type Customer struct {
	ID             string
	FullName       string
	DocumentNumber string
	Phone          string
	Email          string
	CreatedAt      time.Time
}

// NewCustomer builds a customer after validating its contact data.
func NewCustomer(id, fullName, documentNumber, phone, email string, createdAt time.Time) (Customer, error) {
	fullName = strings.TrimSpace(fullName)
	documentNumber = strings.TrimSpace(documentNumber)
	phone = strings.TrimSpace(phone)
	email = strings.TrimSpace(strings.ToLower(email))
	if id == "" {
		return Customer{}, fmt.Errorf("%w: customer identifier is required", ErrInvalidInput)
	}
	if fullName == "" {
		return Customer{}, fmt.Errorf("%w: customer name is required", ErrInvalidInput)
	}
	if strings.ContainsAny(fullName, "<>;\"--") {
		return Customer{}, fmt.Errorf("%w: customer name contains invalid characters", ErrInvalidInput)
	}
	if documentNumber == "" {
		return Customer{}, fmt.Errorf("%w: customer document number is required", ErrInvalidInput)
	}
	for _, ch := range documentNumber {
		if !(ch >= 'A' && ch <= 'Z') && !(ch >= 'a' && ch <= 'z') && !(ch >= '0' && ch <= '9') && ch != '-' {
			return Customer{}, fmt.Errorf("%w: customer document number contains invalid characters", ErrInvalidInput)
		}
	}
	if len(documentNumber) < 4 || len(documentNumber) > 20 {
		return Customer{}, fmt.Errorf("%w: customer document number must be between 4 and 20 characters", ErrInvalidInput)
	}
	if phone == "" {
		return Customer{}, fmt.Errorf("%w: customer phone is required", ErrInvalidInput)
	}
	for _, ch := range phone {
		if !(ch >= '0' && ch <= '9') && ch != '+' && ch != '-' && ch != ' ' {
			return Customer{}, fmt.Errorf("%w: customer phone contains invalid characters", ErrInvalidInput)
		}
	}
	if len(phone) < 7 || len(phone) > 20 {
		return Customer{}, fmt.Errorf("%w: customer phone must be between 7 and 20 characters", ErrInvalidInput)
	}
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") || strings.ContainsAny(email, "<>;\"' ") {
		return Customer{}, fmt.Errorf("%w: customer email is not a valid address", ErrInvalidInput)
	}
	return Customer{
		ID:             id,
		FullName:       fullName,
		DocumentNumber: documentNumber,
		Phone:          phone,
		Email:          email,
		CreatedAt:      createdAt,
	}, nil
}
