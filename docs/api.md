# API Calli Machtia

**Base URL:** `https://calli-machtia-backend.up.railway.app`

---

## Tabla de Contenidos

- [Autenticación](#autenticación)
- [Cursos](#cursos)
- [Módulos y Lecciones (MongoDB)](#módulos-y-lecciones-mongodb)
- [Inscripciones](#inscripciones)
- [Pagos (Stripe)](#pagos-stripe)
- [Admin](#admin)
- [Públicos](#públicos)
- [Códigos de Error](#códigos-de-error)

---

## Formato General

### Respuesta exitosa
```json
{
  "data": { ... }
}
```

### Error
```json
{
  "error": "mensaje descriptivo"
}
```

### Autenticación
Todas las rutas protegidas usan:
```
Authorization: Bearer <access_token>
```

El `accessToken` expira en **15 minutos**. Usar `POST /api/auth/refresh` para obtener uno nuevo.

---

## Autenticación

### POST /api/auth/register

Crear cuenta nueva.

**Body:**
```json
{
  "name": "Juan Pérez",
  "email": "juan@email.com",
  "password": "miPassword123",
  "role": "alumno"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| name | string | sí |
| email | string (email) | sí |
| password | string (min 6) | sí |
| role | "alumno" \| "instructor" | no (default: "alumno") |

**Respuesta 201:**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "Juan Pérez",
      "email": "juan@email.com",
      "role": "alumno"
    },
    "accessToken": "jwt...",
    "refreshToken": "jwt..."
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Error de validación Zod |
| 409 | El email ya está registrado |

---

### POST /api/auth/login

Iniciar sesión.

**Body:**
```json
{
  "email": "juan@email.com",
  "password": "miPassword123"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| email | string (email) | sí |
| password | string | sí |

**Respuesta 200:**
```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "Juan Pérez",
      "email": "juan@email.com",
      "role": "alumno",
      "avatar_url": null,
      "bio": null,
      "created_at": "2026-01-01T00:00:00.000Z"
    },
    "accessToken": "jwt...",
    "refreshToken": "jwt..."
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Error de validación |
| 401 | Email o contraseña incorrectos |

---

### GET /api/auth/me

Obtener datos del usuario autenticado.

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": {
    "id": "uuid",
    "name": "Juan Pérez",
    "email": "juan@email.com",
    "role": "alumno",
    "avatar_url": null,
    "bio": null,
    "created_at": "2026-01-01T00:00:00.000Z"
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 401 | Token inválido o expirado |
| 404 | Usuario no encontrado |

---

### POST /api/auth/refresh

Renovar access token usando refresh token.

**Body:**
```json
{
  "refresh_token": "jwt..."
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| refresh_token | string | sí |

**Respuesta 200:**
```json
{
  "data": {
    "accessToken": "jwt..."
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Error de validación |
| 401 | Refresh token inválido o expirado |

---

### PUT /api/auth/profile

Actualizar perfil del usuario.

**Headers:** `Authorization: Bearer <token>`

**Body** (al menos un campo):
```json
{
  "name": "Juan Pérez Actualizado",
  "avatar_url": "https://...",
  "bio": "Instructor de matemáticas"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| name | string (max 100) | no |
| avatar_url | string (max 500) | no |
| bio | string (max 1000) | no |

**Respuesta 200:** Mismo formato que `GET /api/auth/me`.

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Error de validación / No hay datos para actualizar |
| 404 | Usuario no encontrado |

---

### POST /api/auth/logout

Cerrar sesión (invalida refresh token en Redis).

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": {
    "message": "Sesión cerrada exitosamente"
  }
}
```

---

### POST /api/auth/forgot-password

Solicitar recuperación de contraseña.

**Body:**
```json
{
  "email": "juan@email.com"
}
```

**Respuesta 200 (siempre, aunque el email no exista):**
```json
{
  "data": {
    "message": "Si el email existe, recibirás un enlace de recuperación"
  }
}
```

---

### POST /api/auth/reset-password

Restablecer contraseña con token.

**Body:**
```json
{
  "token": "token_hex_64_chars",
  "password": "nuevaPassword123"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| token | string | sí |
| password | string (min 6) | sí |

**Respuesta 200:**
```json
{
  "data": {
    "message": "Contraseña actualizada exitosamente"
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Token inválido o ya utilizado / Token expirado |

---

## Cursos

### GET /api/courses

Listar cursos públicos (solo publicados).

**Query params:**

| Parámetro | Tipo | Default | Descripción |
|-----------|------|---------|-------------|
| page | number | 1 | Número de página |
| limit | number | 10 | Items por página (max 100) |
| category | string | — | Filtrar por categoría exacta |
| search | string | — | Búsqueda en título y descripción |

**Respuesta 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "instructor_id": "uuid",
      "title": "Curso de ejemplo",
      "slug": "curso-de-ejemplo-a1b2c3",
      "description": "Descripción del curso",
      "image_url": "https://...",
      "price": "29.99",
      "category": "programación",
      "tags": ["javascript", "web"],
      "published": true,
      "seo_title": null,
      "seo_description": null,
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-01-01T00:00:00.000Z",
      "instructor_name": "Juan Pérez",
      "enrollment_count": 42
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 50,
    "totalPages": 5
  }
}
```

---

### GET /api/courses/slug/:slug

Obtener curso por slug.

**Respuesta 200:** Mismo objeto `data` que en el listado (con `instructor_name` y `enrollment_count`).

**Error 404:** Curso no encontrado

---

### GET /api/courses/:id

Obtener curso por ID.

**Respuesta 200:** Mismo formato que `GET /slug/:slug`.

**Error 404:** Curso no encontrado

---

### POST /api/courses

Crear curso (instructor o admin).

**Headers:** `Authorization: Bearer <token>`

**Body:**
```json
{
  "title": "Mi Curso",
  "description": "Descripción opcional",
  "image_url": "https://...",
  "price": 19.99,
  "category": "diseño",
  "tags": ["ux", "ui"],
  "published": false,
  "seo_title": "Título SEO",
  "seo_description": "Descripción SEO"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| title | string (max 200) | sí |
| description | string | no |
| image_url | string (max 500) | no |
| price | number (min 0) | sí |
| category | string | no |
| tags | string[] | no |
| published | boolean | no (default: false) |
| seo_title | string (max 200) | no |
| seo_description | string (max 500) | no |

> El `slug` se genera automáticamente desde el título.

**Respuesta 201:** Objeto del curso creado.

---

### PUT /api/courses/:id

Actualizar curso (instructor propietario o admin).

**Headers:** `Authorization: Bearer <token>`

**Body** (todos opcionales):
```json
{
  "title": "Nuevo título",
  "description": "Nueva descripción",
  "price": 39.99,
  "published": true,
  "category": "nueva-categoria",
  "tags": ["nuevo"],
  "image_url": "https://...",
  "seo_title": "Nuevo SEO",
  "seo_description": "Nueva descripción SEO"
}
```

**Respuesta 200:** Objeto del curso actualizado.

**Errores:**
| Código | Mensaje |
|--------|---------|
| 403 | No tienes permiso para modificar este curso |
| 404 | Curso no encontrado |

---

### DELETE /api/courses/:id

Eliminar curso (instructor propietario o admin). Hard delete.

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": {
    "message": "Curso eliminado exitosamente"
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 403 | No tienes permiso para eliminar este curso |
| 404 | Curso no encontrado |

---

### GET /api/courses/instructor/mine

Listar cursos del instructor autenticado.

**Headers:** `Authorization: Bearer <token>` (rol: instructor o admin)

**Respuesta 200:** Array de cursos (sin paginación, ordenado por created_at).

---

## Módulos y Lecciones (MongoDB)

### GET /api/courses/:id/modules

Obtener todos los módulos de un curso.

**Respuesta 200:**
```json
{
  "data": [
    {
      "_id": "uuid",
      "course_id": "uuid",
      "title": "Módulo 1",
      "description": "Introducción",
      "order": 0,
      "lessons": [
        {
          "id": "uuid",
          "title": "Lección 1",
          "description": "Primera lección",
          "content": "Contenido markdown/html",
          "video_url": "https://...",
          "duration": 600,
          "order": 0,
          "free": true
        }
      ],
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-01-01T00:00:00.000Z"
    }
  ]
}
```

**Error 404:** Curso no encontrado

---

### GET /api/courses/:id/modules/:moduleId

Obtener un módulo específico.

**Respuesta 200:** Objeto del módulo (mismo formato que en el array).

**Error 404:** Módulo no encontrado

---

### POST /api/courses/:id/modules

Crear módulo (instructor o admin).

**Headers:** `Authorization: Bearer <token>`

**Body:**
```json
{
  "title": "Módulo 1",
  "description": "Descripción",
  "order": 0
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| title | string (max 200) | sí |
| description | string | no |
| order | number (entero) | no (default: 0) |

**Respuesta 201:** Objeto del módulo creado.

---

### PUT /api/courses/:id/modules/:moduleId

Actualizar módulo.

**Headers:** `Authorization: Bearer <token>`

**Body:** Mismos campos que creación (todos opcionales en la lógica de actualización, pero `title` es obligatorio según schema).

**Respuesta 200:** Módulo actualizado.

---

### DELETE /api/courses/:id/modules/:moduleId

Eliminar módulo y sus lecciones.

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": {
    "message": "Módulo eliminado exitosamente"
  }
}
```

---

### POST /api/courses/:id/modules/:moduleId/lessons

Agregar lección a un módulo.

**Headers:** `Authorization: Bearer <token>`

**Body:**
```json
{
  "title": "Lección 1",
  "description": "Descripción",
  "content": "Contenido en markdown",
  "video_url": "https://...",
  "duration": 600,
  "order": 0,
  "free": false
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| title | string (max 200) | sí |
| description | string | no |
| content | string | no |
| video_url | string | no |
| duration | number (entero, min 0) | no |
| order | number (entero) | no |
| free | boolean | no (default: false) |

**Respuesta 201:**
```json
{
  "data": {
    "id": "uuid",
    "title": "Lección 1",
    "description": "...",
    "content": "...",
    "video_url": "...",
    "duration": 600,
    "order": 0,
    "free": false
  }
}
```

---

### PUT /api/courses/:id/modules/:moduleId/lessons/:lessonId

Actualizar lección.

**Headers:** `Authorization: Bearer <token>`

**Body:** Mismos campos que creación.

**Respuesta 200:** Lección actualizada.

---

### DELETE /api/courses/:id/modules/:moduleId/lessons/:lessonId

Eliminar lección.

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": {
    "message": "Lección eliminada exitosamente"
  }
}
```

---

## Inscripciones

### POST /api/enrollments

Inscribirse a un curso.

**Headers:** `Authorization: Bearer <token>`

**Body:**
```json
{
  "course_id": "uuid"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| course_id | string (uuid) | sí |

**Respuesta 201:**
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "course_id": "uuid",
    "status": "active",
    "progress": "0",
    "enrolled_at": "2026-01-01T00:00:00.000Z",
    "completed_at": null,
    "created_at": "2026-01-01T00:00:00.000Z",
    "updated_at": "2026-01-01T00:00:00.000Z"
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 404 | Curso no encontrado |
| 409 | Ya estás inscrito en este curso |

---

### GET /api/enrollments/mine

Mis inscripciones (con datos del curso).

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "course_id": "uuid",
      "status": "active",
      "progress": "0",
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-01-01T00:00:00.000Z",
      "course": {
        "id": "uuid",
        "title": "Curso",
        "slug": "curso-slug",
        "description": "...",
        "image_url": "...",
        "price": "29.99",
        "category": "programación",
        "instructor_name": "Juan Pérez"
      }
    }
  ]
}
```

---

### GET /api/enrollments/course/:courseId

Listar inscripciones de un curso (instructor propietario o admin).

**Headers:** `Authorization: Bearer <token>` (rol: instructor o admin)

**Respuesta 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "course_id": "uuid",
      "status": "active",
      "progress": "0",
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-01-01T00:00:00.000Z",
      "user": {
        "id": "uuid",
        "name": "Juan Pérez",
        "email": "juan@email.com",
        "avatar_url": null
      }
    }
  ]
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 403 | No tienes permiso para ver las inscripciones de este curso |
| 404 | Curso no encontrado |

---

### PUT /api/enrollments/:id/progress

Actualizar progreso.

**Headers:** `Authorization: Bearer <token>`

**Body:**
```json
{
  "progress": 50
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| progress | number (0-100) | sí |

**Respuesta 200:** Objeto de la inscripción actualizado.

**Errores:**
| Código | Mensaje |
|--------|---------|
| 403 | No tienes permiso para actualizar esta inscripción |
| 404 | Inscripción no encontrada |

---

### DELETE /api/enrollments/:id

Cancelar inscripción (propietario o admin).

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": {
    "message": "Inscripción cancelada exitosamente"
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 403 | No tienes permiso para eliminar esta inscripción |
| 404 | Inscripción no encontrada |

---

## Pagos (Stripe)

### POST /api/payments/create-intent

Crear PaymentIntent de Stripe.

**Headers:** `Authorization: Bearer <token>`

**Body:**
```json
{
  "course_id": "uuid"
}
```

**Respuesta 200:**
```json
{
  "data": {
    "payment_intent_id": "pi_...",
    "client_secret": "pi_..._secret_...",
    "amount": 2999,
    "currency": "usd"
  }
}
```

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Ya estás inscrito en este curso |
| 404 | Curso no encontrado |
| 500 | Stripe no está configurado / Error al procesar el pago |

---

### POST /api/payments/webhook

Webhook de Stripe (sin autenticación JWT).

**Headers:** `stripe-signature: <firma>`

**Body:** Raw request body (el evento de Stripe).

**Respuesta 200:**
```json
{
  "received": true
}
```

**Eventos manejados:**
- `payment_intent.succeeded`: marca pago como `completed`, crea inscripción si no existe
- `payment_intent.payment_failed`: marca pago como `failed`

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Firma de Stripe requerida / Firma inválida |
| 500 | Stripe no está configurado |

---

### GET /api/payments

Historial de pagos del usuario.

**Headers:** `Authorization: Bearer <token>`

**Respuesta 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "amount": "29.99",
      "currency": "usd",
      "status": "completed",
      "created_at": "2026-01-01T00:00:00.000Z",
      "course": {
        "id": "uuid",
        "title": "Curso",
        "slug": "curso-slug"
      }
    }
  ]
}
```

Ordenado por `created_at` descendente.

---

## Admin

Todas las rutas admin requieren: `Authorization: Bearer <token>` con rol `admin`.

### GET /api/admin/users

Listar usuarios.

**Query params:**

| Parámetro | Tipo | Default | Descripción |
|-----------|------|---------|-------------|
| page | number | 1 | |
| limit | number | 10 | max 100 |
| search | string | — | Búsqueda en nombre y email |
| role | "alumno" \| "instructor" \| "admin" | — | Filtrar por rol |

**Respuesta 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Juan Pérez",
      "email": "juan@email.com",
      "role": "alumno",
      "avatar_url": null,
      "bio": null,
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-01-01T00:00:00.000Z"
    }
  ],
  "pagination": { "page": 1, "limit": 10, "total": 100, "totalPages": 10 }
}
```

---

### PUT /api/admin/users/:id/role

Cambiar rol de usuario.

**Body:**
```json
{
  "role": "instructor"
}
```

| Campo | Tipo | Obligatorio |
|-------|------|-------------|
| role | "alumno" \| "instructor" \| "admin" | sí |

**Respuesta 200:** Objeto del usuario actualizado.

**Errores:**
| Código | Mensaje |
|--------|---------|
| 400 | Rol inválido. Debe ser: alumno, instructor o admin |
| 404 | Usuario no encontrado |

---

### GET /api/admin/courses

Listar todos los cursos (publicados y no publicados).

**Query params:** `page`, `limit`

**Respuesta 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "instructor_id": "uuid",
      "title": "Curso",
      "slug": "curso-slug",
      "description": "...",
      "price": "29.99",
      "category": "programación",
      "published": true,
      "created_at": "...",
      "updated_at": "...",
      "instructor_name": "Juan Pérez"
    }
  ],
  "pagination": { "page": 1, "limit": 10, "total": 50, "totalPages": 5 }
}
```

---

### GET /api/admin/stats

Estadísticas de la plataforma.

**Respuesta 200:**
```json
{
  "data": {
    "total_users": 150,
    "total_courses": 30,
    "total_enrollments": 200,
    "total_revenue": 4999.99
  }
}
```

> `total_revenue` suma de pagos con status `completed`.

---

### DELETE /api/admin/courses/:id

Eliminar curso permanentemente (sin restricción de propietario).

**Respuesta 200:**
```json
{
  "data": {
    "message": "Curso eliminado permanentemente"
  }
}
```

---

## Públicos

### GET /

Información de la API.

**Respuesta 200:**
```json
{
  "name": "Calli Machtia API",
  "version": "1.0.0",
  "environment": "production"
}
```

---

### GET /health

Health check (retorna 200 aunque los servicios estén desconectados).

**Respuesta 200:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-01T00:00:00.000Z",
  "services": {
    "postgres": "connected",
    "mongodb": "disconnected",
    "redis": "disconnected"
  }
}
```

---

### GET /api/categories

Listar categorías de cursos publicados.

**Respuesta 200:**
```json
{
  "data": ["programación", "diseño", "marketing"]
}
```

---

## Códigos de Error

| Código | Significado |
|--------|-------------|
| 400 | Bad Request — datos inválidos en body/query |
| 401 | Unauthorized — token faltante, inválido o expirado |
| 403 | Forbidden — no tienes permisos para esta acción |
| 404 | Not Found — recurso no existe |
| 409 | Conflict — duplicado (email, inscripción) |
| 429 | Too Many Requests — rate limit excedido |
| 500 | Internal Server Error — error inesperado |

### Mensajes de error comunes

| Mensaje | Causa |
|---------|-------|
| Token de acceso requerido | No se envió el header Authorization |
| Token inválido o expirado | JWT expiró o es inválido |
| Acceso denegado | El rol no tiene permiso para esta ruta |
| Email o contraseña incorrectos | Credenciales inválidas |
| El email ya está registrado | Email duplicado en registro |
| Curso no encontrado | ID o slug inválido |
| Ya estás inscrito en este curso | Intento de inscripción duplicada |
| Demasiadas solicitudes | Rate limit alcanzado (100 req/min por IP) |
