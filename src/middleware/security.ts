/**
 * Middleware de seguridad. Agrega headers HTTP estándar
 * para proteger contra ataques comunes (XSS, clickjacking, MIME sniffing, etc.).
 */
import type { Context, Next } from "hono";

/** Middleware que establece headers de seguridad en todas las respuestas */
export async function securityHeadersMiddleware(c: Context, next: Next) {
  c.header("Strict-Transport-Security", "max-age=31536000; includeSubDomains");
  c.header("X-Content-Type-Options", "nosniff");
  c.header("X-Frame-Options", "DENY");
  c.header("X-XSS-Protection", "1; mode=block");
  c.header("Referrer-Policy", "strict-origin-when-cross-origin");
  await next();
}
