import { HTTPException } from "hono/http-exception";

export class AppError extends HTTPException {
  constructor(status: number, message: string) {
    super(status as any, { message });
  }
}

export function badRequest(message: string): AppError {
  return new AppError(400, message);
}

export function unauthorized(message = "No autorizado"): AppError {
  return new AppError(401, message);
}

export function forbidden(message = "Acceso denegado"): AppError {
  return new AppError(403, message);
}

export function notFound(message = "Recurso no encontrado"): AppError {
  return new AppError(404, message);
}

export function conflict(message: string): AppError {
  return new AppError(409, message);
}

export function tooMany(message = "Demasiadas solicitudes"): AppError {
  return new AppError(429, message);
}

export function internal(message = "Error interno del servidor"): AppError {
  return new AppError(500, message);
}
