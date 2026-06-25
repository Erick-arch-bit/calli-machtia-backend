package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(pool *pgxpool.Pool) {
	ctx := context.Background()

	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "001_create_users",
			sql: `CREATE TABLE IF NOT EXISTS users (
				id UUID PRIMARY KEY,
				name TEXT NOT NULL,
				email TEXT NOT NULL UNIQUE,
				password_hash TEXT NOT NULL,
				role TEXT NOT NULL DEFAULT 'alumno',
				avatar_url TEXT,
				bio TEXT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
		},
		{
			name: "002_create_courses",
			sql: `CREATE TABLE IF NOT EXISTS courses (
				id UUID PRIMARY KEY,
				instructor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				title TEXT NOT NULL,
				slug TEXT NOT NULL UNIQUE,
				description TEXT,
				image_url TEXT,
				price DECIMAL(10,2) NOT NULL DEFAULT 0,
				category TEXT,
				tags TEXT[] DEFAULT '{}',
				published BOOLEAN NOT NULL DEFAULT false,
				seo_title TEXT,
				seo_description TEXT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
		},
		{
			name: "003_create_enrollments",
			sql: `CREATE TABLE IF NOT EXISTS enrollments (
				id UUID PRIMARY KEY,
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
				status TEXT NOT NULL DEFAULT 'active',
				progress DECIMAL(5,2) NOT NULL DEFAULT 0,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				UNIQUE(user_id, course_id)
			)`,
		},
		{
			name: "004_create_payments",
			sql: `CREATE TABLE IF NOT EXISTS payments (
				id UUID PRIMARY KEY,
				user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
				course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
				amount DECIMAL(10,2) NOT NULL,
				currency TEXT NOT NULL DEFAULT 'usd',
				status TEXT NOT NULL DEFAULT 'pending',
				stripe_payment_intent_id TEXT,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
		},
		{
			name: "005_create_password_resets",
			sql: `CREATE TABLE IF NOT EXISTS password_resets (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email TEXT NOT NULL,
				token TEXT NOT NULL,
				expires_at TIMESTAMPTZ NOT NULL,
				used BOOLEAN NOT NULL DEFAULT false,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			)`,
		},
		{
			name: "006_create_indexes",
			sql: `CREATE INDEX IF NOT EXISTS idx_courses_instructor_id ON courses(instructor_id);
				CREATE INDEX IF NOT EXISTS idx_courses_slug ON courses(slug);
				CREATE INDEX IF NOT EXISTS idx_courses_category ON courses(category);
				CREATE INDEX IF NOT EXISTS idx_courses_published ON courses(published);
				CREATE INDEX IF NOT EXISTS idx_enrollments_user_id ON enrollments(user_id);
				CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments(course_id);
				CREATE INDEX IF NOT EXISTS idx_enrollments_user_course ON enrollments(user_id, course_id);
				CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
				CREATE INDEX IF NOT EXISTS idx_payments_course_id ON payments(course_id);
				CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
				CREATE INDEX IF NOT EXISTS idx_password_resets_email ON password_resets(email)`,
		},
	}

	for _, m := range migrations {
		_, err := pool.Exec(ctx, m.sql)
		if err != nil {
			log.Fatalf("Migration %s failed: %v", m.name, err)
		}
		fmt.Printf("Migration %s applied\n", m.name)
	}
}
