package usecase

import (
	"context"
	"fmt"
	"time"

	"workshop/internal/domain"
)

// ServiceOrderSummary is the read model of the order list: the order plus the
// plate of its vehicle and the name of the technician holding it.
type ServiceOrderSummary struct {
	Order            domain.ServiceOrder
	VehiclePlate     string
	TechnicianName   string
	TechnicianUserID string
}

// ServiceOrderRepository is the narrow port the service order use case needs.
// UpdateStatus writes the order and its transition record in one transaction.
type ServiceOrderRepository interface {
	Save(ctx context.Context, order domain.ServiceOrder) error
	FindByID(ctx context.Context, id string) (domain.ServiceOrder, error)
	List(ctx context.Context, status string) ([]ServiceOrderSummary, error)
	ListByVehicle(ctx context.Context, vehicleID string) ([]domain.ServiceOrder, error)
	UpdateStatus(ctx context.Context, order domain.ServiceOrder, transition domain.StatusTransition) error
	ListTransition(ctx context.Context, serviceOrderID string) ([]domain.StatusTransition, error)
	CountByStatus(ctx context.Context) (map[string]int, error)
	NextOrderNumber(ctx context.Context) (string, error)
}

// ServiceOrderUseCase opens orders at check-in and advances their lifecycle.
type ServiceOrderUseCase struct {
	order        ServiceOrderRepository
	vehicle      VehicleRepository
	assignment   AssignmentRepository
	technician   TechnicianRepository
	diagnostic   DiagnosticRepository
	intervention InterventionRepository
	newID        func() string
	now          func() time.Time
}

// NewServiceOrderUseCase wires the service order use case.
func NewServiceOrderUseCase(
	order ServiceOrderRepository,
	vehicle VehicleRepository,
	assignment AssignmentRepository,
	technician TechnicianRepository,
	diagnostic DiagnosticRepository,
	intervention InterventionRepository,
	newID func() string,
	now func() time.Time,
) ServiceOrderUseCase {
	return ServiceOrderUseCase{
		order:        order,
		vehicle:      vehicle,
		assignment:   assignment,
		technician:   technician,
		diagnostic:   diagnostic,
		intervention: intervention,
		newID:        newID,
		now:          now,
	}
}

// Open registers the check-in of a vehicle and returns the created order.
// A vehicle cannot have two active orders simultaneously.
func (s ServiceOrderUseCase) Open(ctx context.Context, vehicleID, reportedFailure string) (domain.ServiceOrder, error) {
	if _, err := s.vehicle.FindByID(ctx, vehicleID); err != nil {
		return domain.ServiceOrder{}, err
	}
	existingOrders, err := s.order.ListByVehicle(ctx, vehicleID)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	for _, existing := range existingOrders {
		if existing.Status.IsOpen() {
			return domain.ServiceOrder{}, fmt.Errorf("%w: vehicle already has an active service order (%s)", domain.ErrConflict, existing.OrderNumber)
		}
	}
	orderNumber, err := s.order.NextOrderNumber(ctx)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	order, err := domain.NewServiceOrder(s.newID(), orderNumber, vehicleID, reportedFailure, s.now())
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	if err := s.order.Save(ctx, order); err != nil {
		return domain.ServiceOrder{}, err
	}
	return order, nil
}

// List returns the orders, optionally filtered by a lifecycle status.
// For technicians, only orders assigned to them are returned.
func (s ServiceOrderUseCase) List(ctx context.Context, status string, actorRole domain.Role, actorUserID string) ([]ServiceOrderSummary, error) {
	listed, err := s.order.List(ctx, status)
	if err != nil {
		return nil, err
	}
	if actorRole != domain.RoleAdministrator {
		filtered := make([]ServiceOrderSummary, 0, len(listed))
		for _, item := range listed {
			if item.TechnicianUserID == actorUserID {
				filtered = append(filtered, item)
			}
		}
		return filtered, nil
	}
	return listed, nil
}

// Find returns one order by its identifier. Technicians can only view orders assigned to them.
func (s ServiceOrderUseCase) Find(ctx context.Context, orderID string, actorRole domain.Role, actorUserID string) (domain.ServiceOrder, error) {
	if actorRole != domain.RoleAdministrator {
		if s.assignment != nil && s.technician != nil {
			if _, err := requireAssignedTechnician(ctx, s.assignment, s.technician, orderID, actorUserID); err != nil {
				return domain.ServiceOrder{}, err
			}
		}
	}
	return s.order.FindByID(ctx, orderID)
}

// ListTransition returns the status history of an order.
func (s ServiceOrderUseCase) ListTransition(ctx context.Context, orderID string) ([]domain.StatusTransition, error) {
	return s.order.ListTransition(ctx, orderID)
}

// Advance moves an order to the next status. The domain rejects a move outside
// the lifecycle before anything is written, so the stored order is untouched.
// Non-administrators must be the technician currently assigned to the order.
// Advancing to IN_DIAGNOSIS requires a recorded diagnostic.
// Advancing to IN_REPAIR or READY requires at least one recorded intervention.
// Reaching DELIVERED releases the technician who held the order.
func (s ServiceOrderUseCase) Advance(ctx context.Context, orderID string, next domain.ServiceOrderStatus, actorUserID string, actorRole domain.Role) (domain.ServiceOrder, error) {
	if actorRole != domain.RoleAdministrator {
		if s.assignment != nil && s.technician != nil {
			if _, err := requireAssignedTechnician(ctx, s.assignment, s.technician, orderID, actorUserID); err != nil {
				return domain.ServiceOrder{}, err
			}
		}
	}
	if next == domain.StatusInDiagnosis && s.diagnostic != nil {
		if _, err := s.diagnostic.FindByServiceOrder(ctx, orderID); err != nil {
			return domain.ServiceOrder{}, fmt.Errorf("%w: cannot advance to %s without a recorded diagnostic", domain.ErrInvalidTransition, next)
		}
	}
	if (next == domain.StatusInRepair || next == domain.StatusReady) && s.intervention != nil {
		interventions, err := s.intervention.ListByServiceOrder(ctx, orderID)
		if err != nil || len(interventions) == 0 {
			return domain.ServiceOrder{}, fmt.Errorf("%w: cannot advance to %s without at least one recorded intervention", domain.ErrInvalidTransition, next)
		}
	}
	order, err := s.order.FindByID(ctx, orderID)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	changedAt := s.now()
	if !changedAt.After(order.UpdatedAt) {
		changedAt = order.UpdatedAt.Add(time.Second)
	}
	transition, err := order.MoveTo(next, s.newID(), actorUserID, changedAt)
	if err != nil {
		return domain.ServiceOrder{}, err
	}
	if err := s.order.UpdateStatus(ctx, order, transition); err != nil {
		return domain.ServiceOrder{}, err
	}
	if next == domain.StatusDelivered {
		if err := s.assignment.ReleaseByServiceOrder(ctx, order.ID, changedAt); err != nil {
			return domain.ServiceOrder{}, err
		}
	}
	return order, nil
}
