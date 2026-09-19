# ADR-004: Máquina de Estados Finita e Invariantes del Ciclo de Servicio

## Estado
Aceptado

## Contexto
En el flujo operativo previo existían inconsistencias graves de lógica de negocio detectadas en auditoría:
1. **Saltos de estado arbitrarios:** Se permitía transicionar una orden directamente de `RECEIVED` a `READY` o `DELIVERED` sin pasar por diagnóstico ni reparación.
2. **Falta de prerrequisitos técnicos:** Se podía marcar una orden como `IN_REPAIR` sin haber diagnosticado el vehículo, y marcarla como `READY` sin haber registrado ninguna intervención física.
3. **Modificación de órdenes cerradas:** Una vez entregado el vehículo (`DELIVERED`), era posible seguir agregando diagnósticos, intervenciones o alterar su estado.
4. **Colisión de órdenes vehiculares:** Un mismo vehículo podía registrar múltiples órdenes abiertas simultáneamente en el taller, distorsionando el historial clínico y la asignación de responsabilidades.

## Decisión

Modelar el ciclo de vida de la orden de servicio como una **Máquina de Estados Finita (FSM)** estricta con validación de invariantes en el núcleo de dominio:

### 1. Transición Estrictamente Secuencial
La secuencia permitida es unidireccional y sin omisiones:
$$\text{RECEIVED} \longrightarrow \text{IN\_DIAGNOSIS} \longrightarrow \text{IN\_REPAIR} \longrightarrow \text{READY} \longrightarrow \text{DELIVERED}$$
Cualquier intento de saltar estados (ej. `RECEIVED -> READY`) o retroceder (ej. `IN_REPAIR -> RECEIVED`) es rechazado con `domain.ErrInvalidTransition` (`HTTP 400 Bad Request`).

### 2. Validación de Prerrequisitos en Casos de Uso
- Para avanzar a `IN_REPAIR`: el caso de uso consulta `diagnosticRepo.GetByOrderID(orderID)`. Si el arreglo está vacío, la operación se cancela con error `domain.ErrDiagnosticRequired`.
- Para avanzar a `READY`: el caso de uso consulta `interventionRepo.GetByOrderID(orderID)`. Si no hay intervenciones registradas, se cancela con `domain.ErrInterventionRequired`.

### 3. Inmutabilidad en Estado Final (`DELIVERED`)
- El estado `DELIVERED` es un estado terminal.
- Los casos de uso de registro de diagnóstico (`CreateDiagnostic`) e intervención (`CreateIntervention`) consultan el estado actual de la orden; si es `DELIVERED`, retornan `domain.ErrOrderDelivered` / `HTTP 400 Bad Request`.

### 4. Unicidad de Orden Activa por Vehículo
- Al intentar crear una orden (`CreateServiceOrder`), el repositorio vehicular verifica si existe una orden con `status != DELIVERED` para el mismo `VehicleID`. De existir, se aborta la transacción con `domain.ErrConflict` (`HTTP 409 Conflict`).

## Consecuencias

### Positivas
- **Integridad absoluta del proceso operativo:** Ningún vehículo puede ser despachado sin diagnóstico documentado ni reparaciones justificadas.
- **Historial clínico inalterable:** Se preserva la verdad jurídica y técnica de cada servicio realizado en el taller.
- **Eliminación de colisiones de inventario y facturación.**
