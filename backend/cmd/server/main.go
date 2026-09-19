package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"workshop/internal/config"
	"workshop/internal/repository"
	transport "workshop/internal/transport/http"
)

// main bootstraps the process: it reads the configuration, attempts to open
// the database pool (or gracefully falls back to seeded in-memory store in standalone mode),
// and starts the HTTP server.
func main() {
	settings, err := config.Load()
	if err != nil {
		log.Printf("Aviso de configuracion: %v. Usando configuracion por defecto de desarrollo.", err)
		settings = config.DefaultConfig()
	}

	var server *http.Server

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	database, err := repository.Open(ctx, settings.DatabaseDSN, settings.DatabaseTimeout)
	cancel()

	if err != nil {
		log.Printf("[MODO AUTONOMO / EN MEMORIA] Base de datos MySQL no detectada (%v).", err)
		log.Printf("[MODO AUTONOMO / EN MEMORIA] Servidor iniciado con datos semilla precargados.")
		log.Printf("[CREDENCIALES DISPONIBLES]:")
		log.Printf("  - Jefe de taller:  admin    / Admin2026*")
		log.Printf("  - Tecnico 1:       jperez   / Admin2026*")
		log.Printf("  - Tecnico 2:       lramirez / Admin2026*")
		server = transport.NewStandaloneServer(settings)
	} else {
		defer func() { _ = database.Close() }()
		log.Printf("[MODO MYSQL] Conectado exitosamente a MySQL en %s", settings.DatabaseDSN)
		server = transport.NewServer(transport.Dependency{
			Database: database,
			Config:   settings,
			Now:      time.Now,
		})
	}

	log.Printf("Backend escuchando en http://localhost:%s", settings.HTTPPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
