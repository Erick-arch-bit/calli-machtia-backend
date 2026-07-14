/**
 * Script de migración de base de datos.
 * Crea las tablas necesarias (users, courses, enrollments, payments, password_resets)
 * y los índices correspondientes en PostgreSQL.
 */
import { drizzle } from "drizzle-orm/node-postgres";
import pg from "pg";
import { config } from "../config";

const { Pool } = pg;

async function migrate() {
  const pool = new Pool({ connectionString: config.databaseUrl });

  const client = await pool.connect();

  try {
    await client.query(`CREATE EXTENSION IF NOT EXISTS pgcrypto`);
    await client.query(`CREATE TABLE IF NOT EXISTS users (
      id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
      name TEXT NOT NULL,
      email TEXT NOT NULL UNIQUE,
      password_hash TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'alumno',
      avatar_url TEXT,
      bio TEXT,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
      updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
    )`);
    await client.query(`CREATE TABLE IF NOT EXISTS courses (
      id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
      instructor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      slug TEXT NOT NULL UNIQUE,
      description TEXT,
      image_url TEXT,
      price DECIMAL(10,2) NOT NULL DEFAULT 0,
      category TEXT,
      tags TEXT[] NOT NULL DEFAULT '{}',
      published BOOLEAN DEFAULT false,
      seo_title TEXT,
      seo_description TEXT,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
      updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
    )`);
    await client.query(`CREATE TABLE IF NOT EXISTS enrollments (
      id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
      status TEXT NOT NULL DEFAULT 'active',
      progress DECIMAL(5,2) DEFAULT 0,
      enrolled_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
      completed_at TIMESTAMPTZ,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
      updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
      UNIQUE(user_id, course_id)
    )`);
    await client.query(`CREATE TABLE IF NOT EXISTS payments (
      id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
      amount DECIMAL(10,2) NOT NULL,
      currency TEXT NOT NULL DEFAULT 'usd',
      stripe_payment_id TEXT,
      status TEXT NOT NULL DEFAULT 'pending',
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
    )`);
    await client.query(`CREATE TABLE IF NOT EXISTS password_resets (
      id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
      email TEXT NOT NULL,
      token TEXT NOT NULL,
      expires_at TIMESTAMPTZ NOT NULL,
      used BOOLEAN DEFAULT false NOT NULL,
      created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
    )`);

    await client.query("CREATE INDEX IF NOT EXISTS idx_courses_instructor_id ON courses(instructor_id)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_courses_slug ON courses(slug)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_courses_category ON courses(category) WHERE published = true");
    await client.query("CREATE INDEX IF NOT EXISTS idx_courses_published ON courses(published) WHERE published = true");
    await client.query("CREATE INDEX IF NOT EXISTS idx_enrollments_user_id ON enrollments(user_id)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments(course_id)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_payments_course_id ON payments(course_id)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token)");
    await client.query("CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email)");

    console.log("Migration completed successfully");
  } finally {
    client.release();
    await pool.end();
  }
}

migrate().catch((err) => {
  console.error("Migration failed:", err);
  process.exit(1);
});
