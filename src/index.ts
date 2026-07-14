/**
 * Punto de entrada del servidor Calli Machtia API.
 * Configura middlewares globales, define rutas principales, conecta servicios
 * (PostgreSQL, MongoDB, Redis) e inicia el servidor HTTP con Bun.
 */
import { Hono } from "hono";
import { HTTPException } from "hono/http-exception";
import { config } from "./config";
import { connectMongoDB, pingMongoDB, connectRedis, pingRedis, pingPostgres, db } from "./db";
import { corsMiddleware, securityHeadersMiddleware, rateLimitMiddleware } from "./middleware";
import {
  authHandler,
  courseHandler,
  enrollmentHandler,
  lessonHandler,
  paymentHandler,
  adminHandler,
  uploadHandler,
  muxHandler,
} from "./handlers";
import { eq } from "drizzle-orm";
import { courses } from "./schema/courses";

const app = new Hono();

app.use("*", corsMiddleware);
app.use("*", securityHeadersMiddleware);
app.use("*", rateLimitMiddleware);

/** Manejador global de errores. Convierte HTTPException en respuestas JSON estructuradas. */
app.onError((err, c) => {
  if (err instanceof HTTPException) {
    return c.json({ error: err.message }, err.status);
  }

  console.error("Unhandled error:", err);
  return c.json({ error: "Error interno del servidor" }, 500);
});

/** Ruta raíz: información de la API */
app.get("/", (c) => {
  return c.json({
    name: "Calli Machtia API",
    version: "1.0.0",
    environment: config.env,
  });
});

/** Health check: verifica el estado de conexión de PostgreSQL, MongoDB y Redis */
app.get("/health", async (c) => {
  const pg = await pingPostgres().catch(() => false);
  const mongo = await pingMongoDB().catch(() => false);
  const redis = await pingRedis().catch(() => false);

  return c.json({
    status: "healthy",
    timestamp: new Date().toISOString(),
    services: {
      postgres: pg ? "connected" : "disconnected",
      mongodb: mongo ? "connected" : "disconnected",
      redis: redis ? "connected" : "disconnected",
    },
  });
});

/** Obtiene la lista de categorías de cursos publicados */
app.get("/api/categories", async (c) => {
  const result = await db
    .select({ category: courses.category })
    .from(courses)
    .where(eq(courses.published, true))
    .groupBy(courses.category);

  const categories = result
    .map((r) => r.category)
    .filter((c): c is string => c !== null);

  return c.json({ data: categories });
});

app.route("/api/auth", authHandler);
app.route("/api/courses", courseHandler);
app.route("/api/courses", lessonHandler);
app.route("/api/enrollments", enrollmentHandler);
app.route("/api/payments", paymentHandler);
app.route("/api/admin", adminHandler);
app.route("/api/uploads", uploadHandler);
app.route("/api/uploads", muxHandler);

/**
 * Inicializa los servicios y arranca el servidor HTTP.
 * En entorno de test omite la conexión a MongoDB y Redis.
 * Si el puerto configurado está ocupado, intenta con el siguiente.
 */
async function start() {
  if (config.env !== "test") {
    await connectMongoDB();
    await connectRedis();
  }

  console.log(`Server starting on port ${config.port} in ${config.env} mode`);
  console.log(`DATABASE_URL set: ${!!process.env.DATABASE_URL}`);
  console.log(`MONGODB_URI set: ${!!process.env.MONGODB_URI}`);
  console.log(`REDIS_URL set: ${!!process.env.REDIS_URL}`);

  try {
    const server = Bun.serve({
      port: config.port,
      hostname: "0.0.0.0",
      fetch: app.fetch,
    });
    console.log(`Server running on ${server.hostname}:${server.port}`);
  } catch (err: any) {
    if (err.code === "EADDRINUSE") {
      const fallback = parseInt(config.port) + 1;
      console.log(`Port ${config.port} in use, trying ${fallback}`);
      const server = Bun.serve({
        port: fallback,
        hostname: "0.0.0.0",
        fetch: app.fetch,
      });
      console.log(`Server running on ${server.hostname}:${server.port} (fallback)`);
    } else {
      throw err;
    }
  }
}

start().catch((err) => {
  console.error("Failed to start server:", err);
  process.exit(1);
});

export default app;
