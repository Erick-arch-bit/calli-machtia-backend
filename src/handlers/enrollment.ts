import { Hono } from "hono";
import { eq, and, isNull } from "drizzle-orm";
import { z } from "zod";
import { v4 as uuidv4 } from "uuid";
import { db } from "../db/postgres";
import { enrollments } from "../schema/enrollments";
import { courses } from "../schema/courses";
import { users } from "../schema/users";
import { authMiddleware, roleMiddleware } from "../middleware/auth";
import { badRequest, notFound, conflict, forbidden } from "../lib/errors";
import type { Context } from "hono";

const enrollment = new Hono();

const enrollSchema = z.object({
  course_id: z.string().uuid("ID de curso inválido"),
});

const progressSchema = z.object({
  progress: z.number().min(0).max(100, "El progreso debe estar entre 0 y 100"),
});

enrollment.post("/", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;
  const body = await c.req.json();
  const parsed = enrollSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { course_id } = parsed.data;

  const courseExists = await db
    .select()
    .from(courses)
    .where(and(eq(courses.id, course_id), isNull(courses.deleted_at)))
    .limit(1);

  if (courseExists.length === 0) {
    throw notFound("Curso no encontrado");
  }

  const existing = await db
    .select()
    .from(enrollments)
    .where(and(eq(enrollments.user_id, userId), eq(enrollments.course_id, course_id)))
    .limit(1);

  if (existing.length > 0) {
    throw conflict("Ya estás inscrito en este curso");
  }

  const enrollmentId = uuidv4();
  await db.insert(enrollments).values({
    id: enrollmentId,
    user_id: userId,
    course_id,
  });

  const created = await db
    .select()
    .from(enrollments)
    .where(eq(enrollments.id, enrollmentId))
    .limit(1);

  return c.json({ data: created[0] }, 201);
});

enrollment.get("/mine", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;

  const result = await db
    .select({
      id: enrollments.id,
      user_id: enrollments.user_id,
      course_id: enrollments.course_id,
      status: enrollments.status,
      progress: enrollments.progress,
      created_at: enrollments.created_at,
      updated_at: enrollments.updated_at,
      course: {
        id: courses.id,
        title: courses.title,
        slug: courses.slug,
        description: courses.description,
        image_url: courses.image_url,
        price: courses.price,
        category: courses.category,
        instructor_name: users.name,
      },
    })
    .from(enrollments)
    .leftJoin(courses, eq(enrollments.course_id, courses.id))
    .leftJoin(users, eq(courses.instructor_id, users.id))
    .where(and(eq(enrollments.user_id, userId), isNull(courses.deleted_at)));

  return c.json({ data: result });
});

enrollment.get("/course/:courseId", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const { courseId } = c.req.param();
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;

  const courseData = await db.select().from(courses).where(eq(courses.id, courseId)).limit(1);
  if (courseData.length === 0) {
    throw notFound("Curso no encontrado");
  }

  if (role !== "admin" && courseData[0].instructor_id !== userId) {
    throw forbidden("No tienes permiso para ver las inscripciones de este curso");
  }

  const result = await db
    .select({
      id: enrollments.id,
      user_id: enrollments.user_id,
      course_id: enrollments.course_id,
      status: enrollments.status,
      progress: enrollments.progress,
      created_at: enrollments.created_at,
      updated_at: enrollments.updated_at,
      user: {
        id: users.id,
        name: users.name,
        email: users.email,
        avatar_url: users.avatar_url,
      },
    })
    .from(enrollments)
    .leftJoin(users, eq(enrollments.user_id, users.id))
    .where(eq(enrollments.course_id, courseId));

  return c.json({ data: result });
});

enrollment.put("/:id/progress", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;
  const { id } = c.req.param();
  const body = await c.req.json();
  const parsed = progressSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const enrollmentData = await db
    .select()
    .from(enrollments)
    .where(eq(enrollments.id, id))
    .limit(1);

  if (enrollmentData.length === 0) {
    throw notFound("Inscripción no encontrada");
  }

  if (enrollmentData[0].user_id !== userId) {
    throw forbidden("No tienes permiso para actualizar esta inscripción");
  }

  await db
    .update(enrollments)
    .set({ progress: parsed.data.progress.toString() })
    .where(eq(enrollments.id, id));

  const updated = await db
    .select()
    .from(enrollments)
    .where(eq(enrollments.id, id))
    .limit(1);

  return c.json({ data: updated[0] });
});

enrollment.delete("/:id", authMiddleware, async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id } = c.req.param();

  const enrollmentData = await db
    .select()
    .from(enrollments)
    .where(eq(enrollments.id, id))
    .limit(1);

  if (enrollmentData.length === 0) {
    throw notFound("Inscripción no encontrada");
  }

  if (role !== "admin" && enrollmentData[0].user_id !== userId) {
    throw forbidden("No tienes permiso para eliminar esta inscripción");
  }

  await db.delete(enrollments).where(eq(enrollments.id, id));

  return c.json({ data: { message: "Inscripción cancelada exitosamente" } });
});

export default enrollment;
