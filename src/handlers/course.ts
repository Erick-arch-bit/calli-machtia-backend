import { Hono } from "hono";
import { eq, and, like, or, sql, count } from "drizzle-orm";
import { z } from "zod";
import { v4 as uuidv4 } from "uuid";
import { db } from "../db/postgres";
import { courses } from "../schema/courses";
import { users } from "../schema/users";
import { enrollments } from "../schema/enrollments";
import { authMiddleware, roleMiddleware } from "../middleware/auth";
import { badRequest, notFound, forbidden } from "../lib/errors";
import type { Context } from "hono";

const course = new Hono();

const createCourseSchema = z.object({
  title: z.string().min(1, "El título es requerido").max(200),
  description: z.string().optional(),
  image_url: z.string().max(500).optional(),
  price: z.number().min(0, "El precio debe ser mayor o igual a 0"),
  category: z.string().optional(),
  tags: z.array(z.string()).optional(),
  published: z.boolean().optional(),
  seo_title: z.string().max(200).optional(),
  seo_description: z.string().max(500).optional(),
});

const updateCourseSchema = z.object({
  title: z.string().min(1).max(200).optional(),
  description: z.string().optional(),
  image_url: z.string().max(500).optional(),
  price: z.number().min(0).optional(),
  category: z.string().optional(),
  tags: z.array(z.string()).optional(),
  published: z.boolean().optional(),
  seo_title: z.string().max(200).optional(),
  seo_description: z.string().max(500).optional(),
});

function generateSlug(text: string): string {
  const base = text
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9-]/g, "")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
  const suffix = Math.random().toString(36).substring(2, 8);
  return `${base}-${suffix}`;
}

course.get("/instructor/mine", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;

  const result = await db
    .select()
    .from(courses)
    .where(eq(courses.instructor_id, userId))
    .orderBy(courses.created_at);

  return c.json({ data: result });
});

course.get("/", async (c: Context) => {
  const query = c.req.query();
  const page = Math.max(1, parseInt(query.page || "1"));
  const limit = Math.min(100, Math.max(1, parseInt(query.limit || "10")));
  const category = query.category;
  const search = query.search;
  const offset = (page - 1) * limit;

  const conditions = [eq(courses.published, true)];

  if (category) {
    conditions.push(eq(courses.category, category));
  }

  if (search) {
    const searchCondition = or(
      like(courses.title, `%${search}%`),
      like(sql`COALESCE(${courses.description}, '')`, `%${search}%`)
    );
    if (searchCondition) conditions.push(searchCondition);
  }

  const totalResult = await db
    .select({ count: count() })
    .from(courses)
    .where(and(...conditions));

  const total = totalResult[0]?.count ?? 0;

  const result = await db
    .select({
      id: courses.id,
      instructor_id: courses.instructor_id,
      title: courses.title,
      slug: courses.slug,
      description: courses.description,
      image_url: courses.image_url,
      price: courses.price,
      category: courses.category,
      tags: courses.tags,
      published: courses.published,
      seo_title: courses.seo_title,
      seo_description: courses.seo_description,
      created_at: courses.created_at,
      updated_at: courses.updated_at,
      instructor_name: users.name,
      enrollment_count: sql<number>`(
        SELECT COUNT(*)::int FROM ${enrollments} WHERE ${enrollments.course_id} = ${courses.id}
      )`,
    })
    .from(courses)
    .leftJoin(users, eq(courses.instructor_id, users.id))
    .where(and(...conditions))
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

course.get("/slug/:slug", async (c: Context) => {
  const { slug } = c.req.param();

  const result = await db
    .select({
      id: courses.id,
      instructor_id: courses.instructor_id,
      title: courses.title,
      slug: courses.slug,
      description: courses.description,
      image_url: courses.image_url,
      price: courses.price,
      category: courses.category,
      tags: courses.tags,
      published: courses.published,
      seo_title: courses.seo_title,
      seo_description: courses.seo_description,
      created_at: courses.created_at,
      updated_at: courses.updated_at,
      instructor_name: users.name,
      enrollment_count: sql<number>`(
        SELECT COUNT(*)::int FROM ${enrollments} WHERE ${enrollments.course_id} = ${courses.id}
      )`,
    })
    .from(courses)
    .leftJoin(users, eq(courses.instructor_id, users.id))
    .where(eq(courses.slug, slug))
    .limit(1);

  if (result.length === 0) {
    throw notFound("Curso no encontrado");
  }

  return c.json({ data: result[0] });
});

course.get("/:id", async (c: Context) => {
  const { id } = c.req.param();

  const result = await db
    .select({
      id: courses.id,
      instructor_id: courses.instructor_id,
      title: courses.title,
      slug: courses.slug,
      description: courses.description,
      image_url: courses.image_url,
      price: courses.price,
      category: courses.category,
      tags: courses.tags,
      published: courses.published,
      seo_title: courses.seo_title,
      seo_description: courses.seo_description,
      created_at: courses.created_at,
      updated_at: courses.updated_at,
      instructor_name: users.name,
      enrollment_count: sql<number>`(
        SELECT COUNT(*)::int FROM ${enrollments} WHERE ${enrollments.course_id} = ${courses.id}
      )`,
    })
    .from(courses)
    .leftJoin(users, eq(courses.instructor_id, users.id))
    .where(eq(courses.id, id))
    .limit(1);

  if (result.length === 0) {
    throw notFound("Curso no encontrado");
  }

  return c.json({ data: result[0] });
});

course.post("/", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const body = await c.req.json();
  const parsed = createCourseSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const { title, description, image_url, price, category, tags, published, seo_title, seo_description } = parsed.data;
  const slug = generateSlug(title);
  const courseId = uuidv4();

  await db.insert(courses).values({
    id: courseId,
    instructor_id: userId,
    title,
    slug,
    description: description || null,
    image_url: image_url || null,
    price: price.toString(),
    category: category || null,
    tags: tags || [],
    published: published || false,
    seo_title: seo_title || null,
    seo_description: seo_description || null,
  });

  const created = await db.select().from(courses).where(eq(courses.id, courseId)).limit(1);

  return c.json({ data: created[0] }, 201);
});

course.put("/:id", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id } = c.req.param();
  const body = await c.req.json();
  const parsed = updateCourseSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const existing = await db.select().from(courses).where(eq(courses.id, id)).limit(1);
  if (existing.length === 0) {
    throw notFound("Curso no encontrado");
  }

  if (role !== "admin" && existing[0].instructor_id !== userId) {
    throw forbidden("No tienes permiso para modificar este curso");
  }

  const updateData: Record<string, unknown> = {};
  if (parsed.data.title !== undefined) {
    updateData.title = parsed.data.title;
    updateData.slug = generateSlug(parsed.data.title);
  }
  if (parsed.data.description !== undefined) updateData.description = parsed.data.description;
  if (parsed.data.image_url !== undefined) updateData.image_url = parsed.data.image_url;
  if (parsed.data.price !== undefined) updateData.price = parsed.data.price.toString();
  if (parsed.data.category !== undefined) updateData.category = parsed.data.category;
  if (parsed.data.tags !== undefined) updateData.tags = parsed.data.tags;
  if (parsed.data.published !== undefined) updateData.published = parsed.data.published;
  if (parsed.data.seo_title !== undefined) updateData.seo_title = parsed.data.seo_title;
  if (parsed.data.seo_description !== undefined) updateData.seo_description = parsed.data.seo_description;

  await db.update(courses).set(updateData).where(eq(courses.id, id));

  const updated = await db.select().from(courses).where(eq(courses.id, id)).limit(1);

  return c.json({ data: updated[0] });
});

course.delete("/:id", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id } = c.req.param();

  const existing = await db.select().from(courses).where(eq(courses.id, id)).limit(1);
  if (existing.length === 0) {
    throw notFound("Curso no encontrado");
  }

  if (role !== "admin" && existing[0].instructor_id !== userId) {
    throw forbidden("No tienes permiso para eliminar este curso");
  }

  await db.delete(courses).where(eq(courses.id, id));

  return c.json({ data: { message: "Curso eliminado exitosamente" } });
});

export default course;
