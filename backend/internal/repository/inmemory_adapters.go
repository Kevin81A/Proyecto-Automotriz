package repository

import (
	"context"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

type InMemoryUserRepo struct{ Store *InMemoryStore }

func (r InMemoryUserRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	return r.Store.FindByUsername(ctx, username)
}
func (r InMemoryUserRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	return r.Store.FindByID(ctx, id)
}

type InMemoryCustomerRepo struct{ Store *InMemoryStore }

func (r InMemoryCustomerRepo) Save(ctx context.Context, customer domain.Customer) error {
	return r.Store.Save(ctx, customer)
}
func (r InMemoryCustomerRepo) List(ctx context.Context) ([]domain.Customer, error) {
	return r.Store.List(ctx)
}
func (r InMemoryCustomerRepo) FindByID(ctx context.Context, id string) (domain.Customer, error) {
	return r.Store.FindCustomerByID(ctx, id)
}

type InMemoryVehicleRepo struct{ Store *InMemoryStore }

func (r InMemoryVehicleRepo) Save(ctx context.Context, vehicle domain.Vehicle) error {
	return r.Store.SaveVehicle(ctx, vehicle)
}
func (r InMemoryVehicleRepo) List(ctx context.Context) ([]usecase.VehicleWithOwner, error) {
	return r.Store.ListVehicles(ctx)
}
func (r InMemoryVehicleRepo) FindByID(ctx context.Context, id string) (usecase.VehicleWithOwner, error) {
	return r.Store.FindVehicleByID(ctx, id)
}

type InMemoryTechnicianRepo struct{ Store *InMemoryStore }

func (r InMemoryTechnicianRepo) ListWorkload(ctx context.Context) ([]domain.TechnicianWorkload, error) {
	return r.Store.ListWorkload(ctx)
}
func (r InMemoryTechnicianRepo) FindByID(ctx context.Context, id string) (domain.Technician, error) {
	return r.Store.FindTechnicianByID(ctx, id)
}
func (r InMemoryTechnicianRepo) FindByUserID(ctx context.Context, userID string) (domain.Technician, error) {
	return r.Store.FindTechnicianByUserID(ctx, userID)
}

type InMemoryOrderRepo struct{ Store *InMemoryStore }

func (r InMemoryOrderRepo) Save(ctx context.Context, order domain.ServiceOrder) error {
	return r.Store.SaveOrder(ctx, order)
}
func (r InMemoryOrderRepo) FindByID(ctx context.Context, id string) (domain.ServiceOrder, error) {
	return r.Store.FindOrderByID(ctx, id)
}
func (r InMemoryOrderRepo) List(ctx context.Context, status string) ([]usecase.ServiceOrderSummary, error) {
	return r.Store.ListOrders(ctx, status)
}
func (r InMemoryOrderRepo) ListByVehicle(ctx context.Context, vehicleID string) ([]domain.ServiceOrder, error) {
	return r.Store.ListOrdersByVehicle(ctx, vehicleID)
}
func (r InMemoryOrderRepo) UpdateStatus(ctx context.Context, order domain.ServiceOrder, transition domain.StatusTransition) error {
	return r.Store.UpdateOrderStatus(ctx, order, transition)
}
func (r InMemoryOrderRepo) ListTransition(ctx context.Context, serviceOrderID string) ([]domain.StatusTransition, error) {
	return r.Store.ListTransitions(ctx, serviceOrderID)
}
func (r InMemoryOrderRepo) CountByStatus(ctx context.Context) (map[string]int, error) {
	return r.Store.CountOrdersByStatus(ctx)
}
func (r InMemoryOrderRepo) NextOrderNumber(ctx context.Context) (string, error) {
	return r.Store.NextOrderNumber(ctx)
}

type InMemoryAssignmentRepo struct{ Store *InMemoryStore }

func (r InMemoryAssignmentRepo) Save(ctx context.Context, assignment domain.Assignment) error {
	return r.Store.SaveAssignment(ctx, assignment)
}
func (r InMemoryAssignmentRepo) FindActiveByServiceOrder(ctx context.Context, serviceOrderID string) (domain.Assignment, error) {
	return r.Store.FindActiveAssignmentByOrder(ctx, serviceOrderID)
}
func (r InMemoryAssignmentRepo) FindActiveByTechnician(ctx context.Context, technicianID string) (domain.Assignment, error) {
	return r.Store.FindActiveAssignmentByTech(ctx, technicianID)
}
func (r InMemoryAssignmentRepo) ReleaseByServiceOrder(ctx context.Context, serviceOrderID string, releasedAt time.Time) error {
	return r.Store.ReleaseAssignmentByOrder(ctx, serviceOrderID, releasedAt)
}

type InMemoryDiagnosticRepo struct{ Store *InMemoryStore }

func (r InMemoryDiagnosticRepo) Save(ctx context.Context, diagnostic domain.Diagnostic) error {
	return r.Store.SaveDiagnostic(ctx, diagnostic)
}
func (r InMemoryDiagnosticRepo) FindByServiceOrder(ctx context.Context, serviceOrderID string) (domain.Diagnostic, error) {
	return r.Store.FindDiagnosticByOrder(ctx, serviceOrderID)
}
func (r InMemoryDiagnosticRepo) ListByVehicle(ctx context.Context, vehicleID string) ([]domain.Diagnostic, error) {
	return r.Store.ListDiagnosticsByVehicle(ctx, vehicleID)
}

type InMemoryInterventionRepo struct{ Store *InMemoryStore }

func (r InMemoryInterventionRepo) Save(ctx context.Context, intervention domain.Intervention) error {
	return r.Store.SaveIntervention(ctx, intervention)
}
func (r InMemoryInterventionRepo) FindByID(ctx context.Context, id string) (domain.Intervention, error) {
	return r.Store.FindInterventionByID(ctx, id)
}
func (r InMemoryInterventionRepo) ListByServiceOrder(ctx context.Context, serviceOrderID string) ([]domain.Intervention, error) {
	return r.Store.ListInterventionsByOrder(ctx, serviceOrderID)
}
func (r InMemoryInterventionRepo) ListByVehicle(ctx context.Context, vehicleID string) ([]domain.Intervention, error) {
	return r.Store.ListInterventionsByVehicle(ctx, vehicleID)
}

type InMemoryWarrantyRepo struct{ Store *InMemoryStore }

func (r InMemoryWarrantyRepo) Save(ctx context.Context, warranty domain.Warranty) error {
	return r.Store.SaveWarranty(ctx, warranty)
}
func (r InMemoryWarrantyRepo) FindByID(ctx context.Context, id string) (domain.Warranty, error) {
	return r.Store.FindWarrantyByID(ctx, id)
}
func (r InMemoryWarrantyRepo) List(ctx context.Context) ([]usecase.WarrantyView, error) {
	return r.Store.ListWarranties(ctx)
}
func (r InMemoryWarrantyRepo) ListByVehicle(ctx context.Context, vehicleID string) ([]domain.Warranty, error) {
	return r.Store.ListWarrantiesByVehicle(ctx, vehicleID)
}
