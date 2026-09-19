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

func TestAuditBAC_TechnicianBlockedFromSensitiveEndpoints(t *testing.T) {
	issuer := NewTokenIssuer("secret-test-key-32-chars-long!!", time.Hour)
	now := time.Now

	customerHandler := NewCustomerHandler(usecase.NewCustomerUseCase(newFakeCustomerStore(), func() string { return "id-1" }, now))
	vehicleHandler := NewVehicleHandler(usecase.NewVehicleUseCase(newFakeVehicleStore("v-1"), newFakeCustomerStore(), func() string { return "id-2" }, now))
	technicianHandler := NewTechnicianHandler(usecase.NewTechnicianUseCase(fakeTechnicianWorkloadRepo{}))
	warrantyHandler := NewWarrantyHandler(usecase.NewWarrantyUseCase(&fakeWarrantyRepo{}, &fakeInterventionRepo{}, func() string { return "id-3" }, now))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/customer", requireAdministratorMiddleware(customerHandler.List))
	mux.HandleFunc("GET /api/vehicle", requireAdministratorMiddleware(vehicleHandler.List))
	mux.HandleFunc("GET /api/technician", requireAdministratorMiddleware(technicianHandler.List))
	mux.HandleFunc("GET /api/warranty", requireAdministratorMiddleware(warrantyHandler.List))

	handler := authMiddleware(issuer, now)(mux)

	// Issue token for technician jperez
	techToken, _, err := issuer.Issue("user-jperez", domain.RoleTechnician, now())
	if err != nil {
		t.Fatalf("issuing technician token failed: %v", err)
	}

	// Issue token for administrator
	adminToken, _, err := issuer.Issue("user-admin", domain.RoleAdministrator, now())
	if err != nil {
		t.Fatalf("issuing admin token failed: %v", err)
	}

	endpoints := []string{"/api/customer", "/api/vehicle", "/api/technician", "/api/warranty"}

	for _, endpoint := range endpoints {
		// Test Technician -> must be 403 Forbidden
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req.Header.Set("Authorization", "Bearer "+techToken)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Technician accessing %s MUST receive 403 Forbidden, got %d", endpoint, rec.Code)
		}

		// Test Administrator -> must be 200 OK
		adminReq := httptest.NewRequest(http.MethodGet, endpoint, nil)
		adminReq.Header.Set("Authorization", "Bearer "+adminToken)
		adminRec := httptest.NewRecorder()
		handler.ServeHTTP(adminRec, adminReq)

		if adminRec.Code != http.StatusOK {
			t.Fatalf("Administrator accessing %s MUST receive 200 OK, got %d", endpoint, adminRec.Code)
		}
	}
}

func TestAuditBOLA_TechnicianBlockedFromUnassignedOrder(t *testing.T) {
	issuer := NewTokenIssuer("secret-test-key-32-chars-long!!", time.Hour)
	now := time.Now

	order := orderFixture(t, "order-1", domain.StatusReceived)
	orders := newFakeOrderStore(order)
	assignments := &fakeAssignmentStore{}

	// Assign order-1 to technician-jperez (user-jperez)
	assignment, _ := domain.NewAssignment("assign-1", "order-1", "tech-jperez", now())
	_ = assignments.Save(context.Background(), assignment)

	techStore := &fakeTechnicianStoreFixture{
		byUserID: map[string]domain.Technician{
			"user-jperez":   {ID: "tech-jperez", UserID: "user-jperez", Specialty: "Motor", CreatedAt: now()},
			"user-lramirez": {ID: "tech-lramirez", UserID: "user-lramirez", Specialty: "Frenos", CreatedAt: now()},
		},
	}

	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleStore("v-1"), assignments, techStore,
		nil, nil, func() string { return "t-1" }, now,
	)
	orderHandler := NewServiceOrderHandler(useCase)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/service-order/{serviceOrderId}/status", orderHandler.Advance)
	handler := authMiddleware(issuer, now)(mux)

	// Issue token for lramirez (NOT assigned to order-1)
	lramirezToken, _, _ := issuer.Issue("user-lramirez", domain.RoleTechnician, now())

	req := httptest.NewRequest(
		http.MethodPost, "/api/service-order/order-1/status",
		strings.NewReader(`{"status":"IN_DIAGNOSIS"}`),
	)
	req.Header.Set("Authorization", "Bearer "+lramirezToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("Unassigned technician lramirez advancing order-1 MUST receive 403 Forbidden, got %d body %s", rec.Code, rec.Body.String())
	}

	// Issue token for jperez (assigned technician) -> allowed
	jperezToken, _, _ := issuer.Issue("user-jperez", domain.RoleTechnician, now())

	jperezReq := httptest.NewRequest(
		http.MethodPost, "/api/service-order/order-1/status",
		strings.NewReader(`{"status":"IN_DIAGNOSIS"}`),
	)
	jperezReq.Header.Set("Authorization", "Bearer "+jperezToken)
	jperezRec := httptest.NewRecorder()
	handler.ServeHTTP(jperezRec, jperezReq)

	if jperezRec.Code != http.StatusOK {
		t.Fatalf("Assigned technician jperez advancing order-1 MUST receive 200 OK, got %d body %s", jperezRec.Code, jperezRec.Body.String())
	}
}

func TestAuditRateLimit_BruteForceBlockedAfter5Attempts(t *testing.T) {
	now := time.Now()
	rateLimiter := NewRateLimiter(5, 15*time.Minute, func() time.Time { return now })

	dummyAuth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failure(w, domain.ErrUnauthorized)
	})

	protectedHandler := rateLimiter.Middleware(dummyAuth)

	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/session", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rec := httptest.NewRecorder()
		protectedHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d should be 401 Unauthorized, got %d", i, rec.Code)
		}
	}

	// 6th attempt must be rejected with 429 Too Many Requests
	req := httptest.NewRequest(http.MethodPost, "/api/session", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("6th attempt MUST be rejected with 429 Too Many Requests, got %d", rec.Code)
	}

	var payload errorPayload
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal rate limit error payload: %v", err)
	}
	if payload.Code != "too_many_requests" {
		t.Fatalf("expected code 'too_many_requests', got %q", payload.Code)
	}
}

func TestAuditSanitization_XSSAndSQLInjectionNeutralized(t *testing.T) {
	customerStore := newFakeCustomerStore()
	existingCustomer, _ := domain.NewCustomer("customer-1", "Ana Gomez", "DOC-1234", "555-1234", "ana@example.com", time.Now())
	_ = customerStore.Save(context.Background(), existingCustomer)
	customerHandler := newCustomerHandler(customerStore)

	// Payload with stored XSS
	body := `{"fullName":"O'Connor <script>alert('XSS')</script>","documentNumber":"DOC-1234","phone":"555-4321","email":"clean@email.com"}`
	req := asCaller(httptest.NewRequest(http.MethodPost, "/api/customer", strings.NewReader(body)), domain.RoleAdministrator)
	rec := httptest.NewRecorder()
	customerHandler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("valid sanitized customer must be created with 201, got %d body %s", rec.Code, rec.Body.String())
	}
	var customer customerResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &customer)

	if strings.Contains(customer.FullName, "<script>") || strings.Contains(customer.FullName, "alert") {
		t.Fatalf("XSS was not stripped from customer full name: %q", customer.FullName)
	}
	if customer.FullName != "O'Connor" {
		t.Fatalf("expected sanitized name 'O'Connor', got %q", customer.FullName)
	}

	// Malicious vehicle payload with SQL injection in plate
	vehicleHandler := newVehicleTestHandler(newFakeVehicleAdapter(), customerStore)
	badPlateBody := `{"customerId":"customer-1","plate":"' OR 1=1--","vin":"VIN123456789","brand":"Toyota","model":"Corolla","modelYear":2022}`
	badReq := asCaller(httptest.NewRequest(http.MethodPost, "/api/vehicle", strings.NewReader(badPlateBody)), domain.RoleAdministrator)
	badRec := httptest.NewRecorder()
	vehicleHandler.Create(badRec, badReq)

	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("vehicle with SQL injection in plate MUST be rejected with 400 Bad Request, got %d", badRec.Code)
	}
}

type fakeTechnicianStoreFixture struct {
	byUserID map[string]domain.Technician
}

func (f *fakeTechnicianStoreFixture) ListWorkload(_ context.Context) ([]domain.TechnicianWorkload, error) {
	return nil, nil
}
func (f *fakeTechnicianStoreFixture) FindByID(_ context.Context, id string) (domain.Technician, error) {
	for _, tech := range f.byUserID {
		if tech.ID == id {
			return tech, nil
		}
	}
	return domain.Technician{}, domain.ErrNotFound
}
func (f *fakeTechnicianStoreFixture) FindByUserID(_ context.Context, userID string) (domain.Technician, error) {
	tech, ok := f.byUserID[userID]
	if !ok {
		return domain.Technician{}, domain.ErrNotFound
	}
	return tech, nil
}

type fakeWarrantyRepo struct{}
func (f *fakeWarrantyRepo) Save(_ context.Context, _ domain.Warranty) error { return nil }
func (f *fakeWarrantyRepo) FindByID(_ context.Context, _ string) (domain.Warranty, error) { return domain.Warranty{}, domain.ErrNotFound }
func (f *fakeWarrantyRepo) List(_ context.Context) ([]usecase.WarrantyView, error) { return nil, nil }
func (f *fakeWarrantyRepo) ListByVehicle(_ context.Context, _ string) ([]domain.Warranty, error) { return nil, nil }

type fakeInterventionRepo struct{}
func (f *fakeInterventionRepo) Save(_ context.Context, _ domain.Intervention) error { return nil }
func (f *fakeInterventionRepo) FindByID(_ context.Context, _ string) (domain.Intervention, error) { return domain.Intervention{}, domain.ErrNotFound }
func (f *fakeInterventionRepo) ListByServiceOrder(_ context.Context, _ string) ([]domain.Intervention, error) { return nil, nil }
func (f *fakeInterventionRepo) ListByVehicle(_ context.Context, _ string) ([]domain.Intervention, error) { return nil, nil }
