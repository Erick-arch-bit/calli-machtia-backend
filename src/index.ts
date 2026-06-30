import { Hono } from "hono";
import { HTTPException } from "hono/http-exception";
import { config } from "./config";
import { connectMongoDB, pingMongoDB } from "./db/mongodb";
import { connectRedis, pingRedis } from "./db/redis";
import { pingPostgres } from "./db/postgres";
import { corsMiddleware } from "./middleware/cors";
import { securityHeadersMiddleware } from "./middleware/security";
import { rateLimitMiddleware } from "./middleware/ratelimit";
import authHandler from "./handlers/auth";
import courseHandler from "./handlers/course";
import enrollmentHandler from "./handlers/enrollment";
import lessonHandler from "./handlers/lesson";
import paymentHandler from "./handlers/payment";
import adminHandler from "./handlers/admin";
import { eq } from "drizzle-orm";
import { db } from "./db/postgres";
import { courses } from "./schema/courses";

const app = new Hono();

// Global middleware
app.use("*", corsMiddleware);
app.use("*", securityHeadersMiddleware);
app.use("*", rateLimitMiddleware);

// Error handler
app.onError((err, c) => {
  if (err instanceof HTTPException) {
    return c.json({ error: err.message }, err.status);
  }

  console.error("Unhandled error:", err);
  return c.json({ error: "Error interno del servidor" }, 500);
});

// API info
app.get("/", (c) => {
  return c.json({
    name: "Calli Machtia API",
    version: "1.0.0",
    environment: config.env,
  });
});

// Health check
app.get("/health", async (c) => {
  const pg = await pingPostgres();
  const mongo = await pingMongoDB();
  const redis = await pingRedis();

  const healthy = pg;

  return c.json(
    {
      status: healthy ? "healthy" : "unhealthy",
      timestamp: new Date().toISOString(),
      services: {
        postgres: pg ? "connected" : "disconnected",
        mongodb: mongo ? "connected" : "disconnected",
        redis: redis ? "connected" : "disconnected",
      },
    },
    healthy ? 200 : 503
  );
});

// Categories
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

// Mount routes
app.route("/api/auth", authHandler);
app.route("/api/courses", courseHandler);
app.route("/api/courses", lessonHandler);
app.route("/api/enrollments", enrollmentHandler);
app.route("/api/payments", paymentHandler);
app.route("/api/admin", adminHandler);

// Connect services and start server
async function start() {
  if (config.env !== "test") {
    await connectMongoDB();
    await connectRedis();
  }

  console.log(`Server starting on port ${config.port} in ${config.env} mode`);

  Bun.serve({
    port: config.port,
    fetch: app.fetch,
  });
}

start().catch((err) => {
  console.error("Failed to start server:", err);
  process.exit(1);
});

export default app;
