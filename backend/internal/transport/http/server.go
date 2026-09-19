package http

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"

	"workshop/internal/config"
	"workshop/internal/repository"
	"workshop/internal/usecase"
)

// Dependency is everything the API needs from the outside world.
type Dependency struct {
	Database *sql.DB
	Config   config.Config
	Now      func() time.Time
	NewID    func() string
}

// Repositories groups the ports needed by the HTTP layer.
type Repositories struct {
	User         usecase.UserRepository
	Customer     usecase.CustomerRepository
	Vehicle      usecase.VehicleRepository
	Technician   usecase.TechnicianRepository
	Order        usecase.ServiceOrderRepository
	Assignment   usecase.AssignmentRepository
	Diagnostic   usecase.DiagnosticRepository
	Intervention usecase.InterventionRepository
	Warranty     usecase.WarrantyRepository
}

// NewServer builds the HTTP server using a live MySQL connection pool.
func NewServer(dependency Dependency) *http.Server {
	now := dependency.Now
	if now == nil {
		now = time.Now
	}
	newID := dependency.NewID
	if newID == nil {
		newID = func() string { return uuid.NewString() }
	}
	settings := dependency.Config
	timeout := settings.DatabaseTimeout

	repos := Repositories{
		User:         repository.NewUserRepository(dependency.Database, timeout),
		Customer:     repository.NewCustomerRepository(dependency.Database, timeout),
		Vehicle:      repository.NewVehicleRepository(dependency.Database, timeout),
		Technician:   repository.NewTechnicianRepository(dependency.Database, timeout),
		Order:        repository.NewServiceOrderRepository(dependency.Database, timeout),
		Assignment:   repository.NewAssignmentRepository(dependency.Database, timeout),
		Diagnostic:   repository.NewDiagnosticRepository(dependency.Database, timeout),
		Intervention: repository.NewInterventionRepository(dependency.Database, timeout),
		Warranty:     repository.NewWarrantyRepository(dependency.Database, timeout),
	}
	return NewServerWithRepositories(repos, settings, now, newID)
}

// NewStandaloneServer builds an HTTP server using the in-memory repository store.
func NewStandaloneServer(settings config.Config) *http.Server {
	store := repository.NewInMemoryStore()
	now := time.Now
	newID := func() string { return uuid.NewString() }

	repos := Repositories{
		User:         repository.InMemoryUserRepo{Store: store},
		Customer:     repository.InMemoryCustomerRepo{Store: store},
		Vehicle:      repository.InMemoryVehicleRepo{Store: store},
		Technician:   repository.InMemoryTechnicianRepo{Store: store},
		Order:        repository.InMemoryOrderRepo{Store: store},
		Assignment:   repository.InMemoryAssignmentRepo{Store: store},
		Diagnostic:   repository.InMemoryDiagnosticRepo{Store: store},
		Intervention: repository.InMemoryInterventionRepo{Store: store},
		Warranty:     repository.InMemoryWarrantyRepo{Store: store},
	}
	return NewServerWithRepositories(repos, settings, now, newID)
}

// NewServerWithRepositories wires handlers from any repository implementation.
func NewServerWithRepositories(repos Repositories, settings config.Config, now func() time.Time, newID func() string) *http.Server {
	issuer := NewTokenIssuer(settings.TokenSecret, settings.TokenTTL)

	authHandler := NewAuthHandler(usecase.NewAuthenticateUser(repos.User, issuer, now))
	customerHandler := NewCustomerHandler(usecase.NewCustomerUseCase(repos.Customer, newID, now))
	vehicleHandler := NewVehicleHandler(usecase.NewVehicleUseCase(repos.Vehicle, repos.Customer, newID, now))
	technicianHandler := NewTechnicianHandler(usecase.NewTechnicianUseCase(repos.Technician))
	orderUseCase := usecase.NewServiceOrderUseCase(
		repos.Order, repos.Vehicle, repos.Assignment, repos.Technician,
		repos.Diagnostic, repos.Intervention, newID, now,
	)
	orderHandler := NewServiceOrderHandler(orderUseCase)
	assignmentHandler := NewAssignmentHandler(
		usecase.NewAssignmentUseCase(repos.Assignment, repos.Order, repos.Technician, newID, now),
	)
	diagnosticHandler := NewDiagnosticHandler(usecase.NewDiagnosticUseCase(
		repos.Diagnostic, repos.Order, repos.Assignment, repos.Technician, newID, now,
	))
	interventionHandler := NewInterventionHandler(usecase.NewInterventionUseCase(
		repos.Intervention, repos.Order, repos.Assignment, repos.Technician, newID, now,
	))
	warrantyHandler := NewWarrantyHandler(usecase.NewWarrantyUseCase(repos.Warranty, repos.Intervention, newID, now))
	timelineHandler := NewTimelineHandler(usecase.NewTimelineUseCase(
		repos.Vehicle, repos.Order, repos.Diagnostic, repos.Intervention, repos.Warranty,
	))
	dashboardHandler := NewDashboardHandler(usecase.NewDashboardUseCase(repos.Order, repos.Technician))

	rateLimiter := NewRateLimiter(30, 30*time.Second, now)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/customer", requireAdministratorMiddleware(customerHandler.List))
	protected.HandleFunc("POST /api/customer", customerHandler.Create)
	protected.HandleFunc("GET /api/vehicle", requireAdministratorMiddleware(vehicleHandler.List))
	protected.HandleFunc("POST /api/vehicle", vehicleHandler.Create)
	protected.HandleFunc("GET /api/vehicle/{vehicleId}/timeline", timelineHandler.Build)
	protected.HandleFunc("GET /api/technician", requireAdministratorMiddleware(technicianHandler.List))
	protected.HandleFunc("GET /api/service-order", orderHandler.List)
	protected.HandleFunc("POST /api/service-order", orderHandler.Create)
	protected.HandleFunc("GET /api/service-order/{serviceOrderId}", orderHandler.Find)
	protected.HandleFunc("GET /api/service-order/{serviceOrderId}/transition", orderHandler.ListTransition)
	protected.HandleFunc("POST /api/service-order/{serviceOrderId}/status", orderHandler.Advance)
	protected.HandleFunc("GET /api/service-order/{serviceOrderId}/assignment", assignmentHandler.Find)
	protected.HandleFunc("POST /api/service-order/{serviceOrderId}/assignment", assignmentHandler.Assign)
	protected.HandleFunc("GET /api/service-order/{serviceOrderId}/diagnostic", diagnosticHandler.Find)
	protected.HandleFunc("POST /api/service-order/{serviceOrderId}/diagnostic", diagnosticHandler.Record)
	protected.HandleFunc("GET /api/service-order/{serviceOrderId}/intervention", interventionHandler.List)
	protected.HandleFunc("POST /api/service-order/{serviceOrderId}/intervention", interventionHandler.Register)
	protected.HandleFunc("GET /api/warranty", requireAdministratorMiddleware(warrantyHandler.List))
	protected.HandleFunc("POST /api/warranty", warrantyHandler.Issue)
	protected.HandleFunc("GET /api/dashboard", dashboardHandler.Build)

	root := http.NewServeMux()
	root.HandleFunc("GET /api/health", health)
	root.HandleFunc("POST /api/session", rateLimiter.Middleware(authHandler.SignIn))
	root.Handle("/api/", authMiddleware(issuer, now)(protected))

	return &http.Server{
		Addr:              ":" + settings.HTTPPort,
		Handler:           chain(root, corsMiddleware(settings.AllowedOrigin)),
		ReadTimeout:       settings.RequestTimeout,
		ReadHeaderTimeout: settings.RequestTimeout,
		WriteTimeout:      settings.RequestTimeout,
		IdleTimeout:       2 * settings.RequestTimeout,
	}
}

func health(writer http.ResponseWriter, _ *http.Request) {
	respond(writer, http.StatusOK, map[string]string{"status": "ok"})
}
