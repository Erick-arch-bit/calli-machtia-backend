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

export async function pingRedis(): Promise<boolean> {
  if (!redis) return false;
  try {
    await redis.ping();
    return true;
  } catch {
    return false;
  }
}

export function getRedis(): Redis | null {
  return redis;
}
