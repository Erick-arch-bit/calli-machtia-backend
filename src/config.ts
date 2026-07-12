export const config = {
  port: parseInt(process.env.PORT || "8080"),
  databaseUrl: process.env.DATABASE_URL || "postgresql://localhost:5432/calli_machtia",
  mongodbUri: process.env.MONGODB_URI || "mongodb://localhost:27017",
  mongodbDatabase: process.env.MONGODB_DATABASE || "calli_machtia",
  redisUrl: process.env.REDIS_URL || "redis://localhost:6379",
  jwtSecret: (() => {
    if (!process.env.JWT_SECRET && (process.env.ENV === "production" || process.env.NODE_ENV === "production")) {
      throw new Error("JWT_SECRET must be set in production");
    }
    return process.env.JWT_SECRET || "dev-secret-change-in-production";
  })(),
  jwtAccessExpiry: process.env.JWT_ACCESS_EXPIRY || "15m",
  jwtRefreshExpiry: process.env.JWT_REFRESH_EXPIRY || "7d",
  stripeSecretKey: process.env.STRIPE_SECRET_KEY || "",
  stripeWebhookSecret: process.env.STRIPE_WEBHOOK_SECRET || "",
  corsOrigin: (process.env.CORS_ORIGIN || "http://localhost:3000").split(",").map((s) => s.trim()),
  rateLimit: parseInt(process.env.RATE_LIMIT || "100"),
  env: process.env.ENV || "development",
  resendApiKey: process.env.RESEND_API_KEY || "",
  resendFrom: process.env.RESEND_FROM || "onboarding@resend.dev",
  frontendUrl: process.env.FRONTEND_URL || "http://localhost:3000",
  cloudinaryUrl: process.env.CLOUDINARY_URL || "",
  muxTokenId: process.env.MUX_TOKEN_ID || "",
  muxTokenSecret: process.env.MUX_TOKEN_SECRET || "",
  muxWebhookSecret: process.env.MUX_WEBHOOK_SECRET || "",
};
