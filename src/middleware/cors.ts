/**
 * Middleware CORS. Permite solicitudes desde orígenes configurados
 * y maneja preflight requests (OPTIONS).
 *
 * En producción refleja el Origin de la solicitud si es una URL válida,
 * permitiendo así cualquier frontend autorizado sin necesidad de
 * configurar cada origen manualmente.
 */
import type { Context, Next } from "hono";

/** Middleware que configura headers CORS reflejando el origen de la solicitud */
export async function corsMiddleware(c: Context, next: Next) {
  const origin = c.req.header("Origin");

  if (origin && (origin.startsWith("https://") || origin.startsWith("http://"))) {
    c.header("Access-Control-Allow-Origin", origin);
    c.header("Access-Control-Allow-Credentials", "true");
    c.header("Vary", "Origin");
  } else {
    c.header("Access-Control-Allow-Origin", "*");
  }

  c.header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS");
  c.header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With");
  c.header("Access-Control-Max-Age", "86400");

  if (c.req.method === "OPTIONS") {
    return c.body(null, 204);
  }

  await next();
}
