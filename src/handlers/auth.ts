import { Hono } from "hono";
import bcrypt from "bcryptjs";
import crypto from "crypto";
import { eq, and } from "drizzle-orm";
import { z } from "zod";
import { v4 as uuidv4 } from "uuid";
import { db } from "../db/postgres";
import { getRedis } from "../db/redis";
import { users } from "../schema/users";
import { passwordResets } from "../schema/password-resets";
import {
  generateAccessToken,
  generateRefreshToken,
  verifyRefreshToken,
  getRefreshTokenTTL,
} from "../lib/jwt";
import { authMiddleware } from "../middleware/auth";
import { badRequest, unauthorized, notFound, conflict } from "../lib/errors";
import type { Context } from "hono";

const auth = new Hono();

const registerSchema = z.object({
  name: z.string().min(1, "El nombre es requerido").max(100),
  email: z.string().email("Email inválido"),
  password: z.string().min(6, "La contraseña debe tener al menos 6 caracteres").max(100),
  role: z.enum(["alumno", "instructor"]).optional().default("alumno"),
});

const loginSchema = z.object({
  email: z.string().email("Email inválido"),
  password: z.string().min(1, "La contraseña es requerida"),
});

const refreshSchema = z.object({
  refresh_token: z.string().min(1, "Refresh token requerido"),
});

const profileSchema = z.object({
  name: z.string().min(1).max(100).optional(),
  avatar_url: z.string().max(500).optional(),
  bio: z.string().max(1000).optional(),
});

const forgotPasswordSchema = z.object({
  email: z.string().email("Email inválido"),
});

const resetPasswordSchema = z.object({
  token: z.string().min(1, "Token requerido"),
  password: z.string().min(6, "La contraseña debe tener al menos 6 caracteres").max(100),
});

function generateSlug(text: string): string {
  return text
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9-]/g, "")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

auth.post("/register", async (c: Context) => {
  const body = await c.req.json();
  const parsed = registerSchema.safeParse(body);
  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { name, email, password, role } = parsed.data;

  const existing = await db.select().from(users).where(eq(users.email, email)).limit(1);
  if (existing.length > 0) {
    throw conflict("El email ya está registrado");
  }

  const password_hash = await bcrypt.hash(password, 12);
  const userId = uuidv4();

  await db.insert(users).values({
    id: userId,
    name,
    email,
    password_hash,
    role,
  });

  const user = { id: userId, name, email, role };

  const accessToken = generateAccessToken({ user_id: userId, role });
  const refreshToken = generateRefreshToken({ user_id: userId });

  const redis = getRedis();
  if (redis) {
    try {
      await redis.set(`refresh_token:${userId}`, refreshToken, "EX", getRefreshTokenTTL());
    } catch { /* graceful degradation */ }
  }

  return c.json({ data: { user, accessToken, refreshToken } }, 201);
});

auth.post("/login", async (c: Context) => {
  const body = await c.req.json();
  const parsed = loginSchema.safeParse(body);
  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { email, password } = parsed.data;

  const result = await db.select().from(users).where(eq(users.email, email)).limit(1);
  if (result.length === 0) {
    throw unauthorized("Email o contraseña incorrectos");
  }

  const user = result[0];
  const valid = await bcrypt.compare(password, user.password_hash);
  if (!valid) {
    throw unauthorized("Email o contraseña incorrectos");
  }

  const userData = {
    id: user.id,
    name: user.name,
    email: user.email,
    role: user.role,
    avatar_url: user.avatar_url,
    bio: user.bio,
    created_at: user.created_at,
  };

  const accessToken = generateAccessToken({ user_id: user.id, role: user.role });
  const refreshToken = generateRefreshToken({ user_id: user.id });

  const redis = getRedis();
  if (redis) {
    try {
      await redis.set(`refresh_token:${user.id}`, refreshToken, "EX", getRefreshTokenTTL());
    } catch { /* graceful degradation */ }
  }

  return c.json({ data: { user: userData, accessToken, refreshToken } });
});

auth.get("/me", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;

  const result = await db.select().from(users).where(eq(users.id, userId)).limit(1);
  if (result.length === 0) {
    throw notFound("Usuario no encontrado");
  }

  const user = result[0];
  return c.json({
    data: {
      id: user.id,
      name: user.name,
      email: user.email,
      role: user.role,
      avatar_url: user.avatar_url,
      bio: user.bio,
      created_at: user.created_at,
    },
  });
});

auth.post("/refresh", async (c: Context) => {
  const body = await c.req.json();
  const parsed = refreshSchema.safeParse(body);
  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { refresh_token } = parsed.data;

  let payload;
  try {
    payload = verifyRefreshToken(refresh_token);
  } catch {
    throw unauthorized("Refresh token inválido o expirado");
  }

  const redis = getRedis();
  if (redis) {
    try {
      const stored = await redis.get(`refresh_token:${payload.user_id}`);
      if (stored !== refresh_token) {
        throw unauthorized("Refresh token inválido o expirado");
      }
    } catch {
      // If Redis is unavailable, fall back to just JWT verification
    }
  }

  const result = await db.select().from(users).where(eq(users.id, payload.user_id)).limit(1);
  if (result.length === 0) {
    throw unauthorized("Usuario no encontrado");
  }

  const accessToken = generateAccessToken({ user_id: payload.user_id, role: result[0].role });

  return c.json({ data: { accessToken } });
});

auth.put("/profile", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;
  const body = await c.req.json();
  const parsed = profileSchema.safeParse(body);
  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const updateData: Record<string, string> = {};
  if (parsed.data.name !== undefined) updateData.name = parsed.data.name;
  if (parsed.data.avatar_url !== undefined) updateData.avatar_url = parsed.data.avatar_url;
  if (parsed.data.bio !== undefined) updateData.bio = parsed.data.bio;

  if (Object.keys(updateData).length === 0) {
    throw badRequest("No hay datos para actualizar");
  }

  await db.update(users).set(updateData).where(eq(users.id, userId));

  const result = await db.select().from(users).where(eq(users.id, userId)).limit(1);
  if (result.length === 0) {
    throw notFound("Usuario no encontrado");
  }

  const user = result[0];
  return c.json({
    data: {
      id: user.id,
      name: user.name,
      email: user.email,
      role: user.role,
      avatar_url: user.avatar_url,
      bio: user.bio,
      created_at: user.created_at,
    },
  });
});

auth.post("/logout", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;

  const redis = getRedis();
  if (redis) {
    try {
      await redis.del(`refresh_token:${userId}`);
    } catch { /* graceful degradation */ }
  }

  return c.json({ data: { message: "Sesión cerrada exitosamente" } });
});

auth.post("/forgot-password", async (c: Context) => {
  const body = await c.req.json();
  const parsed = forgotPasswordSchema.safeParse(body);
  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { email } = parsed.data;

  const result = await db.select().from(users).where(eq(users.email, email)).limit(1);
  if (result.length === 0) {
    return c.json({ data: { message: "Si el email existe, recibirás un enlace de recuperación" } });
  }

  const token = crypto.randomBytes(32).toString("hex");
  const expiresAt = new Date(Date.now() + 3600000); // 1 hour

  await db.insert(passwordResets).values({
    id: uuidv4(),
    email,
    token,
    expires_at: expiresAt,
  });

  // In production, send email here
  console.log(`Password reset token for ${email}: ${token}`);

  return c.json({ data: { message: "Si el email existe, recibirás un enlace de recuperación" } });
});

auth.post("/reset-password", async (c: Context) => {
  const body = await c.req.json();
  const parsed = resetPasswordSchema.safeParse(body);
  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { token, password } = parsed.data;

  const resetRecord = await db
    .select()
    .from(passwordResets)
    .where(
      and(
        eq(passwordResets.token, token),
        eq(passwordResets.used, false)
      )
    )
    .limit(1);

  if (resetRecord.length === 0) {
    throw badRequest("Token inválido o ya utilizado");
  }

  const record = resetRecord[0];
  if (new Date() > record.expires_at) {
    throw badRequest("Token expirado");
  }

  const password_hash = await bcrypt.hash(password, 12);

  await db
    .update(users)
    .set({ password_hash })
    .where(eq(users.email, record.email));

  await db
    .update(passwordResets)
    .set({ used: true })
    .where(eq(passwordResets.id, record.id));

  return c.json({ data: { message: "Contraseña actualizada exitosamente" } });
});

export default auth;
