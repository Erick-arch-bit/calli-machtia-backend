/**
 * Clases y funciones helper para errores HTTP.
 * Extiende HTTPException de Hono con mensajes en español.
 */
import { HTTPException } from "hono/http-exception";

/** Error genérico de aplicación que extiende HTTPException de Hono */
export class AppError extends HTTPException {
  constructor(status: number, message: string) {
    super(status as any, { message });
  }
}

/** Crea un error 400 Bad Request */
export function badRequest(message: string): AppError {
  return new AppError(400, message);
}

/** Crea un error 401 Unauthorized */
export function unauthorized(message = "No autorizado"): AppError {
  return new AppError(401, message);
}

/** Crea un error 403 Forbidden */
export function forbidden(message = "Acceso denegado"): AppError {
  return new AppError(403, message);
}

/** Crea un error 404 Not Found */
export function notFound(message = "Recurso no encontrado"): AppError {
  return new AppError(404, message);
}

/** Crea un error 409 Conflict */
export function conflict(message: string): AppError {
  return new AppError(409, message);
}

/** Crea un error 429 Too Many Requests */
export function tooMany(message = "Demasiadas solicitudes"): AppError {
  return new AppError(429, message);
}

/** Crea un error 500 Internal Server Error */
export function internal(message = "Error interno del servidor"): AppError {
  return new AppError(500, message);
}
