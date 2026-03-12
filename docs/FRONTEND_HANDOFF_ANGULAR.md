# API Pórticos - Guía para Frontend Angular

## 1) Objetivo de negocio
Construir una plataforma donde cada usuario pueda:
- Registrar sus vehículos.
- Consultar pórticos/peajes y sus tarifas.
- Registrar pasos por pórtico.
- Consultar histórico y resumen de gasto por periodos.

Meta del MVP:
- Entregar valor operativo real para usuarios finales.
- Controlar consumo de API y uso indebido.
- Dejar base lista para monetización futura (B2B/B2C).

## 2) Principios del producto
- Seguridad por defecto (Zero Trust): nunca confiar en datos del cliente.
- Multi-tenant por usuario: cada usuario solo ve sus propios recursos sensibles (vehículos y pasos).
- Trazabilidad: cada paso por pórtico queda registrado como evento histórico.
- Escalabilidad incremental: hoy PostgreSQL local/cloud; más adelante caché distribuido y despliegue cloud completo.

## 3) Arquitectura funcional (visión frontend)
Dominios disponibles:
- `accounts`: creación de cuenta pública + administración de cuentas por rol.
- `porticos`: catálogo de pórticos y tarifas.
- `vehiculos`: CRUD de vehículos del usuario autenticado.
- `pasos`: registro y consulta histórica de pasos por pórtico.

Base URL local:
- `http://localhost:3200/api/v1`

## 4) Autenticación y autorización
Proveedor de auth:
- Supabase Auth (JWT firmado ES256).

Regla frontend:
- Login y refresh token se hace con Supabase.
- Angular envía `Authorization: Bearer <access_token>` en cada request protegida.

Roles actuales:
- `admin`: gestión avanzada (recursos administrativos).
- `partner`: consumo de API para integraciones.
- `reader`: consumo base de lectura y operaciones permitidas de usuario.

Importante:
- El backend resuelve rol desde DB (source of truth), no solo desde claims del token.

## 5) Alcance funcional del MVP para Angular
### 5.1 Cuentas
- Registro público de usuario final.
- Inicio de sesión (vía Supabase).
- Sesión persistente con refresh seguro.

### 5.2 Pórticos
- Listado de pórticos.
- Detalle de pórtico (ubicación, tarifas por tipo de vehículo, horarios).

### 5.3 Vehículos (owner-scoped)
- Crear vehículo.
- Listar mis vehículos.
- Editar mis vehículos.
- Eliminar mis vehículos.

### 5.4 Pasos (owner-scoped)
- Registrar paso por pórtico.
- Ver historial por rango de fechas.
- Ver resumen por `day | week | month`.

## 6) Contratos API que frontend debe respetar
### Headers
- `Authorization: Bearer <token>` (excepto endpoints públicos).
- `Content-Type: application/json` en POST/PUT.

### Fechas
- En requests usar `RFC3339` (`2026-03-05T14:30:00Z`) o formato `YYYY-MM-DD` cuando aplique query range.

### Paginación
- Query params: `limit`, `offset`.
- Respuesta incluye `data`, `limit`, `offset` cuando corresponde.

### Errores
Formato estándar:
```json
{
  "error": {
    "type": "VALIDATION_ERROR|UNAUTHORIZED|FORBIDDEN|NOT_FOUND|CONFLICT|INTERNAL_ERROR",
    "code": "ERROR_CODE",
    "message": "mensaje",
    "details": "opcional"
  },
  "timestamp": "2026-03-05T12:00:00Z"
}
```

## 7) Reglas de UX obligatorias
- Mostrar mensajes de error del backend (`error.message`) sin exponer datos técnicos.
- Manejar `401` con flujo de renovación/cierre de sesión.
- Manejar `403` mostrando “no autorizado”.
- Manejar `429` (rate limit) con retry progresivo y feedback al usuario.
- Formularios con validación previa (email, patente, rangos de fecha).

## 8) Flujos de pantallas recomendados
1. Auth:
- Sign up
- Login
- Recuperar sesión

2. Home:
- Resumen rápido: cantidad de pasos del periodo + gasto acumulado.

3. Vehículos:
- Lista de mis vehículos
- Crear/editar vehículo

4. Pórticos:
- Lista + detalle de pórtico (tarifas)

5. Pasos:
- Registrar paso
- Historial (filtros por fecha, vehículo, pórtico)
- Resumen (agrupación día/semana/mes)

## 9) Reglas de colaboración Backend-Frontend
- No inventar campos ni rutas: usar contratos acordados.
- Cualquier cambio de contrato debe versionarse y avisarse antes.
- Mantener un changelog compartido por sprint (breaking/non-breaking).
- Usar entorno `.env` por ambiente (`dev`, `staging`, `prod`) en Angular.

## 10) KPIs del MVP
- Registro de usuario exitoso.
- Registro y consulta de vehículos sin fuga entre usuarios.
- Registro de paso exitoso y visible en historial.
- Resumen por periodo consistente con histórico.
- Tasa de error controlada (4xx esperado, 5xx mínimo).

## 11) Fuera de alcance MVP (por ahora)
- Facturación/pagos.
- Panel analítico avanzado.
- Notificaciones push complejas.
- Integraciones masivas en tiempo real.

## 12) Checklist de entrega Frontend
- [ ] Flujo auth completo integrado con Supabase.
- [ ] Interceptor HTTP con Bearer token y manejo 401/403/429.
- [ ] Módulo Vehículos completo.
- [ ] Módulo Pórticos lectura completa.
- [ ] Módulo Pasos (crear + historial + resumen).
- [ ] Manejo unificado de errores.
- [ ] Pruebas E2E de rutas críticas del negocio.

---
Si este documento cambia, actualizar versión y fecha en el encabezado del PR.
