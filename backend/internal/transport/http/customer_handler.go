package http

import (
	"net/http"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

// createCustomerRequest is the payload of the customer form.
type createCustomerRequest struct {
	FullName       string `json:"fullName"`
	DocumentNumber string `json:"documentNumber"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
}

// customerResponse is one row of the customer table.
type customerResponse struct {
	ID             string `json:"id"`
	FullName       string `json:"fullName"`
	DocumentNumber string `json:"documentNumber"`
	Phone          string `json:"phone"`
	Email          string `json:"email"`
	CreatedAt      string `json:"createdAt"`
}

// CustomerHandler exposes the customer registry.
type CustomerHandler struct {
	customer usecase.CustomerUseCase
}

// NewCustomerHandler wires the customer handler.
func NewCustomerHandler(customer usecase.CustomerUseCase) CustomerHandler {
	return CustomerHandler{customer: customer}
}

// Create registers a customer.
func (h CustomerHandler) Create(writer http.ResponseWriter, request *http.Request) {
	if _, err := requireAdministrator(request.Context()); err != nil {
		failure(writer, err)
		return
	}
	var payload createCustomerRequest
	if err := decode(writer, request, &payload); err != nil {
		failure(writer, err)
		return
	}
	fullName := sanitizeString(payload.FullName)
	documentNumber := sanitizeString(payload.DocumentNumber)
	phone := sanitizeString(payload.Phone)
	email := sanitizeString(payload.Email)

	created, err := h.customer.Create(
		request.Context(), fullName, documentNumber, phone, email,
	)
	if err != nil {
		failure(writer, err)
		return
	}
	respond(writer, http.StatusCreated, toCustomerResponse(created))
}

// List returns every registered customer.
func (h CustomerHandler) List(writer http.ResponseWriter, request *http.Request) {
	if _, err := requireAdministrator(request.Context()); err != nil {
		failure(writer, err)
		return
	}
	listed, err := h.customer.List(request.Context())
	if err != nil {
		failure(writer, err)
		return
	}
	payload := make([]customerResponse, 0, len(listed))
	for _, item := range listed {
		payload = append(payload, toCustomerResponse(item))
	}
	respond(writer, http.StatusOK, payload)
}

func toCustomerResponse(customer domain.Customer) customerResponse {
	return customerResponse{
		ID:             customer.ID,
		FullName:       customer.FullName,
		DocumentNumber: customer.DocumentNumber,
		Phone:          customer.Phone,
		Email:          customer.Email,
		CreatedAt:      formatTime(customer.CreatedAt),
	}
}

// formatTime renders a moment in RFC3339 so the interface can parse it with a
// single rule everywhere.
func formatTime(moment time.Time) string {
	if moment.IsZero() {
		return ""
	}
	return moment.UTC().Format(time.RFC3339)
}
