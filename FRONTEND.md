# API Golan — Guía de Integración para Frontend

**Base URL:** `http://localhost:8090`  
**Formato:** JSON en todos los endpoints  
**Auth:** Header `Authorization: Bearer <token>` en rutas protegidas

---

## Autenticación

### Registrarse
```
POST /auth/register
```
**Body:**
```json
{
  "name": "Profe Angel",
  "email": "angel@test.com",
  "password": "123456",
  "role": "host"
}
```
> `role` puede ser `"host"` o `"participant"`. Si no se manda, por defecto es `"participant"`.

**Respuesta exitosa `201`:**
```json
{
  "id": 1,
  "name": "Profe Angel",
  "email": "angel@test.com",
  "role": "host"
}
```

**Errores posibles:**
| Código | Mensaje |
|---|---|
| 400 | `"el email ya está registrado"` |
| 400 | `"nombre, email y contraseña son requeridos"` |

---

### Login
```
POST /auth/login
```
**Body:**
```json
{
  "email": "angel@test.com",
  "password": "123456"
}
```

**Respuesta exitosa `200`:**
```json
{
  "token": "eyJhbGci...",
  "user": {
    "id": 1,
    "name": "Profe Angel",
    "email": "angel@test.com",
    "role": "host"
  }
}
```
> Guarda el `token` en el almacenamiento local del frontend. Dura **24 horas**.

**Errores posibles:**
| Código | Mensaje |
|---|---|
| 401 | `"credenciales inválidas"` |

---

## Salas

### Crear sala 🔒 Solo host
```
POST /rooms
Headers: Authorization: Bearer <token>
```
**Sin body.**

**Respuesta exitosa `201`:**
```json
{
  "code": "KD7B45",
  "status": "waiting"
}
```
> Guarda el `code`. Es el que los alumnos usan para unirse.

---

### Ver sala 🔒
```
GET /rooms/{code}
Headers: Authorization: Bearer <token>
```

**Respuesta exitosa `200`:**
```json
{
  "id": 1,
  "code": "KD7B45",
  "host_id": 1,
  "status": "waiting",
  "created_at": "2026-02-24T07:17:00Z"
}
```

---

### Unirse a una sala 🔒 Solo participantes
```
POST /rooms/{code}/join
Headers: Authorization: Bearer <token>
```
**Sin body.**

**Respuesta exitosa `200`:**
```json
{
  "message": "te uniste a la sala"
}
```

**Errores posibles:**
| Código | Mensaje |
|---|---|
| 400 | `"sala no encontrada"` |
| 400 | `"ya estás en esta sala"` |
| 400 | `"la sala ya terminó"` |

---

### Iniciar sesión 🔒 Solo host
```
PATCH /rooms/{code}/start
Headers: Authorization: Bearer <token>
```
**Sin body.**

**Respuesta exitosa `200`:**
```json
{
  "message": "sesión iniciada"
}
```
> También hace broadcast por WebSocket con evento `session_started` a todos en la sala.

---

### Terminar sesión 🔒 Solo host
```
PATCH /rooms/{code}/end
Headers: Authorization: Bearer <token>
```
**Sin body.**

**Respuesta exitosa `200`:**
```json
{
  "message": "sesión finalizada"
}
```
> También hace broadcast por WebSocket con evento `session_ended`.

---

## Puntos y Ranking

### Dar / quitar puntos 🔒 Solo host
```
POST /rooms/{code}/score
Headers: Authorization: Bearer <token>
         Content-Type: application/json
```
**Body:**
```json
{
  "target_user_id": 2,
  "delta": 10
}
```
> `delta` puede ser positivo (`+10`, `+5`) o negativo (`-5`). La sesión debe estar `active`.

**Respuesta exitosa `200`:**
```json
{
  "message": "puntos actualizados"
}
```
> También hace broadcast por WebSocket con evento `score_update` y el ranking completo a todos.

**Errores posibles:**
| Código | Mensaje |
|---|---|
| 400 | `"solo el host puede modificar los puntos"` |
| 400 | `"la sesión no está activa"` |

---

### Ver ranking 🔒
```
GET /rooms/{code}/ranking
Headers: Authorization: Bearer <token>
```

**Respuesta exitosa `200`:**
```json
[
  { "user_id": 2, "user_name": "Alumno 1", "points": 30, "position": 1 },
  { "user_id": 3, "user_name": "Alumno 2", "points": 20, "position": 2 }
]
```

---

## WebSocket (tiempo real)

### Conectarse
```
ws://localhost:8090/ws?room=KD7B45
```

**Ejemplo en JavaScript:**
```javascript
const socket = new WebSocket("ws://localhost:8090/ws?room=KD7B45");

socket.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  switch (msg.event) {
    case "score_update":
      // msg.payload = array del ranking actualizado
      actualizarRanking(msg.payload);
      break;

    case "session_started":
      // La sesión comenzó
      mostrarPantallDeJuego();
      break;

    case "session_ended":
      // La sesión terminó
      mostrarResultadoFinal();
      break;
  }
};
```

**Formato de todos los mensajes recibidos:**
```json
{
  "event": "score_update",
  "room": "KD7B45",
  "payload": { ... }
}
```

---

## Errores generales de autenticación

| Código | Mensaje | Causa |
|---|---|---|
| 401 | `"token requerido"` | No se mandó el header Authorization |
| 401 | `"token inválido o expirado"` | Token malo o vencido (genera uno nuevo con login) |
| 403 | `"solo el host puede realizar esta acción"` | Un participante intentó una acción de host |

---

## Flujo completo resumido

```
1. Host:        POST /auth/register  (role: host)
2. Host:        POST /auth/login     → guarda token
3. Host:        POST /rooms          → guarda code (ej: KD7B45)

4. Alumno:      POST /auth/register  (role: participant)
5. Alumno:      POST /auth/login     → guarda token
6. Alumno:      POST /rooms/KD7B45/join
7. Alumno:      Conectar WebSocket   ws://.../ws?room=KD7B45

8. Host:        PATCH /rooms/KD7B45/start
                ↳ WebSocket broadcast → "session_started"

9. Host:        POST /rooms/KD7B45/score  { target_user_id: X, delta: 10 }
                ↳ WebSocket broadcast → "score_update" con ranking

10. Cualquiera: GET /rooms/KD7B45/ranking

11. Host:       PATCH /rooms/KD7B45/end
                ↳ WebSocket broadcast → "session_ended"
```
