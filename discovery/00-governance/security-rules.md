# Technical Security Rules

> Mandatory technical controls that apply to all project code.
> These rules complement the security policy (`security-policy.md`) with
> concrete implementation practices.

---

## OWASP Top 10 — Controls per category

### A01 — Broken Access Control

```go
// ❌ BAD — trusting client role or skipping permission checks
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    orderID := chi.URLParam(r, "id")
    order, _ := h.usecase.GetByID(orderID)
    json.NewEncoder(w).Encode(order) // Exposes any order to any caller!
}

// ✅ GOOD — extract user from validated context and enforce RBAC & ownership
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    user := middleware.UserFromContext(r.Context())
    orderID := chi.URLParam(r, "id")
    order, err := h.usecase.GetByID(r.Context(), user, orderID)
    if errors.Is(err, domain.ErrUnauthorized) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    json.NewEncoder(w).Encode(order)
}
```

**Rules:**
- Every protected endpoint MUST have the authentication & authorization middleware applied (`RequireRole(domain.RoleAdmin, domain.RoleTechnician)`).
- Endpoints returning customers, vehicles, technicians, and warranties require `RoleAdmin`.
- Technicians can ONLY view, diagnose, or advance orders explicitly assigned to their `TechnicianID` (Object-Level Authorization / BOLA protection).
- Unassigned orders are strictly visible to Administrators.

### A02 — Cryptographic Failures

**Rules:**
- Passwords: use **bcrypt** with `bcrypt.DefaultCost` (≥ 10-12). Independent salts and unique hashes for every user.
- Session Tokens / Cookies: session cookies emitted with `HttpOnly; Secure; SameSite=Strict`.
- Support both `Authorization: Bearer <token>` and cookie-based authentication.
- Sensitive data in transit: HTTPS mandatory in all environments outside local development.
- Zero credential logging: passwords and session tokens are strictly excluded from structured logs.

### A03 — Injection

**SQL & XSS Sanitization:**
```go
// ❌ BAD — direct string concatenation or unescaped HTML
query := fmt.Sprintf("SELECT * FROM vehicles WHERE plate = '%s'", plate)

// ✅ GOOD — parameterized queries and input sanitization
const query = "SELECT id, plate, vin FROM vehicles WHERE plate = ?"
err := db.QueryRowContext(ctx, query, plate).Scan(&id, &plate, &vin)

// ✅ Anti-XSS & Anti-SQLi validation
if containsDangerousSQL(plate) {
    return domain.ErrInvalidInput
}
sanitizedNotes := sanitizeHTML(notes)
```

**Rules:**
- Parameterized queries ALWAYS via `database/sql` placeholders (`?`). Zero concatenated strings.
- Validate and sanitize all user input before saving: strip dangerous HTML/script tags (`<script>`, `<iframe>`).
- Reject obvious SQL injection payloads in identifiers, plates, and names with HTTP 400 Bad Request.

### A04 — Insecure Design

- Every HU that exposes user data must undergo privacy review
- Bulk query endpoints have mandatory pagination (maximum [100] records per page)
- Do not expose sequential internal IDs; use UUIDs

### A05 — Security Misconfiguration

```
# Verification checklist per environment
□ Stack traces NOT visible in production
□ Security headers configured (Helmet.js or equivalent):
  - X-Content-Type-Options: nosniff
  - X-Frame-Options: DENY
  - Content-Security-Policy defined
  - Strict-Transport-Security in production
□ Unnecessary ports closed
□ Development credentials NOT in production
```

### A06 — Vulnerable Components

**Rules:**
- Run `npm audit` (or equivalent) before each release
- **Critical/High** vulnerabilities block the deploy
- Renew dependencies each sprint (at least once)
- Do not use `latest` versions without pinning in `package.json`; use exact versions or conservative ranges

### A07 — Identification and Authentication Failures

- Session token / JWT with expiration for authenticated sessions.
- Strict Rate Limiting on `POST /api/session`: maximum 5 failed/login attempts per IP address within a 15-minute window.
- The 6th request triggers an immediate HTTP 429 Too Many Requests (`{"error": "too_many_requests"}`).
- Cryptographic verification with unique per-user Bcrypt salt hashes.
- Session tokens transmitted via `HttpOnly; SameSite=Strict` cookies and validated on each request.

### A08 — Software and Data Integrity Failures

- Verify Docker image checksum before using in production
- Third-party webhooks must verify cryptographic signature
- Validate that messages from the broker (Kafka/RabbitMQ) have the expected schema

### A09 — Security Logging and Monitoring Failures

- Every failed authentication must be logged with IP, timestamp, and user-agent
- Log delete operations with who, when, and what was deleted
- Security logs are retained for a minimum of **90 days**
- Automatic alerts configured for:
  - More than [50] 401/403 errors in 5 minutes
  - Access to a resource from an unexpected country (if applicable)

### A10 — Server-Side Request Forgery (SSRF)

- URLs constructed from user input MUST be validated against an allowlist of permitted domains
- Do not fetch from private IPs (192.168.x.x, 10.x.x.x, 127.x.x.x) from the server

---

## User input handling

```typescript
// Example with Zod — always validate in the Controller/Adapter layer
const CreateUserSchema = z.object({
  email: z.string().email().max(255),
  name: z.string().min(1).max(100).trim(),
  role: z.enum(['ADMIN', 'USER', 'VIEWER']),
});

// The result is typed and sanitized
const parsed = CreateUserSchema.parse(req.body);
```

**Rule:** All external inputs (HTTP body, query params, path params, broker messages)
pass through a validation schema before reaching the domain.

---

## Secure error handling

```typescript
// ❌ BAD — exposes internal details
res.status(500).json({ error: error.message, stack: error.stack });

// ✅ GOOD — generic message + traceId for internal correlation
res.status(500).json({
  error: 'INTERNAL_SERVER_ERROR',
  message: 'Internal server error',
  traceId: req.headers['x-trace-id'],
});
```

---

## Correlations

- Security policy (management, access, vault) → `00-governance/security-policy.md`
- System threat model → `05-architecture/security-threat-model.md`
- Authentication and JWT → `07-api/authentication.md`
- Observability and security logs → `13-operations/observability.md`
