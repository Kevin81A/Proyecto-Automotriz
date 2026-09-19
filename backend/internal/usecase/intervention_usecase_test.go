package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeInterventionRepo struct {
	intervention map[string]domain.Intervention
}

func (f *fakeInterventionRepo) Save(_ context.Context, intervention domain.Intervention) error {
	f.intervention[intervention.ID] = intervention
	return nil
}

func (f *fakeInterventionRepo) FindByID(_ context.Context, id string) (domain.Intervention, error) {
	item, ok := f.intervention[id]
	if !ok {
		return domain.Intervention{}, domain.ErrNotFound
	}
	return item, nil
}

func (f *fakeInterventionRepo) ListByServiceOrder(_ context.Context, serviceOrderID string) ([]domain.Intervention, error) {
	listed := make([]domain.Intervention, 0)
	for _, item := range f.intervention {
		if item.ServiceOrderID == serviceOrderID {
			listed = append(listed, item)
		}
	}
	return listed, nil
}

func (f *fakeInterventionRepo) ListByVehicle(_ context.Context, _ string) ([]domain.Intervention, error) {
	return nil, nil
}

func TestRegisterInterventionRejectsDeliveredOrder(t *testing.T) {
	order := buildOrder(t, "order-1", "OS-0001")
	order.Status = domain.StatusDelivered
	orders := newFakeOrderRepository(order)

	assignments := &fakeAssignmentRepository{}
	assignment, _ := domain.NewAssignment("assignment-1", "order-1", "technician-1", fixedClock()())
	_ = assignments.Save(context.Background(), assignment)

	tech, _ := domain.NewTechnician("technician-1", "user-1", "Mecanica", time.Now())
	technicians := newFakeTechnicianRepository(tech)
	interventionRepo := &fakeInterventionRepo{intervention: make(map[string]domain.Intervention)}

	useCase := usecase.NewInterventionUseCase(interventionRepo, orders, assignments, technicians, sequentialID(), fixedClock())

	_, err := useCase.Register(context.Background(), "order-1", "user-1", "Cambio de aceite", 1.5, nil)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("registering intervention on delivered order must return ErrInvalidTransition, got %v", err)
	}
}
