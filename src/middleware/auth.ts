import type { Context, Next } from "hono";
import { verifyAccessToken } from "../lib/jwt";
import { unauthorized } from "../lib/errors";

export async function authMiddleware(c: Context, next: Next) {
  const authHeader = c.req.header("Authorization");
  if (!authHeader || !authHeader.startsWith("Bearer ")) {
    throw unauthorized("Token de acceso requerido");
  }

  const token = authHeader.split(" ")[1];
  try {
    const payload = verifyAccessToken(token);
    c.set("user_id", payload.user_id);
    c.set("role", payload.role);
    await next();
  } catch {
    throw unauthorized("Token inválido o expirado");
  }
}

export function roleMiddleware(...roles: string[]) {
  return async (c: Context, next: Next) => {
    const role = c.get("role") as string;
    if (!role || !roles.includes(role)) {
      throw unauthorized("Acceso denegado");
    }
    await next();
  };
}
