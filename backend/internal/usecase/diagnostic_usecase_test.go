package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeDiagnosticRepo struct {
	diagnostic map[string]domain.Diagnostic
}

func (f *fakeDiagnosticRepo) Save(_ context.Context, diagnostic domain.Diagnostic) error {
	f.diagnostic[diagnostic.ID] = diagnostic
	return nil
}

func (f *fakeDiagnosticRepo) FindByServiceOrder(_ context.Context, serviceOrderID string) (domain.Diagnostic, error) {
	for _, d := range f.diagnostic {
		if d.ServiceOrderID == serviceOrderID {
			return d, nil
		}
	}
	return domain.Diagnostic{}, domain.ErrNotFound
}

func (f *fakeDiagnosticRepo) ListByVehicle(_ context.Context, _ string) ([]domain.Diagnostic, error) {
	return nil, nil
}

func TestRecordDiagnosticRejectsDeliveredOrder(t *testing.T) {
	order := buildOrder(t, "order-1", "OS-0001")
	order.Status = domain.StatusDelivered
	orders := newFakeOrderRepository(order)

	assignments := &fakeAssignmentRepository{}
	assignment, _ := domain.NewAssignment("assignment-1", "order-1", "technician-1", fixedClock()())
	_ = assignments.Save(context.Background(), assignment)

	tech, _ := domain.NewTechnician("technician-1", "user-1", "Mecanica", time.Now())
	technicians := newFakeTechnicianRepository(tech)
	diagRepo := &fakeDiagnosticRepo{diagnostic: make(map[string]domain.Diagnostic)}

	useCase := usecase.NewDiagnosticUseCase(diagRepo, orders, assignments, technicians, sequentialID(), fixedClock())

	_, err := useCase.Record(context.Background(), "order-1", "user-1", "Falla detectada", "Inyector")
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("recording diagnostic on delivered order must return ErrInvalidTransition, got %v", err)
	}
}
