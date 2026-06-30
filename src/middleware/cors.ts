import type { Context, Next } from "hono";
import { config } from "../config";

export async function corsMiddleware(c: Context, next: Next) {
  const origin = c.req.header("Origin") || "";
  const allowed = config.corsOrigin.includes(origin) || config.corsOrigin.includes("*");

  if (allowed) {
    c.header("Access-Control-Allow-Origin", origin);
    c.header("Access-Control-Allow-Credentials", "true");
  } else if (config.env === "development") {
    c.header("Access-Control-Allow-Origin", origin || "*");
    c.header("Access-Control-Allow-Credentials", "true");
  }

  c.header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS");
  c.header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With");
  c.header("Access-Control-Max-Age", "86400");

  if (c.req.method === "OPTIONS") {
    return c.body(null, 204);
  }

  await next();
}
