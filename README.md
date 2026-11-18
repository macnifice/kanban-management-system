# Kanban

React (Vite) + Go + MySQL, todo en Docker.

```bash
cp .env.example .env      # cambia JWT_SECRET
docker compose up --build
```

- Frontend: http://127.0.0.1:5173 (si `localhost` apunta a otro proceso en IPv6, usa 127.0.0.1 o cambia `FRONTEND_PORT`)
- API: http://127.0.0.1:8080/api/health
- MySQL: localhost:3306

Crea una cuenta desde la pantalla de acceso (el registro es abierto; los tableros son compartidos por todos los usuarios).

## Funcionalidad
- **Tableros**: crear, listar, editar y eliminar.
- **Columnas**: nombre, orden (drag & drop), **límite de WIP** y tipo (normal / en curso / terminado) desde el botón ⚙.
- **Tarjetas**: título, descripción, prioridad, fecha de creación, asignado y etiquetas; se mueven entre columnas.
- **Regla de WIP**: el servidor responde `409 wip_limit` si la columna destino está llena (también al crear). La UI además lo bloquea al arrastrar.
- **Historial** de movimientos por tarjeta (`card_events`), base del cycle time.
- **Dashboard** (`#/boards/:id/metrics`): tarjetas por columna, WIP actual vs. límite, totales por estado y cycle time (primera entrada a una columna «en curso» → entrada a una «terminada»).

## Arquitectura del backend (DDD)

```
backend/internal/
├── domain/                  Reglas de negocio puras (sin SQL ni HTTP)
│   ├── shared/              Errores y validaciones comunes
│   ├── identity/            Contexto de usuarios: User, puertos (repositorio, hasher, tokens)
│   └── kanban/              Contexto principal
│       ├── board.go, column.go, card.go, label.go   Agregados (raíces) con sus invariantes
│       ├── values.go        Objetos de valor: Priority, Kind, WIPLimit
│       ├── card.go          Servicios de dominio CreateCard / MoveCard (regla de WIP) y evento Movement
│       ├── cycletime.go     Cálculo del cycle time (lógica pura)
│       └── repositories.go  Puertos: repositorios y UnitOfWork
├── application/             Casos de uso (orquestan agregados dentro de una transacción)
│   ├── identity/            Registro, login, perfil
│   └── kanban/              BoardService, CardService, LabelService (comandos) y QueryService + ReadModel (consultas)
├── infrastructure/          Adaptadores
│   ├── mysql/               Repositorios, UnitOfWork, ReadModel, migraciones
│   └── security/            bcrypt y JWT
└── interfaces/httpapi/      Capa HTTP: rutas, DTOs y traducción de errores de dominio a códigos HTTP
cmd/api/main.go              Raíz de composición (cablea las capas)
```

- **Regla de dependencias**: `domain ← application ← infrastructure / interfaces`. La verifica `internal/architecture_test.go`.
- **Agregados**: `Board`, `Column` (límite de WIP y tipo), `Card` (con sus etiquetas y eventos de movimiento) y `Label`; `User` en el contexto de identidad.
- **Regla de WIP** en el dominio (`Column.CanAccept`, `MoveCard`, `CreateCard`). Para que sea correcta bajo concurrencia, la unidad de trabajo bloquea las columnas (en orden de id) y usa `READ COMMITTED`.
- **CQRS ligero**: los comandos pasan por los agregados; las lecturas (tablero, historial, métricas) van por un `ReadModel` con SQL directo.

## Base de datos
Las migraciones están en `backend/internal/infrastructure/mysql/migrations/*.sql` y se aplican solas al arrancar el backend (tabla `schema_migrations`). Para empezar de cero: `docker compose down -v`.

## Pruebas
- `docker compose exec backend go test ./...`: pruebas unitarias del dominio (WIP, movimientos, cycle time) y de la regla de dependencias entre capas.
- `scripts/smoke.sh`: prueba de humo de la API (auth, WIP, historial, cycle time, etiquetas) contra el stack levantado.
- `scripts/race.sh` y `scripts/race_create.sh`: concurrencia sobre el límite de WIP (mover y crear tarjetas a la vez).
