@echo off
set MYSQL_HOST=127.0.0.1
set MYSQL_PORT=3307
set MYSQL_DATABASE=workshop
set MYSQL_USER=workshop_app
set MYSQL_PASSWORD=replace_with_secure_password
set TOKEN_SECRET=super_secret_jwt_token_2026_production_key
set ALLOWED_ORIGIN=http://localhost:4173
set HTTP_PORT=8080

echo Iniciando Backend Go en http://localhost:8080...
go run ./cmd/server
