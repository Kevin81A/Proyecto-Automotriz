package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"workshop/internal/domain"
	"workshop/internal/usecase"
)

// InMemoryStore holds an in-memory database suitable for local development and standalone mode.
type InMemoryStore struct {
	mu            sync.RWMutex
	Users         map[string]domain.User
	Customers     map[string]domain.Customer
	Vehicles      map[string]domain.Vehicle
	Technicians   map[string]domain.Technician
	Orders        map[string]domain.ServiceOrder
	Assignments   map[string]domain.Assignment
	Diagnostics   map[string]domain.Diagnostic
	Interventions map[string]domain.Intervention
	Warranties    map[string]domain.Warranty
	Transitions   map[string][]domain.StatusTransition
	orderSeq      int
}

// NewInMemoryStore creates and seeds a ready-to-use in-memory repository store.
func NewInMemoryStore() *InMemoryStore {
	now := time.Now()
	store := &InMemoryStore{
		Users:         make(map[string]domain.User),
		Customers:     make(map[string]domain.Customer),
		Vehicles:      make(map[string]domain.Vehicle),
		Technicians:   make(map[string]domain.Technician),
		Orders:        make(map[string]domain.ServiceOrder),
		Assignments:   make(map[string]domain.Assignment),
		Diagnostics:   make(map[string]domain.Diagnostic),
		Interventions: make(map[string]domain.Intervention),
		Warranties:    make(map[string]domain.Warranty),
		Transitions:   make(map[string][]domain.StatusTransition),
		orderSeq:      100,
	}

	// Password hashes for "Admin2026*"
	adminHash := "$2a$10$rlHYFOjjPwZBrfBFsJ5iKej7Zh1/lM112fWc1JAJ3dho9spgO0ZZ2"
	jperezHash := "$2a$10$UlVTcZBSsZTaNgt3ycKmfOI6ry3gpxZXePjrf2sT2m/VQ7mls71wm"
	lramirezHash := "$2a$10$YD1MEeeeYPeD3IjQojznc.rWF549bIkac0TOU4M2y4.1kc9XtuoQW"

	// 1. Users
	adminUser, _ := domain.NewUser("11111111-1111-4111-8111-111111111111", "admin", adminHash, "Jefe de taller", domain.RoleAdministrator, now)
	jperezUser, _ := domain.NewUser("22222222-2222-4222-8222-222222222222", "jperez", jperezHash, "Juan Perez", domain.RoleTechnician, now)
	lramirezUser, _ := domain.NewUser("33333333-3333-4333-8333-333333333333", "lramirez", lramirezHash, "Laura Ramirez", domain.RoleTechnician, now)

	store.Users["admin"] = adminUser
	store.Users["jperez"] = jperezUser
	store.Users["lramirez"] = lramirezUser

	// 2. Technicians
	tech1, _ := domain.NewTechnician("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1", jperezUser.ID, "Motor y transmision", now)
	tech2, _ := domain.NewTechnician("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2", lramirezUser.ID, "Frenos y suspension", now)
	store.Technicians[tech1.ID] = tech1
	store.Technicians[tech2.ID] = tech2

	// 3. Customers
	cust1, _ := domain.NewCustomer("cust-1111-1111", "Carlos Rodriguez", "10102020", "3001234567", "carlos@example.com", now)
	cust2, _ := domain.NewCustomer("cust-2222-2222", "Maria Gomez", "20203030", "3109876543", "maria@example.com", now)
	store.Customers[cust1.ID] = cust1
	store.Customers[cust2.ID] = cust2

	// 4. Vehicles
	veh1, _ := domain.NewVehicle("veh-1111-1111", cust1.ID, "ABC-123", "9BWZZZ377VT004251", "Toyota", "Corolla", 2022, now)
	veh2, _ := domain.NewVehicle("veh-2222-2222", cust2.ID, "XYZ-789", "1HGCR2F83HA001234", "Honda", "Civic", 2021, now)
	store.Vehicles[veh1.ID] = veh1
	store.Vehicles[veh2.ID] = veh2

	// 5. Service Orders
	// Order 1: Assigned to jperez (tech1)
	ord1, _ := domain.NewServiceOrder("ord-1111-1111", "SO-2026-0001", veh1.ID, "Ruido anormal en transmision al acelerar", now)
	ord1.Status = domain.StatusInDiagnosis
	store.Orders[ord1.ID] = ord1
	asgn1, _ := domain.NewAssignment("asgn-1111-1111", ord1.ID, tech1.ID, now)
	store.Assignments[ord1.ID] = asgn1

	// Order 2: Assigned to lramirez (tech2)
	ord2, _ := domain.NewServiceOrder("ord-2222-2222", "SO-2026-0002", veh2.ID, "Vibracion al frenar y cambio de pastillas", now)
	ord2.Status = domain.StatusInRepair
	store.Orders[ord2.ID] = ord2
	asgn2, _ := domain.NewAssignment("asgn-2222-2222", ord2.ID, tech2.ID, now)
	store.Assignments[ord2.ID] = asgn2

	// Diagnostic for Order 2
	diag2, _ := domain.NewDiagnostic("diag-2222-2222", ord2.ID, tech2.ID, "Desgaste severo en pastillas de freno delanteras", "Discos y pastillas delanteras", now)
	store.Diagnostics[ord2.ID] = diag2

	// Order 3: Unassigned (RECEIVED)
	ord3, _ := domain.NewServiceOrder("ord-3333-3333", "SO-2026-0003", veh1.ID, "Inspeccion de luces y cambio de aceite", now)
	ord3.Status = domain.StatusReceived
	store.Orders[ord3.ID] = ord3

	return store
}

// User methods
func (s *InMemoryStore) FindByUsername(_ context.Context, username string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.Users[username]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (s *InMemoryStore) FindByID(_ context.Context, id string) (domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.Users {
		if user.ID == id {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

// Customer methods
func (s *InMemoryStore) Save(_ context.Context, customer domain.Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Customers[customer.ID] = customer
	return nil
}

func (s *InMemoryStore) List(_ context.Context) ([]domain.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.Customer
	for _, c := range s.Customers {
		list = append(list, c)
	}
	return list, nil
}

func (s *InMemoryStore) FindCustomerByID(_ context.Context, id string) (domain.Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.Customers[id]
	if !ok {
		return domain.Customer{}, domain.ErrNotFound
	}
	return c, nil
}

// Vehicle methods
func (s *InMemoryStore) SaveVehicle(_ context.Context, vehicle domain.Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Vehicles[vehicle.ID] = vehicle
	return nil
}

func (s *InMemoryStore) ListVehicles(_ context.Context) ([]usecase.VehicleWithOwner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []usecase.VehicleWithOwner
	for _, v := range s.Vehicles {
		ownerName := "Desconocido"
		if c, ok := s.Customers[v.CustomerID]; ok {
			ownerName = c.FullName
		}
		list = append(list, usecase.VehicleWithOwner{
			Vehicle:   v,
			OwnerID:   v.CustomerID,
			OwnerName: ownerName,
		})
	}
	return list, nil
}

func (s *InMemoryStore) FindVehicleByID(_ context.Context, id string) (usecase.VehicleWithOwner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Vehicles[id]
	if !ok {
		return usecase.VehicleWithOwner{}, domain.ErrNotFound
	}
	ownerName := "Desconocido"
	if c, ok := s.Customers[v.CustomerID]; ok {
		ownerName = c.FullName
	}
	return usecase.VehicleWithOwner{
		Vehicle:   v,
		OwnerID:   v.CustomerID,
		OwnerName: ownerName,
	}, nil
}

// Technician methods
func (s *InMemoryStore) ListWorkload(_ context.Context) ([]domain.TechnicianWorkload, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.TechnicianWorkload
	for _, t := range s.Technicians {
		fullName := "Técnico"
		if u, ok := s.Users[t.UserID]; ok {
			fullName = u.FullName
		}
		busy := false
		activeOrderID := ""
		activeOrderNumber := ""
		activePlate := ""
		for _, a := range s.Assignments {
			if a.TechnicianID == t.ID && a.IsActive {
				busy = true
				activeOrderID = a.ServiceOrderID
				if o, ok := s.Orders[a.ServiceOrderID]; ok {
					activeOrderNumber = o.OrderNumber
					if v, ok := s.Vehicles[o.VehicleID]; ok {
						activePlate = v.Plate
					}
				}
				break
			}
		}
		list = append(list, domain.TechnicianWorkload{
			Technician:         t,
			FullName:           fullName,
			Busy:               busy,
			ActiveOrderID:      activeOrderID,
			ActiveOrderNumber:  activeOrderNumber,
			ActiveVehiclePlate: activePlate,
		})
	}
	return list, nil
}

func (s *InMemoryStore) FindTechnicianByID(_ context.Context, id string) (domain.Technician, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.Technicians[id]
	if !ok {
		return domain.Technician{}, domain.ErrNotFound
	}
	return t, nil
}

func (s *InMemoryStore) FindTechnicianByUserID(_ context.Context, userID string) (domain.Technician, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.Technicians {
		if t.UserID == userID {
			return t, nil
		}
	}
	return domain.Technician{}, domain.ErrNotFound
}

// ServiceOrder methods
func (s *InMemoryStore) SaveOrder(_ context.Context, order domain.ServiceOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Orders[order.ID] = order
	return nil
}

func (s *InMemoryStore) FindOrderByID(_ context.Context, id string) (domain.ServiceOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.Orders[id]
	if !ok {
		return domain.ServiceOrder{}, domain.ErrNotFound
	}
	return o, nil
}

func (s *InMemoryStore) ListOrders(_ context.Context, status string) ([]usecase.ServiceOrderSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []usecase.ServiceOrderSummary
	for _, o := range s.Orders {
		if status != "" && string(o.Status) != status {
			continue
		}
		plate := "N/A"
		if v, ok := s.Vehicles[o.VehicleID]; ok {
			plate = v.Plate
		}
		techName := "Sin asignar"
		techUserID := ""
		if a, ok := s.Assignments[o.ID]; ok && a.IsActive {
			if t, ok := s.Technicians[a.TechnicianID]; ok {
				techName = t.Specialty
				if u, ok := s.Users[t.UserID]; ok {
					techName = u.FullName
				}
				techUserID = t.UserID
			}
		}
		list = append(list, usecase.ServiceOrderSummary{
			Order:            o,
			VehiclePlate:     plate,
			TechnicianName:   techName,
			TechnicianUserID: techUserID,
		})
	}
	return list, nil
}

func (s *InMemoryStore) ListOrdersByVehicle(_ context.Context, vehicleID string) ([]domain.ServiceOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.ServiceOrder
	for _, o := range s.Orders {
		if o.VehicleID == vehicleID {
			list = append(list, o)
		}
	}
	return list, nil
}

func (s *InMemoryStore) UpdateOrderStatus(_ context.Context, order domain.ServiceOrder, transition domain.StatusTransition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Orders[order.ID] = order
	s.Transitions[order.ID] = append(s.Transitions[order.ID], transition)
	return nil
}

func (s *InMemoryStore) ListTransitions(_ context.Context, orderID string) ([]domain.StatusTransition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Transitions[orderID], nil
}

func (s *InMemoryStore) CountOrdersByStatus(_ context.Context) (map[string]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	counts := make(map[string]int)
	for _, o := range s.Orders {
		counts[string(o.Status)]++
	}
	return counts, nil
}

func (s *InMemoryStore) NextOrderNumber(_ context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orderSeq++
	return fmt.Sprintf("SO-2026-%04d", s.orderSeq), nil
}

// Assignment methods
func (s *InMemoryStore) SaveAssignment(_ context.Context, assignment domain.Assignment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Assignments[assignment.ServiceOrderID] = assignment
	return nil
}

func (s *InMemoryStore) FindActiveAssignmentByOrder(_ context.Context, orderID string) (domain.Assignment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.Assignments[orderID]
	if !ok || !a.IsActive {
		return domain.Assignment{}, domain.ErrNotFound
	}
	return a, nil
}

func (s *InMemoryStore) FindActiveAssignmentByTech(_ context.Context, techID string) (domain.Assignment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.Assignments {
		if a.TechnicianID == techID && a.IsActive {
			return a, nil
		}
	}
	return domain.Assignment{}, domain.ErrNotFound
}

func (s *InMemoryStore) ReleaseAssignmentByOrder(_ context.Context, orderID string, releasedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a, ok := s.Assignments[orderID]; ok {
		a.ReleasedAt = &releasedAt
		s.Assignments[orderID] = a
	}
	return nil
}

// Diagnostic methods
func (s *InMemoryStore) SaveDiagnostic(_ context.Context, diagnostic domain.Diagnostic) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Diagnostics[diagnostic.ServiceOrderID] = diagnostic
	return nil
}

func (s *InMemoryStore) FindDiagnosticByOrder(_ context.Context, orderID string) (domain.Diagnostic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.Diagnostics[orderID]
	if !ok {
		return domain.Diagnostic{}, domain.ErrNotFound
	}
	return d, nil
}

func (s *InMemoryStore) ListDiagnosticsByVehicle(_ context.Context, vehicleID string) ([]domain.Diagnostic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.Diagnostic
	for _, d := range s.Diagnostics {
		if o, ok := s.Orders[d.ServiceOrderID]; ok && o.VehicleID == vehicleID {
			list = append(list, d)
		}
	}
	return list, nil
}

// Intervention methods
func (s *InMemoryStore) SaveIntervention(_ context.Context, intervention domain.Intervention) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Interventions[intervention.ID] = intervention
	return nil
}

func (s *InMemoryStore) FindInterventionByID(_ context.Context, id string) (domain.Intervention, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.Interventions[id]
	if !ok {
		return domain.Intervention{}, domain.ErrNotFound
	}
	return i, nil
}

func (s *InMemoryStore) ListInterventionsByOrder(_ context.Context, orderID string) ([]domain.Intervention, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.Intervention
	for _, i := range s.Interventions {
		if i.ServiceOrderID == orderID {
			list = append(list, i)
		}
	}
	return list, nil
}

func (s *InMemoryStore) ListInterventionsByVehicle(_ context.Context, vehicleID string) ([]domain.Intervention, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.Intervention
	for _, i := range s.Interventions {
		if o, ok := s.Orders[i.ServiceOrderID]; ok && o.VehicleID == vehicleID {
			list = append(list, i)
		}
	}
	return list, nil
}

// Warranty methods
func (s *InMemoryStore) SaveWarranty(_ context.Context, warranty domain.Warranty) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Warranties[warranty.ID] = warranty
	return nil
}

func (s *InMemoryStore) FindWarrantyByID(_ context.Context, id string) (domain.Warranty, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.Warranties[id]
	if !ok {
		return domain.Warranty{}, domain.ErrNotFound
	}
	return w, nil
}

func (s *InMemoryStore) ListWarranties(_ context.Context) ([]usecase.WarrantyView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []usecase.WarrantyView
	now := time.Now()
	for _, w := range s.Warranties {
		plate := "N/A"
		orderNum := "N/A"
		desc := "Garantia de servicio"
		if i, ok := s.Interventions[w.InterventionID]; ok {
			desc = i.Description
			if o, ok := s.Orders[i.ServiceOrderID]; ok {
				orderNum = o.OrderNumber
				if v, ok := s.Vehicles[o.VehicleID]; ok {
					plate = v.Plate
				}
			}
		}
		list = append(list, usecase.WarrantyView{
			Warranty:                w,
			OrderNumber:             orderNum,
			VehiclePlate:            plate,
			InterventionDescription: desc,
			Valid:                   w.IsValidAt(now),
		})
	}
	return list, nil
}

func (s *InMemoryStore) ListWarrantiesByVehicle(_ context.Context, vehicleID string) ([]domain.Warranty, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []domain.Warranty
	for _, w := range s.Warranties {
		if i, ok := s.Interventions[w.InterventionID]; ok {
			if o, ok := s.Orders[i.ServiceOrderID]; ok && o.VehicleID == vehicleID {
				list = append(list, w)
			}
		}
	}
	return list, nil
}
