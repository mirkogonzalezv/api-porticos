# AGENTE_API.md

Este documento define las reglas de trabajo para un agente enfocado en la API de pórticos TAG para Chile y la APP en Flutter
que estaremos trabajando.

## Objetivo del agente

Diseñar e implementar una API backend en Go para:

- Registrar pórticos.
- Exponer pórticos para consumo de la app móvil.
- Preparar base para detección de eventos de paso, tarifas y vehículos.

El enfoque debe ser incremental (MVP primero), con arquitectura limpia y mantenible.

## Stack técnico base

- Lenguaje: Go (estable actual del proyecto).
- API: HTTP REST con `gin`
- Configuración: variables de entorno.
- Persistencia:
  - Fase 1: repositorio en memoria o JSON seed.
  - Fase 2: Firestore con Firebase o Supabase.
- Observabilidad mínima: logs estructurados y endpoint de healthcheck.
- Enfocado en Zero Trust
- Enfocado en Security Software Developer
- Enfocado en OWASP top 10+
- Comportate como un Google Developer Expert

## Arquitectura objetivo

Separación por capas y responsabilidades:

1. `domain`

- Entidades puras (`Portico`, `Vehiculo`, `EventoPaso`, `Tarifa`).
- Reglas de negocio.
- Interfaces de repositorio.

2. `application` o `service`

- Casos de uso: `CreatePortico`, `ListPorticos`, `GetPorticoByID`, `UpdatePortico`, `DeletePortico`.
- Orquestación entre repositorios, validaciones y reglas de dominio.

3. `infrastructure`

- Implementación de repositorios (memoria/DB).
- Carga de seeds.
- Cliente externo si se integra API de tarifas.

4. `interfaces/http` o `transport/http`

- Handlers HTTP.
- Mapeo request/response JSON.
- Mapeo de errores de negocio a status HTTP.

Dependencias permitidas:

- `http` -> `application` -> `domain`
- `infrastructure` implementa interfaces de `domain`
- `domain` no depende de capas externas

## Modelo inicial recomendado

### Portico

- `id` (string, único)
- `latitude` (float64)
- `longitude` (float64)
- `bearing` (float64, 0-360)
- `detectionRadiusMeters` (float64, >0)

### Reglas mínimas de validación

- `id` obligatorio.
- `latitude` entre -90 y 90.
- `longitude` entre -180 y 180.
- `bearing` entre 0 y 360.
- `detectionRadiusMeters` mayor que 0.

## Endpoints MVP (v1)

- `GET /health`
- `GET /api/v1/porticos`
- `GET /api/v1/porticos/{id}`
- `POST /api/v1/porticos`
- `PUT /api/v1/porticos/{id}`
- `DELETE /api/v1/porticos/{id}`

### Contrato JSON base

```json
{
  "id": "p-001",
  "latitude": -33.45,
  "longitude": -70.6667,
  "bearing": 180.0,
  "detectionRadiusMeters": 100.0
}
```

## Convenciones de implementación

- Contexto en todas las operaciones (`context.Context`).
- Evitar lógica de negocio en handlers.
- Errores de dominio tipados y mapeo HTTP consistente:
  - validación -> `400`
  - no encontrado -> `404`
  - conflicto -> `409`
  - error interno -> `500`
- Respuestas JSON consistentes y versionadas (`/api/v1`).
- No exponer detalles internos de infraestructura al cliente.

## Calidad mínima obligatoria

- Tests unitarios en:
  - validaciones de dominio
  - casos de uso
  - handlers principales
- `go test ./...` en verde antes de cerrar cambios.
- Formato y limpieza:
  - `gofmt`
  - `go vet`

## Seguridad y operación (mínimo)

- Timeouts de servidor HTTP.
- Límite de tamaño de body en requests de escritura.
- Validación estricta de input JSON.
- CORS explícito según ambiente si aplica.
- Variables de entorno para puerto y DSN.
- Zero Trust
- OWASP top 10+
- Rate Limit

## Roadmap sugerido

1. CRUD de pórticos con repositorio en memoria + seed.
2. Persistencia PostgreSQL + migraciones.
3. Vehículos por usuario.
4. Registro de eventos de paso offline/online sync.
5. Integración de tarifas por horario y tipo de vehículo.

## Regla de colaboración con el equipo Flutter

- Mantener compatibilidad del contrato JSON con la app móvil.
- Versionar cambios breaking en `/api/v2`.
- Publicar ejemplos de request/response por endpoint.
- No agregar ni modificar archivos a menos que el desarrollador lo decida
- Siempre muestra la respuesta como deberia verse el archivo de forma correcta
- Valida que se respeten las reglas de seguridad
