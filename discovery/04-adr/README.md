# Architectural Decision Records (ADRs)

Este directorio documenta las decisiones técnicas y arquitectónicas más importantes adoptadas durante el diseño, refactorización y resolución de auditorías de seguridad del Sistema de Soporte Técnico Automotriz.

## Índice de Registros

| Código | Título | Estado | Fecha | Decisión Principal |
|---|---|---|---|---|
| [ADR-001](file:///c:/Users/esqui/OneDrive/Desktop/TRABAJOS%20EN%20CLASE%202026/Septiembre/19/Proyecto%20-%20Mejora/discovery/04-adr/ADR-001-clean-architecture-go.md) | Adopción de Clean Architecture en Go | Aceptado | 2026-09 | Estructura en 4 capas desacopladas (`domain`, `usecase`, `repository`, `transport/http`) con inversión de dependencias. |
| [ADR-002](file:///c:/Users/esqui/OneDrive/Desktop/TRABAJOS%20EN%20CLASE%202026/Septiembre/19/Proyecto%20-%20Mejora/discovery/04-adr/ADR-002-rbac-and-bola-protection.md) | Mitigación de OWASP A01: RBAC y Aislamiento de Órdenes (BOLA) | Aceptado | 2026-09 | Middleware de roles (`RequireRole`) y verificación de propiedad en Use Case para técnicos. |
| [ADR-003](file:///c:/Users/esqui/OneDrive/Desktop/TRABAJOS%20EN%20CLASE%202026/Septiembre/19/Proyecto%20-%20Mejora/discovery/04-adr/ADR-003-session-cookies-and-rate-limiting.md) | Autenticación Segura, Cookies HttpOnly y Rate Limiting | Aceptado | 2026-09 | Hasheo Bcrypt individualizado, cookies `HttpOnly; SameSite=Strict`, y limitador de tasa deslizante en memoria contra fuerza bruta. |
| [ADR-004](file:///c:/Users/esqui/OneDrive/Desktop/TRABAJOS%20EN%20CLASE%202026/Septiembre/19/Proyecto%20-%20Mejora/discovery/04-adr/ADR-004-strict-lifecycle-and-state-machine.md) | Máquina de Estados Finita e Invariantes del Ciclo de Servicio | Aceptado | 2026-09 | Transición estrictamente lineal de estados, validación previa de diagnósticos/intervenciones e inmutabilidad en estado entregado. |
