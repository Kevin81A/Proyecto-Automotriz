package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeTechnicianWorkloadRepo struct{}

func (f fakeTechnicianWorkloadRepo) ListWorkload(_ context.Context) ([]domain.TechnicianWorkload, error) {
	return []domain.TechnicianWorkload{
		{
			Technician: domain.Technician{ID: "t-1", UserID: "u-1", Specialty: "Motor", CreatedAt: time.Now()},
			FullName:   "Juan Perez",
			Busy:       false,
		},
	}, nil
}

func (f fakeTechnicianWorkloadRepo) FindByID(_ context.Context, _ string) (domain.Technician, error) {
	return domain.Technician{}, domain.ErrNotFound
}

func (f fakeTechnicianWorkloadRepo) FindByUserID(_ context.Context, _ string) (domain.Technician, error) {
	return domain.Technician{}, domain.ErrNotFound
}

func TestTechnicianListRequiresAdministrator(t *testing.T) {
	handler := NewTechnicianHandler(usecase.NewTechnicianUseCase(fakeTechnicianWorkloadRepo{}))
	protectedList := requireAdministratorMiddleware(handler.List)

	// Technician attempt -> 403 Forbidden
	req := asCaller(httptest.NewRequest(http.MethodGet, "/api/technician", nil), domain.RoleTechnician)
	rec := httptest.NewRecorder()
	protectedList(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("technician accessing /api/technician must receive 403 Forbidden, got %d", rec.Code)
	}

	// Administrator attempt -> 200 OK
	adminReq := asCaller(httptest.NewRequest(http.MethodGet, "/api/technician", nil), domain.RoleAdministrator)
	adminRec := httptest.NewRecorder()
	protectedList(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("administrator accessing /api/technician must receive 200 OK, got %d", adminRec.Code)
	}
}
