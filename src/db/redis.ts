/**
 * Conexión a Redis mediante ioredis.
 * Se usa para rate limiting y almacenamiento de refresh tokens.
 * Implementa degradación graceful: si Redis no está disponible,
 * la aplicación continúa funcionando con funcionalidad limitada.
 */
import Redis from "ioredis";
import { config } from "../config";

let redis: Redis | null = null;

try {
  redis = new Redis(config.redisUrl, {
    maxRetriesPerRequest: 3,
    retryStrategy(times) {
      if (times > 3) return null;
      return Math.min(times * 200, 2000);
    },
    lazyConnect: true,
  });

  redis.on("error", (err) => {
    console.warn("Redis connection error:", err.message);
  });
} catch {
  console.warn("Redis not available, rate limiting and refresh tokens will be degraded");
}

/** Conecta a Redis (conexión lazy). Si falla, deshabilita Redis para la sesión. */
export async function connectRedis(): Promise<void> {
  if (!redis) return;
  try {
    await redis.connect();
    console.log("Redis connected");
  } catch (err) {
    console.warn("Redis connection failed, running without Redis:", (err as Error).message);
    redis = null;
  }
}

/** Verifica que Redis responde correctamente mediante PING */
export async function pingRedis(): Promise<boolean> {
  if (!redis) return false;
  try {
    await redis.ping();
    return true;
  } catch {
    return false;
  }
}

/** Retorna la instancia de Redis o null si no está disponible */
export function getRedis(): Redis | null {
  return redis;
}
