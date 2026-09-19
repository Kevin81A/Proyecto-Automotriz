# Entities, Value Objects, and Business Rules

> **Dominio del Negocio:** Sistema de Soporte Técnico y Taller Automotriz.
> Este documento traduce las reglas del negocio, el modelo táctico de DDD y los requerimientos de auditoría y seguridad al modelo de dominio del sistema.

---

## 1. Conceptos Tácticos de DDD

### Entidad (Entity)
Objeto con identidad única e inmutable en el tiempo (`ID` UUID o numérico). Dos órdenes son distintas aunque compartan vehículo o técnico.
- `ServiceOrder`, `Vehicle`, `Customer`, `Technician`, `Diagnostic`, `Intervention`, `Warranty`.

### Objeto de Valor (Value Object - VO)
Objeto definido enteramente por sus atributos, sin identidad propia e inmutable. Si un atributo cambia, se instancia un nuevo VO.
- `LicensePlate`, `VIN`, `Money`, `Email`, `PhoneNumber`, `OrderStatus`, `Role`.

### Agregado (Aggregate)
Conjunto de entidades y VOs tratados como una unidad transaccional consistente. El acceso se realiza exclusivamente a través de la **Raíz del Agregado (Aggregate Root)**.
- **Raíz del Agregado:** `ServiceOrder`
  - `Diagnostic` (Entidad interna)
  - `Intervention` (Entidad interna)
  - `PartUsage` (Objeto de valor / entidad interna de intervención)
  - `OrderStatus` (Value Object de estado del ciclo de vida)

---

## 2. Entidades Principales del Sistema

### 2.1 Entidad Raíz: `ServiceOrder` (Orden de Servicio)

**Contexto:** Core Workshop / Soporte Técnico.
**Descripción:** Orquesta y preserva el ciclo de vida completo de un vehículo ingresado al taller, desde su recepción hasta la entrega al cliente.

**Atributos:**
| Atributo | Tipo | Descripción | Requerido | Reglas de Validación |
|---|---|---|---|---|
| `id` | string (UUID) | Identificador único | Sí | Inmutable, generado al crear |
| `vehicleId` | string (UUID) | Automotor asociado | Sí | Debe existir en catálogo vehicular |
| `customerId` | string (UUID) | Cliente propietario/solicitante | Sí | Debe existir en clientes |
| `assignedTechnicianId`| *string (UUID)| Técnico responsable asignado | No | Solo modificable por Administrador |
| `status` | OrderStatus | Estado actual del ciclo | Sí | `RECEIVED`, `IN_DIAGNOSIS`, `IN_REPAIR`, `READY`, `DELIVERED` |
| `priority` | string | Prioridad (`LOW`, `MEDIUM`, `HIGH`, `URGENT`) | Sí | Valor por defecto `MEDIUM` |
| `mileage` | int | Kilometraje al ingresar | Sí | Entero positivo > 0 |
| `customerReportedFault`| string | Falla reportada por el cliente | Sí | Texto saneado anti-XSS |
| `receivedAt` | time.Time | Fecha y hora de recepción | Sí | Timestamp de creación |
| `deliveredAt` | *time.Time | Fecha y hora de entrega | No | Asignado al transicionar a `DELIVERED` |

**Máquina de Estados y Ciclo de Vida:**
```
[RECEIVED] ──(Diagnosticar)──▶ [IN_DIAGNOSIS] ──(Aprobar Diagnóstico)──▶ [IN_REPAIR]
                                                                              │
                                                                       (Completar Reparación)
                                                                              ▼
[DELIVERED] ◀──────(Entregar al Cliente)──────── [READY]
```

**Transiciones Permitidas:**
| Estado Origen | Estado Destino | Condición Previa Requerida | Actor Permitido |
|---|---|---|---|
| *(Nuevo)* | `RECEIVED` | Vehículo sin otra orden activa | Administrador / Recepción |
| `RECEIVED` | `IN_DIAGNOSIS` | Técnico asignado | Administrador / Técnico asignado |
| `IN_DIAGNOSIS`| `IN_REPAIR` | Al menos 1 `Diagnostic` registrado | Administrador / Técnico asignado |
| `IN_REPAIR` | `READY` | Al menos 1 `Intervention` registrada | Administrador / Técnico asignado |
| `READY` | `DELIVERED` | Inspección de calidad conforme | Administrador / Técnico asignado |
| `DELIVERED` | *(Terminal)* | Inmutable; no permite cambios | Ninguno |

---

### 2.2 Entidad: `Diagnostic` (Diagnóstico Técnico)

**Descripción:** Evaluación técnica minuciosa realizada por un técnico donde se establece la causa raíz de la falla.

**Atributos:**
| Atributo | Tipo | Descripción | Reglas |
|---|---|---|---|
| `id` | string (UUID) | Identificador del diagnóstico | Auto-generado |
| `orderId` | string (UUID) | Orden de servicio vinculada | La orden no debe estar en `DELIVERED` |
| `technicianId` | string (UUID) | Técnico evaluador | Debe coincidir con el asignado (o Admin) |
| `rootCause` | string | Causa técnica de la avería | Saneado contra XSS |
| `severity` | string | `LOW`, `MEDIUM`, `HIGH`, `CRITICAL` | Requerido |
| `createdAt` | time.Time | Timestamp de registro | Monotónico (>= `receivedAt`) |

---

### 2.3 Entidad: `Intervention` (Intervención Técnica / Reparación)

**Descripción:** Registro físico de la labor correctiva o preventiva ejecutada sobre el vehículo y repuestos reemplazados.

**Atributos:**
| Atributo | Tipo | Descripción | Reglas |
|---|---|---|---|
| `id` | string (UUID) | Identificador de intervención | Auto-generado |
| `orderId` | string (UUID) | Orden vinculada | Orden en estado `IN_REPAIR` (no `DELIVERED`) |
| `technicianId` | string (UUID) | Técnico ejecutor | Debe coincidir con el asignado |
| `description` | string | Detalle del trabajo efectuado | Requerido, anti-XSS |
| `laborHours` | float64 | Horas invertidas de mano de obra | > 0 |
| `partsReplaced` | []PartUsage | Repuestos instalados | Opcional |
| `createdAt` | time.Time | Timestamp de intervención | Monotónico (>= `diagnostic.createdAt`) |

---

### 2.4 Entidad: `Warranty` (Garantía)

**Descripción:** Certificado de garantía comercial y técnica otorgado al cliente sobre repuestos o mano de obra.

**Atributos:**
| Atributo | Tipo | Descripción | Reglas |
|---|---|---|---|
| `id` | string (UUID) | Identificador único | Auto-generado |
| `orderId` | string (UUID) | Orden de origen | Requerido |
| `vehicleId` | string (UUID) | Vehículo garantizado | Requerido |
| `coverageMonths` | int | Meses de cobertura (ej: 6, 12) | >= 1 |
| `mileageLimit` | int | Límite de kilometraje cubierto | > kilometraje al momento del servicio |
| `status` | string | `ACTIVE`, `CLAIMED`, `EXPIRED` | Por defecto `ACTIVE` |

---

## 3. Invariantes del Negocio (Invariants)

```text
INV-001: Strict Sequential Lifecycle
  - Regla: Las transiciones de estado son estrictamente lineales: RECEIVED -> IN_DIAGNOSIS -> IN_REPAIR -> READY -> DELIVERED.
  - Violación: ErrInvalidTransition si se intenta saltar estados o retroceder.

INV-002: Preconditions on Advance
  - Regla: No se puede avanzar a IN_REPAIR sin al menos un diagnóstico registrado.
  - Regla: No se puede avanzar a READY sin al menos una intervención técnica registrada.
  - Violación: ErrDiagnosticRequired o ErrInterventionRequired.

INV-003: Immutability of Delivered Orders
  - Regla: Una vez entregada (DELIVERED), una orden queda cerrada y sellada permanentemente.
  - Violación: Se rechaza cualquier intento de agregar diagnósticos, intervenciones o cambiar estado con ErrInvalidTransition.

INV-004: Single Active Order per Vehicle
  - Regla: Un vehículo no puede tener dos órdenes de servicio activas simultáneamente en el taller.
  - Violación: Retorna HTTP 409 Conflict (ErrConflict) al intentar abrir una nueva orden si existe una abierta.

INV-005: Technician Order Ownership (BOLA Protection)
  - Regla: Un técnico solo puede visualizar, diagnosticar, intervenir y avanzar órdenes asignadas a su propio ID.
  - Violación: Retorna HTTP 403 Forbidden (ErrUnauthorized) ante cualquier intento de acceso o modificación a órdenes ajenas o sin asignar.

INV-006: Monotonic Chronology of Audit Timestamps
  - Regla: Los timestamps deben guardar orden secuencial: receivedAt <= diagnosticAt <= interventionAt <= deliveredAt.
  - Violación: Registro rechazado por inconsistencia cronológica.

INV-007: Input Sanitization and Safe Types
  - Regla: Todo texto ingresado por el usuario es saneado de etiquetas HTML/scripts ejecutables. Las placas vehiculares son validadas sintácticamente contra patrones de inyección SQL.
  - Violación: HTTP 400 Bad Request si contiene caracteres maliciosos.
```

---

## 4. Implementación en Go (Clean Architecture)

```go
package domain

import (
    "errors"
    "time"
)

var (
    ErrInvalidTransition   = errors.New("invalid status transition")
    ErrDiagnosticRequired  = errors.New("cannot advance to IN_REPAIR without diagnostic")
    ErrInterventionRequired = errors.New("cannot advance to READY without intervention")
    ErrOrderDelivered      = errors.New("cannot modify delivered order")
    ErrUnauthorized        = errors.New("access denied to this service order")
    ErrConflict            = errors.New("active order already exists for this vehicle")
)

type OrderStatus string

const (
    StatusReceived    OrderStatus = "RECEIVED"
    StatusInDiagnosis OrderStatus = "IN_DIAGNOSIS"
    StatusInRepair    OrderStatus = "IN_REPAIR"
    StatusReady       OrderStatus = "READY"
    StatusDelivered   OrderStatus = "DELIVERED"
)

func (o *ServiceOrder) CanAdvanceTo(next OrderStatus) error {
    if o.Status == StatusDelivered {
        return ErrOrderDelivered
    }
    switch o.Status {
    case StatusReceived:
        if next != StatusInDiagnosis {
            return ErrInvalidTransition
        }
    case StatusInDiagnosis:
        if next != StatusInRepair {
            return ErrInvalidTransition
        }
    case StatusInRepair:
        if next != StatusReady {
            return ErrInvalidTransition
        }
    case StatusReady:
        if next != StatusDelivered {
            return ErrInvalidTransition
        }
    default:
        return ErrInvalidTransition
    }
    return nil
}
```

---

## 5. Mapeo con el Código Fuente

| Artefacto del Dominio | Capa en Clean Architecture | Archivo en Repositorio |
|---|---|---|
| Entidad `ServiceOrder` | `internal/domain` | `backend/internal/domain/service_order.go` |
| Entidades `Vehicle`, `Customer`, `Technician` | `internal/domain` | `backend/internal/domain/models.go` |
| Errores e Invariantes del Dominio | `internal/domain` | `backend/internal/domain/errors.go` |
| Caso de Uso: Avanzar Estado | `internal/usecase` | `backend/internal/usecase/service_order_usecase.go` |
| Caso de Uso: Registrar Diagnóstico | `internal/usecase` | `backend/internal/usecase/diagnostic_usecase.go` |
| Control de Acceso y Middleware RBAC | `internal/transport/http` | `backend/internal/transport/http/middleware.go` |
| Controladores y Endpoints HTTP | `internal/transport/http` | `backend/internal/transport/http/service_order_handler.go` |
