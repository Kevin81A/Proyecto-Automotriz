# ADR-001: Adopción de Clean Architecture en Backend Go

## Estado
Aceptado

## Contexto
El sistema de soporte técnico automotriz requiere alta confiabilidad, trazabilidad estricta y aislamiento de la lógica de negocio respecto a los mecanismos de entrega (HTTP REST) y persistencia (MySQL / SQLite). Para garantizar mantenibilidad, testabilidad unitaria sin dependencias externas y cumplimiento con principios SOLID, se requería una arquitectura limpia y desacoplada.

## Decisión
Implementar **Clean Architecture** (Arquitectura Limpia) en el backend desarrollado en Go 1.22+, organizando el código en las siguientes capas concéntricas con regla de dependencia unidireccional hacia el centro:

1. **`internal/domain` (Capa Central / Núcleo):**
   - Contiene entidades puras (`ServiceOrder`, `Vehicle`, `Customer`, `Technician`, `Diagnostic`, `Intervention`, `Warranty`), objetos de valor y definiciones de errores (`ErrInvalidTransition`, `ErrUnauthorized`, `ErrConflict`).
   - Cero dependencias externas (solo librería estándar de Go).
   
2. **`internal/usecase` (Capa de Casos de Uso / Aplicación):**
   - Orquesta la lógica del negocio: `ServiceOrderUseCase`, `DiagnosticUseCase`, `SessionUseCase`, etc.
   - Define interfaces de repositorios para inversión de dependencias (`ServiceOrderRepository`, etc.).
   - Aplica validaciones de permisos de negocio (ej. validación de técnico asignado a la orden).

3. **`internal/repository` (Capa de Adaptadores de Persistencia):**
   - Implementa las interfaces de repositorio mediante SQL directo con consultas preparadas contra MySQL / SQLite.

4. **`internal/transport/http` (Capa de Controladores y Transporte):**
   - Handlers HTTP estándar, enrutamiento, parsing de JSON, serialización de respuestas y middlewares de seguridad (RBAC, Rate Limiting, Sanitización).

## Consecuencias

### Positivas
- **Alta testabilidad:** La lógica de negocio y casos de uso se prueban de forma unitaria instantánea mediante mocks de repositorio sin requerir levantar base de datos real.
- **Independencia de la infraestructura:** Cambiar de base de datos o motor SQL no afecta el dominio ni los casos de uso.
- **Seguridad en capas:** Las políticas de acceso se verifican tanto en el transporte (roles) como en los casos de uso (propiedad del recurso).

### Negativas / Mitigaciones
- Mayor número de archivos y boilerplate de interfaces; mitigado por la simplicidad y claridad de las interfaces en Go (`small, focused interfaces`).
