$env:MYSQL_HOST="127.0.0.1"
$env:MYSQL_PORT="3307"
$env:MYSQL_DATABASE="workshop"
$env:MYSQL_USER="workshop_app"
$env:MYSQL_PASSWORD="replace_with_secure_password"
$env:TOKEN_SECRET="super_secret_jwt_token_2026_production_key"
$env:ALLOWED_ORIGIN="http://localhost:4173"
$env:HTTP_PORT="8080"

Write-Host "Iniciando Backend Go en http://localhost:8080..." -ForegroundColor Cyan
go run ./cmd/server
