# ADR-003: Autenticación Segura, Cookies HttpOnly y Rate Limiting

## Estado
Aceptado

## Contexto
La auditoría de seguridad reveló vulnerabilidades en la capa de autenticación y manejo de sesiones:
1. Vulnerabilidad a ataques de fuerza bruta y DDoS en el endpoint de inicio de sesión (`POST /api/session`) por ausencia de limitador de tasa de peticiones.
2. Posibilidad de robo de tokens mediante Cross-Site Scripting (XSS) si los tokens de sesión se almacenaban en `localStorage`.
3. Contraseñas de prueba con hashes compartidos o débiles en las semillas de la base de datos.

## Decisión

### 1. Rate Limiter Deslizante en Memoria (`internal/transport/http/rate_limiter.go`)
- Implementación de un limitador en memoria thread-safe (`sync.Mutex`) basado en ventanas deslizantes por dirección IP de origen (`r.RemoteAddr` / `X-Forwarded-For`).
- Parámetros: **Máximo 5 solicitudes cada 15 minutos** para `POST /api/session`.
- En la 6ª solicitud dentro de la ventana, el middleware responde inmediatamente con `HTTP 429 Too Many Requests` y cuerpo JSON `{"error": "too_many_requests"}` con cabecera `Retry-After: 900`.

### 2. Cookies Seguras `HttpOnly` y Soporte Dual de Tokens
- Tras la autenticación exitosa en `POST /api/session`, el servidor emite una cookie:
  ```http
  Set-Cookie: session_token=<token>; Path=/; HttpOnly; SameSite=Strict; Secure; Max-Age=86400
  ```
- El backend acepta la sesión tanto desde la cookie `session_token` como desde la cabecera estándar `Authorization: Bearer <token>`, permitiendo compatibilidad plena con clientes móviles, APIs y navegadores modernos.
- Al cerrar sesión (`DELETE /api/session`), la cookie es revocada con `Max-Age=-1`.

### 3. Criptografía y Semillas Independientes
- Todas las contraseñas se almacenan procesadas mediante `bcrypt.GenerateFromPassword(..., bcrypt.DefaultCost)`.
- Se generaron hashes Bcrypt independientes y únicos para los usuarios semilla (`admin`, `jperez`, `lramirez`).

## Consecuencias

### Positivas
- Neutralización efectiva de ataques automatizados de diccionario y credential stuffing.
- Protección de sesiones contra exfiltración mediante scripts maliciosos (XSS) gracias al flag `HttpOnly`.
- Prevención de ataques Cross-Site Request Forgery (CSRF) mediante la política `SameSite=Strict`.
