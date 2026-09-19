package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type fakeVehicleRepository struct {
	vehicle map[string]usecase.VehicleWithOwner
}

func newFakeVehicleRepository(id string) *fakeVehicleRepository {
	vehicle, _ := domain.NewVehicle(id, "customer-1", "ABC123", "VIN0001", "Mazda", "3", 2019, time.Now())
	return &fakeVehicleRepository{
		vehicle: map[string]usecase.VehicleWithOwner{
			id: {Vehicle: vehicle, OwnerID: "customer-1", OwnerName: "Ana Gomez"},
		},
	}
}

func (f *fakeVehicleRepository) Save(_ context.Context, vehicle domain.Vehicle) error {
	f.vehicle[vehicle.ID] = usecase.VehicleWithOwner{Vehicle: vehicle, OwnerID: vehicle.CustomerID}
	return nil
}

func (f *fakeVehicleRepository) List(_ context.Context) ([]usecase.VehicleWithOwner, error) {
	listed := make([]usecase.VehicleWithOwner, 0, len(f.vehicle))
	for _, item := range f.vehicle {
		listed = append(listed, item)
	}
	return listed, nil
}

func (f *fakeVehicleRepository) FindByID(_ context.Context, id string) (usecase.VehicleWithOwner, error) {
	found, ok := f.vehicle[id]
	if !ok {
		return usecase.VehicleWithOwner{}, domain.ErrNotFound
	}
	return found, nil
}

func TestOpenCreatesTheOrderInReceivedStatus(t *testing.T) {
	orders := newFakeOrderRepository()
	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), &fakeAssignmentRepository{}, newFakeTechnicianRepository(), nil, nil, sequentialID(), fixedClock(),
	)

	order, err := useCase.Open(context.Background(), "vehicle-1", "Ruido en el motor")
	if err != nil {
		t.Fatalf("opening a service order for a known vehicle must be accepted: %v", err)
	}
	if order.Status != domain.StatusReceived {
		t.Fatalf("a new order must start in RECEIVED, got %s", order.Status)
	}
	if order.OrderNumber != "OS-0001" {
		t.Fatalf("the order must carry the generated number, got %q", order.OrderNumber)
	}
}

func TestOpenRejectsAnUnknownVehicle(t *testing.T) {
	useCase := usecase.NewServiceOrderUseCase(
		newFakeOrderRepository(), newFakeVehicleRepository("vehicle-1"), &fakeAssignmentRepository{}, newFakeTechnicianRepository(), nil, nil, sequentialID(), fixedClock(),
	)
	if _, err := useCase.Open(context.Background(), "vehicle-9", "Ruido"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an unknown vehicle must be rejected as not found, got %v", err)
	}
}

func TestAdvanceRejectsAnOutOfLifecycleMoveAndWritesNothing(t *testing.T) {
	orders := newFakeOrderRepository(buildOrder(t, "order-1", "OS-0001"))
	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), &fakeAssignmentRepository{}, newFakeTechnicianRepository(), nil, nil, sequentialID(), fixedClock(),
	)

	_, err := useCase.Advance(context.Background(), "order-1", domain.StatusReady, "user-1", domain.RoleAdministrator)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("moving from RECEIVED to READY must be rejected, got %v", err)
	}
	if orders.updates != 0 {
		t.Fatalf("a rejected advance must write nothing, got %d writes", orders.updates)
	}
	stored, _ := orders.FindByID(context.Background(), "order-1")
	if stored.Status != domain.StatusReceived {
		t.Fatalf("the stored order must keep its previous status, got %s", stored.Status)
	}
}

func TestAdvanceWritesTheTransitionRecordWithItsAuthor(t *testing.T) {
	orders := newFakeOrderRepository(buildOrder(t, "order-1", "OS-0001"))
	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), &fakeAssignmentRepository{}, newFakeTechnicianRepository(), nil, nil, sequentialID(), fixedClock(),
	)

	if _, err := useCase.Advance(context.Background(), "order-1", domain.StatusInDiagnosis, "user-1", domain.RoleAdministrator); err != nil {
		t.Fatalf("moving from RECEIVED to IN_DIAGNOSIS must be accepted: %v", err)
	}
	history, _ := orders.ListTransition(context.Background(), "order-1")
	if len(history) != 1 {
		t.Fatalf("one transition record must be written, got %d", len(history))
	}
	if history[0].FromStatus != domain.StatusReceived || history[0].ToStatus != domain.StatusInDiagnosis {
		t.Fatalf("the transition record does not describe the move: %+v", history[0])
	}
	if history[0].ChangedByUserID != "user-1" {
		t.Fatalf("the transition must record its author, got %q", history[0].ChangedByUserID)
	}
}

func TestAdvanceToDeliveredReleasesTheTechnician(t *testing.T) {
	order := buildOrder(t, "order-1", "OS-0001")
	order.Status = domain.StatusReady
	orders := newFakeOrderRepository(order)
	assignments := &fakeAssignmentRepository{}
	assignment, err := domain.NewAssignment("assignment-1", "order-1", "technician-1", fixedClock()())
	if err != nil {
		t.Fatalf("building the assignment fixture failed: %v", err)
	}
	if err := assignments.Save(context.Background(), assignment); err != nil {
		t.Fatalf("storing the assignment fixture failed: %v", err)
	}
	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), assignments, newFakeTechnicianRepository(), nil, nil, sequentialID(), fixedClock(),
	)

	if _, err := useCase.Advance(context.Background(), "order-1", domain.StatusDelivered, "user-1", domain.RoleAdministrator); err != nil {
		t.Fatalf("moving from READY to DELIVERED must be accepted: %v", err)
	}
	if assignments.released != 1 {
		t.Fatalf("delivering the vehicle must release the technician, got %d releases", assignments.released)
	}
	if _, err := assignments.FindActiveByTechnician(context.Background(), "technician-1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the technician must be free after delivery, got %v", err)
	}
}

func TestAdvanceRejectsTechnicianNotAssignedToOrder(t *testing.T) {
	order := buildOrder(t, "order-1", "OS-0001")
	orders := newFakeOrderRepository(order)
	assignments := &fakeAssignmentRepository{}
	// Assign order-1 to technician-1 (user-1)
	assignment, _ := domain.NewAssignment("assignment-1", "order-1", "technician-1", fixedClock()())
	_ = assignments.Save(context.Background(), assignment)

	tech1, _ := domain.NewTechnician("technician-1", "user-1", "Mecanica", time.Now())
	tech2, _ := domain.NewTechnician("technician-2", "user-2", "Electricidad", time.Now())
	technicians := newFakeTechnicianRepository(tech1, tech2)

	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), assignments, technicians, nil, nil, sequentialID(), fixedClock(),
	)

	// user-2 (technician-2) tries to advance order-1
	_, err := useCase.Advance(context.Background(), "order-1", domain.StatusInDiagnosis, "user-2", domain.RoleTechnician)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("an unassigned technician advancing the order must be rejected with ErrForbidden, got %v", err)
	}

	// user-1 (technician-1, assigned) tries to advance order-1 -> must succeed
	if _, err := useCase.Advance(context.Background(), "order-1", domain.StatusInDiagnosis, "user-1", domain.RoleTechnician); err != nil {
		t.Fatalf("the assigned technician must be allowed to advance the order, got %v", err)
	}
}

func TestOpenRejectsVehicleWithExistingActiveOrder(t *testing.T) {
	order := buildOrder(t, "order-1", "OS-0001")
	orders := newFakeOrderRepository(order)
	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), &fakeAssignmentRepository{}, newFakeTechnicianRepository(), nil, nil, sequentialID(), fixedClock(),
	)

	// Trying to open a second order for vehicle-1 while order-1 is still open (RECEIVED)
	_, err := useCase.Open(context.Background(), "vehicle-1", "Falla de encendido")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("opening a new order for a vehicle with an active order must return ErrConflict, got %v", err)
	}

	// Mark order-1 as DELIVERED
	order.Status = domain.StatusDelivered
	_ = orders.Save(context.Background(), order)

	// Now opening a second order must succeed
	newOrder, err := useCase.Open(context.Background(), "vehicle-1", "Falla de encendido")
	if err != nil {
		t.Fatalf("opening a new order after previous order was delivered must succeed, got %v", err)
	}
	if newOrder.ID == "" {
		t.Fatalf("expected created order to have an ID")
	}
}

func TestFindRejectsTechnicianNotAssignedToOrder(t *testing.T) {
	order := buildOrder(t, "order-1", "OS-0001")
	orders := newFakeOrderRepository(order)
	assignments := &fakeAssignmentRepository{}
	assignment, _ := domain.NewAssignment("assignment-1", "order-1", "technician-1", fixedClock()())
	_ = assignments.Save(context.Background(), assignment)

	tech1, _ := domain.NewTechnician("technician-1", "user-1", "Mecanica", time.Now())
	tech2, _ := domain.NewTechnician("technician-2", "user-2", "Electricidad", time.Now())
	technicians := newFakeTechnicianRepository(tech1, tech2)

	useCase := usecase.NewServiceOrderUseCase(
		orders, newFakeVehicleRepository("vehicle-1"), assignments, technicians, nil, nil, sequentialID(), fixedClock(),
	)

	// Administrator can find any order
	if _, err := useCase.Find(context.Background(), "order-1", domain.RoleAdministrator, "admin-user"); err != nil {
		t.Fatalf("admin must be allowed to find any order, got %v", err)
	}

	// Assigned technician can find order
	if _, err := useCase.Find(context.Background(), "order-1", domain.RoleTechnician, "user-1"); err != nil {
		t.Fatalf("assigned technician must be allowed to find order, got %v", err)
	}

	// Unassigned technician is rejected with ErrForbidden
	if _, err := useCase.Find(context.Background(), "order-1", domain.RoleTechnician, "user-2"); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unassigned technician must be rejected with ErrForbidden, got %v", err)
	}
}
