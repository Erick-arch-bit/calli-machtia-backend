import { Hono } from "hono";
import { eq, like, or, count, sql, and, isNull } from "drizzle-orm";
import { z } from "zod";
import { db } from "../db/postgres";
import { users } from "../schema/users";
import { courses } from "../schema/courses";
import { enrollments } from "../schema/enrollments";
import { payments } from "../schema/payments";
import { authMiddleware, roleMiddleware } from "../middleware/auth";
import { badRequest, notFound } from "../lib/errors";
import type { Context } from "hono";

const admin = new Hono();

// Apply auth + admin to all routes
admin.use("*", authMiddleware, roleMiddleware("admin"));

const updateRoleSchema = z.object({
  role: z.enum(["alumno", "instructor", "admin"], {
    errorMap: () => ({ message: "Rol inválido. Debe ser: alumno, instructor o admin" }),
  }),
});

admin.get("/users", async (c: Context) => {
  const query = c.req.query();
  const page = Math.max(1, parseInt(query.page || "1"));
  const limit = Math.min(100, Math.max(1, parseInt(query.limit || "10")));
  const search = query.search;
  const roleFilter = query.role;
  const offset = (page - 1) * limit;

  const conditions = [];

  if (search) {
    conditions.push(
      or(like(users.name, `%${search}%`), like(users.email, `%${search}%`))
    );
  }

  if (roleFilter && ["alumno", "instructor", "admin"].includes(roleFilter)) {
    conditions.push(eq(users.role, roleFilter));
  }

  const whereClause = conditions.length > 0 ? and(...conditions) : undefined;

  const totalResult = await db
    .select({ count: count() })
    .from(users)
    .where(whereClause);

  const total = totalResult[0]?.count ?? 0;

  const result = await db
    .select({
      id: users.id,
      name: users.name,
      email: users.email,
      role: users.role,
      avatar_url: users.avatar_url,
      bio: users.bio,
      created_at: users.created_at,
      updated_at: users.updated_at,
    })
    .from(users)
    .where(whereClause)
    .orderBy(users.created_at)
    .limit(limit)
    .offset(offset);

  return c.json({
    data: result,
    pagination: {
      page,
      limit,
      total,
      totalPages: Math.ceil(total / limit),
    },
  });
});

admin.put("/users/:id/role", async (c: Context) => {
  const { id } = c.req.param();
  const body = await c.req.json();
  const parsed = updateRoleSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const userData = await db.select().from(users).where(eq(users.id, id)).limit(1);
  if (userData.length === 0) {
    throw notFound("Usuario no encontrado");
  }

  await db
    .update(users)
    .set({ role: parsed.data.role })
    .where(eq(users.id, id));

  const updated = await db
    .select({
      id: users.id,
      name: users.name,
      email: users.email,
      role: users.role,
      avatar_url: users.avatar_url,
      bio: users.bio,
      created_at: users.created_at,
    })
    .from(users)
    .where(eq(users.id, id))
    .limit(1);

  return c.json({ data: updated[0] });
});

admin.get("/courses", async (c: Context) => {
  const query = c.req.query();
  const page = Math.max(1, parseInt(query.page || "1"));
  const limit = Math.min(100, Math.max(1, parseInt(query.limit || "10")));
  const offset = (page - 1) * limit;

  const totalResult = await db.select({ count: count() }).from(courses);
  const total = totalResult[0]?.count ?? 0;

  const result = await db
    .select({
      id: courses.id,
      instructor_id: courses.instructor_id,
      title: courses.title,
      slug: courses.slug,
      description: courses.description,
      price: courses.price,
      category: courses.category,
      published: courses.published,
      deleted_at: courses.deleted_at,
      created_at: courses.created_at,
      updated_at: courses.updated_at,
      instructor_name: users.name,
    })
    .from(courses)
    .leftJoin(users, eq(courses.instructor_id, users.id))
    .orderBy(courses.created_at)
    .limit(limit)
    .offset(offset);

  return c.json({
    data: result,
    pagination: {
      page,
      limit,
      total,
      totalPages: Math.ceil(total / limit),
    },
  });
});

admin.get("/stats", async (c: Context) => {
  const totalUsers = await db.select({ count: count() }).from(users);
  const totalCourses = await db.select({ count: count() }).from(courses);
  const totalEnrollments = await db.select({ count: count() }).from(enrollments);

  const revenueResult = await db
    .select({
      total: sql<string>`COALESCE(SUM(CAST(${payments.amount} AS DECIMAL)), 0)`,
    })
    .from(payments)
    .where(eq(payments.status, "completed"));

  return c.json({
    data: {
      total_users: totalUsers[0]?.count ?? 0,
      total_courses: totalCourses[0]?.count ?? 0,
      total_enrollments: totalEnrollments[0]?.count ?? 0,
      total_revenue: parseFloat(revenueResult[0]?.total || "0"),
    },
  });
});

admin.delete("/courses/:id", async (c: Context) => {
  const { id } = c.req.param();

  const courseData = await db.select().from(courses).where(eq(courses.id, id)).limit(1);
  if (courseData.length === 0) {
    throw notFound("Curso no encontrado");
  }

  await db.delete(courses).where(eq(courses.id, id));

  return c.json({ data: { message: "Curso eliminado permanentemente" } });
});

export default admin;
