/**
 * Middleware de rate limiting basado en Redis.
 * Limita el número de solicitudes por IP en una ventana de 60 segundos.
 * Si Redis no está disponible, permite el paso (degradación graceful).
 */
import type { Context, Next } from "hono";
import { config } from "../config";
import { getRedis } from "../db/redis";

/** Middleware que limita solicitudes por IP usando un contador en Redis */
export async function rateLimitMiddleware(c: Context, next: Next) {
  const redis = getRedis();
  if (!redis) {
    await next();
    return;
  }

  const ip = c.req.header("x-forwarded-for")?.split(",")[0]?.trim() ||
    c.req.header("x-real-ip") ||
    "unknown";

  const key = `ratelimit:${ip}`;

  try {
    const current = await redis.incr(key);
    if (current === 1) {
      await redis.expire(key, 60);
    }

    if (current > config.rateLimit) {
      c.header("Retry-After", "60");
      return c.json({ error: "Demasiadas solicitudes. Intente de nuevo en 60 segundos." }, 429);
    }
  } catch {
    // Degradación graceful: si Redis falla, se permite la solicitud
  }

  await next();
}
