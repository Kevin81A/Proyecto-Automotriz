# ADR-002: Mitigación de OWASP A01: Control de Acceso Basado en Roles (RBAC) y Aislamiento de Órdenes (BOLA)

## Estado
Aceptado

## Contexto
Durante la auditoría de seguridad (`auditoria_software.md`), se identificaron fallas críticas de Broken Access Control (OWASP Top 10 - A01):
1. **Broken Object Level Authorization (BOLA):** Los técnicos podían consultar, diagnosticar, intervenir y avanzar de estado órdenes de servicio que no les pertenecían (órdenes sin asignar o asignadas a otros técnicos).
2. **Falta de RBAC en Catálogos:** Usuarios no autenticados o con rol técnico podían listar información confidencial de clientes (`GET /api/customer`), vehículos (`GET /api/vehicle`), técnicos (`GET /api/technician`) y garantías (`GET /api/warranty`).
3. **Fugas en Frontend:** El panel de asignación de mecánicos y las órdenes completas de todos los técnicos eran renderizadas en el cliente web sin discriminación de rol.

## Decisión

Implementar un esquema de defensa en profundidad en tres niveles:

### 1. Middleware HTTP de RBAC (`internal/transport/http/middleware.go`)
- Funciones auxiliares `RequireRole(allowedRoles ...domain.Role)` aplicadas en el enrutador.
- Los endpoints administrativos (`/api/customer`, `/api/vehicle`, `/api/technician`, `/api/warranty`, `/api/service-order/assign`) exigen explícitamente el rol `ADMINISTRATOR`. Si un usuario con rol `TECHNICIAN` o anónimo intenta acceder, el middleware aborta de inmediato con `HTTP 403 Forbidden` (`{"error": "forbidden"}`).

### 2. Aislamiento de Órdenes a Nivel de Caso de Uso (`internal/usecase/service_order_usecase.go`)
- En `GetByID`:
  - Si el usuario es `ADMINISTRATOR`: acceso total concedido.
  - Si el usuario es `TECHNICIAN`: se valida que `order.AssignedTechnicianID != nil && *order.AssignedTechnicianID == user.ID`. Si no está asignada o pertenece a otro técnico, retorna `domain.ErrUnauthorized` -> mapeado a `HTTP 403 Forbidden`.
- En `AdvanceStatus`:
  - Solo el administrador o el técnico específicamente asignado pueden avanzar de estado la orden.
- En `List`:
  - Para `ADMINISTRATOR`: lista todas las órdenes del taller.
  - Para `TECHNICIAN`: filtra automáticamente para retornar únicamente las órdenes asignadas a su ID. Las órdenes sin asignar NO son visibles para el técnico.

### 3. Blindaje de Frontend en React (`frontend/src/App.tsx`)
- Componente `<ProtectedRoute requiredRole="ADMINISTRATOR">` que redirige a dashboard o muestra mensaje de no autorizado.
- Renderizado condicional del selector de asignación de técnico en la vista de detalle de orden: oculto para técnicos, visible solo para administradores.

## Consecuencias

### Positivas
- **Cierre total de la brecha BOLA (OWASP A01):** Un técnico no puede leer ni manipular datos de órdenes ajenas vía UI ni mediante llamadas directas a la API REST.
- **Principio de menor privilegio:** La información de clientes, garantías y empleados queda restringida exclusivamente al personal administrativo.
