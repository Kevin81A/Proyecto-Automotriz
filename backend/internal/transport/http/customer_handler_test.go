package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeCustomerStore struct {
	customer map[string]domain.Customer
}

func newFakeCustomerStore() *fakeCustomerStore {
	return &fakeCustomerStore{customer: make(map[string]domain.Customer)}
}

func (f *fakeCustomerStore) Save(_ context.Context, customer domain.Customer) error {
	f.customer[customer.ID] = customer
	return nil
}

func (f *fakeCustomerStore) FindByID(_ context.Context, id string) (domain.Customer, error) {
	found, ok := f.customer[id]
	if !ok {
		return domain.Customer{}, domain.ErrNotFound
	}
	return found, nil
}

func (f *fakeCustomerStore) List(_ context.Context) ([]domain.Customer, error) {
	listed := make([]domain.Customer, 0, len(f.customer))
	for _, item := range f.customer {
		listed = append(listed, item)
	}
	return listed, nil
}

func newCustomerHandler(store *fakeCustomerStore) CustomerHandler {
	return NewCustomerHandler(usecase.NewCustomerUseCase(
		store,
		func() string { return "cust-1" },
		func() time.Time { return testMoment },
	))
}

func TestListCustomerIsForbiddenForATechnician(t *testing.T) {
	handler := newCustomerHandler(newFakeCustomerStore())
	request := asCaller(
		httptest.NewRequest(http.MethodGet, "/api/customer", nil), domain.RoleTechnician,
	)
	recorder := httptest.NewRecorder()
	handler.List(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("a technician must not list customers, got %d", recorder.Code)
	}
}

func TestListCustomerIsAllowedForAdministrator(t *testing.T) {
	handler := newCustomerHandler(newFakeCustomerStore())
	request := asCaller(
		httptest.NewRequest(http.MethodGet, "/api/customer", nil), domain.RoleAdministrator,
	)
	recorder := httptest.NewRecorder()
	handler.List(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("an administrator must be allowed to list customers, got %d", recorder.Code)
	}
}

func TestCreateCustomerSanitizesHTMLAndScriptTags(t *testing.T) {
	store := newFakeCustomerStore()
	handler := newCustomerHandler(store)
	body := `{"fullName":"O'Connor <script>alert(1)</script>","documentNumber":"DOC-999","phone":"555-1234","email":"test@example.com"}`
	request := asCaller(
		httptest.NewRequest(http.MethodPost, "/api/customer", strings.NewReader(body)),
		domain.RoleAdministrator,
	)
	recorder := httptest.NewRecorder()
	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("creating customer must succeed, got %d body %s", recorder.Code, recorder.Body.String())
	}
	var payload customerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshalling response failed: %v", err)
	}
	if strings.Contains(payload.FullName, "<script>") || strings.Contains(payload.FullName, "alert") {
		t.Fatalf("script tag was not stripped, got %q", payload.FullName)
	}
	if payload.FullName != "O'Connor" {
		t.Fatalf("expected sanitized name 'O'Connor', got %q", payload.FullName)
	}
}
