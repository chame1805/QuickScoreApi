# API Golan — Documentación Técnica

## Descripción
API REST con WebSocket para competencias en tiempo real. Construida con Go, MySQL y arquitectura hexagonal.

## Stack
- **Lenguaje:** Go 1.25
- **Base de datos:** MySQL 8.0
- **WebSocket:** Gorilla WebSocket
- **Autenticación:** JWT (HS256, 24h de vigencia)
- **Hot reload:** Air
- **Contenedores:** Docker + Docker Compose

---

## Arquitectura Hexagonal

```
src/
├── domain/              ← Entidades e interfaces (núcleo, sin dependencias externas)
├── core/                ← Lógica de negocio pura
├── applications/
│   └── usecase/         ← Casos de uso (orquestación)
└── infrastructure/
    ├── db/              ← Conexión MySQL
    ├── repository/      ← Implementación concreta de repositorios
    ├── jwt/             ← Generación y validación de tokens
    └── http/
        ├── handler/     ← Controladores HTTP
        ├── middleware/  ← Autenticación y autorización
        └── router/      ← Registro de rutas
    └── websocket/       ← Hub de conexiones en tiempo real
```

---

## Entidades del Dominio

### User
```go
type User struct {
    ID        int       // identificador único
    Name      string    // nombre del usuario
    Email     string    // email (único)
    Password  string    // bcrypt hash (nunca se expone en JSON)
    Role      Role      // "host" | "participant"
    CreatedAt time.Time
}
```

### Room
```go
type Room struct {
    ID        int
    Code      string     // código de 6 caracteres (ej: "KD7B45")
    HostID    int        // ID del usuario host que la creó
    Status    RoomStatus // "waiting" | "active" | "finished"
    CreatedAt time.Time
}
```

### Participant
```go
type Participant struct {
    ID       int
    RoomID   int
    UserID   int
    JoinedAt time.Time
}
```

### Score
```go
type Score struct {
    ID        int
    RoomID    int
    UserID    int
    Points    int       // puede ser positivo o negativo
    UpdatedAt time.Time
}
```

### RankingEntry (respuesta del ranking)
```go
type RankingEntry struct {
    UserID   int
    UserName string
    Points   int
    Position int  // calculado en tiempo real
}
```

---

## Tablas MySQL

```sql
users        → id, name, email, password, role, created_at
rooms        → id, code, host_id, status, created_at
participants → id, room_id, user_id, joined_at
scores       → id, room_id, user_id, points, updated_at
```

---

## Reglas de Negocio

| Regla | Descripción |
|---|---|
| Solo host crea sala | El endpoint `POST /rooms` requiere `role: host` |
| Solo host inicia/termina | `start` y `end` validan que el requester sea el `host_id` de esa sala |
| Solo host da puntos | `POST /rooms/:code/score` valida rol host |
| Estado de sala | El flujo es estrictamente `waiting → active → finished` |
| Score inicial | Al unirse a una sala el participante arranca con 0 puntos |
| Email único | No se pueden registrar dos usuarios con el mismo email |
| Participante único | Un usuario solo puede unirse una vez a la misma sala |

---

## Levantar el proyecto

```bash
# Primera vez (construye imagen y levanta contenedores)
docker compose up --build

# Siguientes veces
docker compose up

# Bajar contenedores
docker compose down

# Bajar y borrar la base de datos (para reiniciar desde cero)
docker compose down -v
```

### Puertos
| Servicio | Puerto externo | Puerto interno |
|---|---|---|
| API Go | 8090 | 8080 |
| MySQL | 3307 | 3306 |

---

## Variables de entorno (`.env`)

```env
MYSQL_ROOT_PASSWORD=rootpassword
MYSQL_DATABASE=apidb
MYSQL_USER=apiuser
MYSQL_PASSWORD=apipassword
JWT_SECRET=cambia_esto_por_una_clave_segura_en_produccion
```

---

## Eventos WebSocket

Conexión: `ws://localhost:8090/ws?room=CODIGO_SALA`

| Evento | Dirección | Cuándo se dispara |
|---|---|---|
| `session_started` | Server → Todos | El host inicia la sesión |
| `session_ended` | Server → Todos | El host termina la sesión |
| `score_update` | Server → Todos | El host modifica puntos |

Formato de mensaje:
```json
{
  "event": "score_update",
  "room": "KD7B45",
  "payload": [ ...ranking actualizado... ]
}
```
