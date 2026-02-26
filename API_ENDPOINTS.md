# 📚 Documentación de Endpoints - QuickScore API

## 🔧 Base URL
```
http://localhost:8090
```

## 📋 Tabla de Contenidos
- [Autenticación](#autenticación)
- [Salas](#salas)
- [Puntuación](#puntuación)
- [WebSocket](#websocket)

---

## 🔐 Autenticación

### 1. Registrar Usuario
Crea un nuevo usuario en el sistema (host o participante).

**Endpoint:** `POST /auth/register`

**Headers:**
```json
Content-Type: application/json
```

**Body:**
```json
{
  "email": "usuario@ejemplo.com",
  "name": "Nombre Usuario",
  "password": "password123",
  "role": "host"  // "host" o "participant"
}
```

**Respuesta exitosa (201):**
```json
{
  "id": 1,
  "email": "usuario@ejemplo.com",
  "name": "Nombre Usuario",
  "role": "host"
}
```

**Errores posibles:**
- `400`: El email ya está registrado
- `400`: Datos inválidos

---

### 2. Iniciar Sesión (Login)
Autentica un usuario y devuelve un token JWT.

**Endpoint:** `POST /auth/login`

**Headers:**
```json
Content-Type: application/json
```

**Body:**
```json
{
  "email": "usuario@ejemplo.com",
  "password": "password123"
}
```

**Respuesta exitosa (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "Nombre Usuario",
    "email": "usuario@ejemplo.com",
    "role": "host"
  }
}
```

**Errores posibles:**
- `401`: Credenciales inválidas
- `400`: Datos inválidos

**Nota:** El token expira en 24 horas.

---

## 🏠 Salas

### 3. Crear Sala
Crea una nueva sala de juego. Solo usuarios con rol `host` pueden crear salas.

**Endpoint:** `POST /rooms`

**Headers:**
```json
Content-Type: application/json
Authorization: Bearer <TOKEN_HOST>
```

**Body:**
```json
{
  "name": "Nombre de la Sala"
}
```

**Respuesta exitosa (201):**
```json
{
  "code": "ABC123",
  "status": "waiting"
}
```

**Errores posibles:**
- `401`: Token inválido o no proporcionado
- `403`: Solo hosts pueden crear salas
- `400`: Datos inválidos

**Estados de sala:**
- `waiting`: Esperando jugadores
- `active`: Sesión en curso
- `finished`: Sesión finalizada

---

### 4. Obtener Información de Sala
Obtiene los detalles de una sala específica.

**Endpoint:** `GET /rooms/{code}`

**Headers:**
```json
Authorization: Bearer <TOKEN>
```

**Parámetros URL:**
- `code`: Código de la sala (ej: "ABC123")

**Respuesta exitosa (200):**
```json
{
  "id": 1,
  "code": "ABC123",
  "host_id": 1,
  "status": "waiting",
  "created_at": "2026-02-26T10:00:00Z"
}
```

**Errores posibles:**
- `401`: Token inválido
- `404`: Sala no encontrada

---

### 5. Unirse a una Sala
Permite a un participante unirse a una sala existente.

**Endpoint:** `POST /rooms/{code}/join`

**Headers:**
```json
Authorization: Bearer <TOKEN>
```

**Parámetros URL:**
- `code`: Código de la sala

**Respuesta exitosa (200):**
```json
{
  "message": "te uniste a la sala"
}
```

**Errores posibles:**
- `401`: Token inválido
- `404`: Sala no encontrada
- `400`: Ya estás en la sala

---

### 6. Iniciar Sesión de Sala
Comienza una sesión de juego. Solo el host de la sala puede iniciarla.

**Endpoint:** `PATCH /rooms/{code}/start`

**Headers:**
```json
Authorization: Bearer <TOKEN_HOST>
```

**Parámetros URL:**
- `code`: Código de la sala

**Respuesta exitosa (200):**
```json
{
  "message": "sesión iniciada"
}
```

**Errores posibles:**
- `401`: Token inválido
- `403`: Solo el host puede iniciar la sesión
- `400`: La sala ya está activa o finalizada

**Cambios:**
- Estado de la sala: `waiting` → `active`

---

### 7. Finalizar Sesión de Sala
Termina una sesión de juego activa. Solo el host puede finalizarla.

**Endpoint:** `PATCH /rooms/{code}/end`

**Headers:**
```json
Authorization: Bearer <TOKEN_HOST>
```

**Parámetros URL:**
- `code`: Código de la sala

**Respuesta exitosa (200):**
```json
{
  "message": "sesión finalizada"
}
```

**Errores posibles:**
- `401`: Token inválido
- `403`: Solo el host puede finalizar la sesión
- `400`: La sala no está activa

**Cambios:**
- Estado de la sala: `active` → `finished`

---

## 🎯 Puntuación

### 8. Agregar/Restar Puntos
Modifica los puntos de un participante. Solo el host puede hacerlo.

**Endpoint:** `POST /rooms/{code}/score`

**Headers:**
```json
Content-Type: application/json
Authorization: Bearer <TOKEN_HOST>
```

**Parámetros URL:**
- `code`: Código de la sala

**Body:**
```json
{
  "target_user_id": 2,
  "delta": 10  // Positivo para sumar, negativo para restar
}
```

**Respuesta exitosa (200):**
```json
{
  "message": "puntos actualizados"
}
```

**Errores posibles:**
- `401`: Token inválido
- `403`: Solo el host puede modificar puntos
- `404`: Usuario o sala no encontrada
- `400`: El usuario no está en la sala

**Ejemplos:**
- `"delta": 10` → Suma 10 puntos
- `"delta": -5` → Resta 5 puntos

---

### 9. Obtener Ranking
Obtiene la tabla de posiciones de una sala, ordenada por puntos de mayor a menor.

**Endpoint:** `GET /rooms/{code}/ranking`

**Headers:**
```json
Authorization: Bearer <TOKEN>
```

**Parámetros URL:**
- `code`: Código de la sala

**Respuesta exitosa (200):**
```json
[
  {
    "user_id": 2,
    "user_name": "Player One",
    "points": 40,
    "position": 1
  },
  {
    "user_id": 3,
    "user_name": "Player Two",
    "points": 25,
    "position": 2
  }
]
```

**Errores posibles:**
- `401`: Token inválido
- `404`: Sala no encontrada

**Notas:**
- Los participantes se ordenan por puntos (de mayor a menor)
- La posición se asigna automáticamente

---

## 🔌 WebSocket

### 10. Conectar al WebSocket
Establece una conexión WebSocket en tiempo real para recibir actualizaciones de la sala.

**Endpoint:** `GET /ws`

**Parámetros Query:**
- `room`: Código de la sala (requerido)
- `token`: Token JWT del usuario (opcional, recomendado)

**Ejemplo de conexión:**
```
ws://localhost:8090/ws?room=ABC123&token=eyJhbGc...
```

**Mensajes recibidos:**
El servidor enviará mensajes JSON cuando ocurran eventos en la sala:

```json
{
  "type": "score_update",
  "data": {
    "user_id": 2,
    "user_name": "Player One",
    "points": 40
  }
}
```

**Tipos de eventos:**
- `score_update`: Cambio en la puntuación
- `session_started`: Sesión iniciada
- `session_ended`: Sesión finalizada
- `user_joined`: Nuevo participante

**Errores posibles:**
- `400`: Parámetro `room` no proporcionado
- Conexión rechazada: Código de sala inválido

---

## 🔒 Autenticación y Autorización

### Roles de Usuario
- **host**: Puede crear salas, iniciar/finalizar sesiones, modificar puntos
- **participant**: Puede unirse a salas y ver información

### Headers de Autenticación
Todos los endpoints protegidos requieren el header:
```
Authorization: Bearer <TOKEN_JWT>
```

### Obtención del Token
1. Registra un usuario con `POST /auth/register`
2. Inicia sesión con `POST /auth/login`
3. Usa el `token` devuelto en los siguientes requests

---

## 🚀 Flujo Completo de Uso

### Escenario: Crear y gestionar una sesión de juego

**1. El host se registra e inicia sesión**
```bash
POST /auth/register → crea cuenta host
POST /auth/login → obtiene token_host
```

**2. El host crea una sala**
```bash
POST /rooms → recibe código de sala (ej: "ABC123")
```

**3. Los participantes se registran y unen a la sala**
```bash
POST /auth/register → crea cuenta participant
POST /auth/login → obtiene token_participant
POST /rooms/ABC123/join → se une a la sala
```

**4. El host inicia la sesión**
```bash
PATCH /rooms/ABC123/start → sala pasa a estado "active"
```

**5. El host asigna puntos durante el juego**
```bash
POST /rooms/ABC123/score → suma o resta puntos
```

**6. Los participantes consultan el ranking**
```bash
GET /rooms/ABC123/ranking → ve tabla de posiciones
```

**7. El host finaliza la sesión**
```bash
PATCH /rooms/ABC123/end → sala pasa a estado "finished"
```

---

## 📊 Códigos de Estado HTTP

| Código | Significado |
|--------|-------------|
| 200 | Solicitud exitosa |
| 201 | Recurso creado exitosamente |
| 204 | Sin contenido (preflight CORS) |
| 400 | Solicitud incorrecta o datos inválidos |
| 401 | No autenticado o token inválido |
| 403 | Sin permisos para realizar la acción |
| 404 | Recurso no encontrado |
| 500 | Error interno del servidor |

---

## 🌐 CORS

La API está configurada para aceptar peticiones desde cualquier origen:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Authorization, Content-Type`

---

## 📖 Swagger UI

Documentación interactiva disponible en:
```
http://localhost:8090/docs/
```

Desde Swagger puedes probar todos los endpoints directamente desde el navegador.

---

## 🛠️ Ejemplos con cURL

### Crear un host y una sala
```bash
# 1. Registrar host
curl -X POST http://localhost:8090/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"host@test.com","name":"Host","password":"pass123","role":"host"}'

# 2. Login host
TOKEN=$(curl -X POST http://localhost:8090/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"host@test.com","password":"pass123"}' \
  | jq -r '.token')

# 3. Crear sala
curl -X POST http://localhost:8090/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Mi Sala"}'
```

### Participante se une y consulta ranking
```bash
# 1. Registrar participante
curl -X POST http://localhost:8090/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"player@test.com","name":"Player","password":"pass123","role":"participant"}'

# 2. Login participante
PLAYER_TOKEN=$(curl -X POST http://localhost:8090/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"player@test.com","password":"pass123"}' \
  | jq -r '.token')

# 3. Unirse a sala
curl -X POST http://localhost:8090/rooms/ABC123/join \
  -H "Authorization: Bearer $PLAYER_TOKEN"

# 4. Ver ranking
curl -X GET http://localhost:8090/rooms/ABC123/ranking \
  -H "Authorization: Bearer $PLAYER_TOKEN"
```

---

## 🐛 Troubleshooting

### Error: "token inválido o expirado"
- Verifica que el token no haya expirado (24h de validez)
- Asegúrate de incluir "Bearer " antes del token
- Vuelve a hacer login para obtener un nuevo token

### Error: "solo el host puede realizar esta acción"
- Verifica que el usuario que intenta la acción tenga rol "host"
- Confirma que seas el creador de la sala

### Error: "el usuario no está en la sala"
- El participante debe unirse primero con POST /rooms/{code}/join
- Verifica que el user_id sea correcto

---

**Versión:** 1.0  
**Última actualización:** 26 de febrero de 2026
