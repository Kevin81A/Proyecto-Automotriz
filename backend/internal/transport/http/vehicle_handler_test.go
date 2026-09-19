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

type fakeVehicleAdapter struct {
	vehicles map[string]usecase.VehicleWithOwner
}

func newFakeVehicleAdapter() *fakeVehicleAdapter {
	return &fakeVehicleAdapter{vehicles: make(map[string]usecase.VehicleWithOwner)}
}

func (f *fakeVehicleAdapter) Save(_ context.Context, v domain.Vehicle) error {
	f.vehicles[v.ID] = usecase.VehicleWithOwner{Vehicle: v, OwnerID: v.CustomerID, OwnerName: "Ana"}
	return nil
}

func (f *fakeVehicleAdapter) List(_ context.Context) ([]usecase.VehicleWithOwner, error) {
	listed := make([]usecase.VehicleWithOwner, 0, len(f.vehicles))
	for _, item := range f.vehicles {
		listed = append(listed, item)
	}
	return listed, nil
}

func (f *fakeVehicleAdapter) FindByID(_ context.Context, id string) (usecase.VehicleWithOwner, error) {
	v, ok := f.vehicles[id]
	if !ok {
		return usecase.VehicleWithOwner{}, domain.ErrNotFound
	}
	return v, nil
}

func newVehicleTestHandler(vehicles *fakeVehicleAdapter, customers *fakeCustomerStore) VehicleHandler {
	return NewVehicleHandler(usecase.NewVehicleUseCase(
		vehicles,
		customers,
		func() string { return "veh-1" },
		func() time.Time { return testMoment },
	))
}

func TestListVehicleIsForbiddenForATechnician(t *testing.T) {
	handler := newVehicleTestHandler(newFakeVehicleAdapter(), newFakeCustomerStore())
	request := asCaller(
		httptest.NewRequest(http.MethodGet, "/api/vehicle", nil), domain.RoleTechnician,
	)
	recorder := httptest.NewRecorder()
	handler.List(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("a technician must not list vehicles, got %d", recorder.Code)
	}
}

func TestListVehicleIsAllowedForAdministrator(t *testing.T) {
	handler := newVehicleTestHandler(newFakeVehicleAdapter(), newFakeCustomerStore())
	request := asCaller(
		httptest.NewRequest(http.MethodGet, "/api/vehicle", nil), domain.RoleAdministrator,
	)
	recorder := httptest.NewRecorder()
	handler.List(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("an administrator must be allowed to list vehicles, got %d", recorder.Code)
	}
}

func TestCreateVehicleSanitizesHTMLAndScriptTags(t *testing.T) {
	customers := newFakeCustomerStore()
	c, _ := domain.NewCustomer("cust-1", "Ana", "DOC-1", "555-0000", "ana@example.com", testMoment)
	_ = customers.Save(context.Background(), c)

	vehicles := newFakeVehicleAdapter()
	handler := newVehicleTestHandler(vehicles, customers)

	body := `{"customerId":"cust-1","plate":"ABC123","vin":"VIN12345678901234","brand":"Mazda <script>alert(1)</script>","model":"3<b>turbo</b>","modelYear":2020}`
	request := asCaller(
		httptest.NewRequest(http.MethodPost, "/api/vehicle", strings.NewReader(body)),
		domain.RoleAdministrator,
	)
	recorder := httptest.NewRecorder()
	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("creating vehicle must succeed, got %d body %s", recorder.Code, recorder.Body.String())
	}
	var payload vehicleResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshalling response failed: %v", err)
	}
	if strings.Contains(payload.Brand, "<script>") || strings.Contains(payload.Brand, "alert") {
		t.Fatalf("script tag was not stripped, got %q", payload.Brand)
	}
	if payload.Brand != "Mazda" {
		t.Fatalf("expected sanitized brand 'Mazda', got %q", payload.Brand)
	}
	if strings.Contains(payload.Model, "<b>") {
		t.Fatalf("HTML tag was not stripped, got %q", payload.Model)
	}
	if payload.Model != "3turbo" {
		t.Fatalf("expected sanitized model '3turbo', got %q", payload.Model)
	}
}
