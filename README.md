# Sistema de Gestión de Soporte Técnico Automotriz

Sistema web integral para la digitalización operativa de talleres automotrices: recepción de clientes y vehículos, diagnóstico técnico, asignación de mecánicos, seguimiento en tiempo real de órdenes de servicio, registro de intervenciones, gestión de garantías y línea de tiempo clínica por vehículo.

---

## 👥 Equipo de Desarrollo

- **Kevin Eduardo Arguello Solano**
- **Brenda Carolina Galeano**
- **Felipe Castro**
- **Jhorman Jamir**

---

## 🛠️ Arquitectura del Proyecto

El repositorio está organizado con una arquitectura desacoplada y limpia:

```
├── backend/            # API REST desarrollada en Go (Golang)
│   ├── cmd/server/     # Punto de entrada de la aplicación
│   ├── internal/       # Lógica de dominio, casos de uso, repositorios y transporte HTTP
│   └── Dockerfile      # Contenedor para despliegue del backend
├── database/           # Base de datos MySQL
│   ├── migrations/     # Scripts de migración DDL
│   ├── seed/           # Datos iniciales (administrador y datos base)
│   └── docker-compose.db.yml # Entorno local de base de datos
├── discovery/          # Documentación de negocio, arquitectura (ADRs) y gobernanza
├── docs/               # Documento de requerimientos de producto (PRD)
├── frontend/           # Aplicación Web SPA en React 18 + Vite + TypeScript
│   ├── src/            # Componentes, vistas y servicios
│   └── vercel.json     # Configuración de enrutamiento SPA para Vercel
└── vercel.json         # Configuración raíz para despliegue en Vercel
```

---

## 🚀 Despliegue en Vercel (Frontend)

El frontend está 100% preparado para ser desplegado en **[Vercel](https://vercel.com)**:

### Opción A (Recomendada: Importación Directa):
1. Ingresa a tu cuenta de **Vercel** y pulsa en **Add New... > Project**.
2. Selecciona el repositorio `Kevin81A/Proyecto-Automotriz`.
3. En la sección **Root Directory**, selecciona la carpeta `frontend`.
4. Vercel detectará automáticamente **Vite** como framework:
   - **Build Command**: `npm run build`
   - **Output Directory**: `dist`
   - **Install Command**: `npm install`
5. *(Opcional)* En **Environment Variables**, añade:
   - `VITE_API_BASE_URL`: URL pública de tu backend desplegado (ejemplo: `https://tu-backend.up.railway.app/api`).
6. Haz clic en **Deploy**.

### Opción B (Monorepo desde la raíz):
- Si no seleccionas `frontend` como directorio raíz, el archivo `vercel.json` en la raíz se encargará de compilar automáticamente la carpeta `frontend/` y servir `frontend/dist`.

---

## 💻 Ejecución en Entorno Local

### 1. Base de Datos (MySQL)
```bash
cd database
docker compose -f docker-compose.db.yml up -d
```

### 2. Backend (Go)
```bash
cd backend
cp .env.example .env
# Ajustar credenciales en .env si es necesario
go run ./cmd/server
```
El servidor iniciará en `http://localhost:8080`.

Para ejecutar las pruebas:
```bash
go test ./...
```

### 3. Frontend (React + Vite)
```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```
La aplicación web estará disponible en `http://localhost:4173`.
