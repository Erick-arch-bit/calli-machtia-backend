import { Hono } from "hono";
import { eq, and, isNull } from "drizzle-orm";
import { z } from "zod";
import { v4 as uuidv4 } from "uuid";
import { db } from "../db/postgres";
import { courses } from "../schema/courses";
import { ModuleModel } from "../db/mongodb";
import { authMiddleware, roleMiddleware } from "../middleware/auth";
import { badRequest, notFound, forbidden } from "../lib/errors";
import type { Context } from "hono";

const lesson = new Hono();

const moduleSchema = z.object({
  title: z.string().min(1, "El título es requerido").max(200),
  description: z.string().optional(),
  order: z.number().int().min(0).optional(),
});

const lessonSchema = z.object({
  title: z.string().min(1, "El título es requerido").max(200),
  description: z.string().optional(),
  content: z.string().optional(),
  video_url: z.string().optional(),
  duration: z.number().int().min(0).optional(),
  order: z.number().int().min(0).optional(),
  free: z.boolean().optional(),
});

async function verifyCourseAccess(courseId: string, userId: string, role: string): Promise<void> {
  const course = await db
    .select()
    .from(courses)
    .where(and(eq(courses.id, courseId), isNull(courses.deleted_at)))
    .limit(1);

  if (course.length === 0) {
    throw notFound("Curso no encontrado");
  }

  if (role !== "admin" && course[0].instructor_id !== userId) {
    throw forbidden("No tienes permiso para modificar este curso");
  }
}

// GET all modules for a course
lesson.get("/:id/modules", async (c: Context) => {
  const { id } = c.req.param();

  const courseData = await db
    .select()
    .from(courses)
    .where(and(eq(courses.id, id), isNull(courses.deleted_at)))
    .limit(1);

  if (courseData.length === 0) {
    throw notFound("Curso no encontrado");
  }

  const modules = await ModuleModel.find({ course_id: id })
    .sort({ order: 1 })
    .lean()
    .select("-__v");

  return c.json({ data: modules });
});

// GET single module with lessons
lesson.get("/:id/modules/:moduleId", async (c: Context) => {
  const { id, moduleId } = c.req.param();

  const moduleData = await ModuleModel.findOne({
    _id: moduleId,
    course_id: id,
  }).lean().select("-__v");

  if (!moduleData) {
    throw notFound("Módulo no encontrado");
  }

  return c.json({ data: moduleData });
});

// POST create module
lesson.post("/:id/modules", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id } = c.req.param();

  await verifyCourseAccess(id, userId, role);

  const body = await c.req.json();
  const parsed = moduleSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const moduleId = uuidv4();
  const now = new Date();

  const newModule = new ModuleModel({
    _id: moduleId,
    course_id: id,
    title: parsed.data.title,
    description: parsed.data.description || "",
    order: parsed.data.order || 0,
    lessons: [],
    created_at: now,
    updated_at: now,
  });

  await newModule.save();

  return c.json({ data: newModule.toObject() }, 201);
});

// PUT update module
lesson.put("/:id/modules/:moduleId", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id, moduleId } = c.req.param();

  await verifyCourseAccess(id, userId, role);

  const body = await c.req.json();
  const parsed = moduleSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const updateData: Record<string, unknown> = { updated_at: new Date() };
  if (parsed.data.title !== undefined) updateData.title = parsed.data.title;
  if (parsed.data.description !== undefined) updateData.description = parsed.data.description;
  if (parsed.data.order !== undefined) updateData.order = parsed.data.order;

  const updated = await ModuleModel.findOneAndUpdate(
    { _id: moduleId, course_id: id },
    { $set: updateData },
    { new: true }
  ).lean().select("-__v");

  if (!updated) {
    throw notFound("Módulo no encontrado");
  }

  return c.json({ data: updated });
});

// DELETE module
lesson.delete("/:id/modules/:moduleId", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id, moduleId } = c.req.param();

  await verifyCourseAccess(id, userId, role);

  const deleted = await ModuleModel.findOneAndDelete({ _id: moduleId, course_id: id });

  if (!deleted) {
    throw notFound("Módulo no encontrado");
  }

  return c.json({ data: { message: "Módulo eliminado exitosamente" } });
});

// POST add lesson to module
lesson.post("/:id/modules/:moduleId/lessons", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id, moduleId } = c.req.param();

  await verifyCourseAccess(id, userId, role);

  const body = await c.req.json();
  const parsed = lessonSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const lessonId = uuidv4();
  const newLesson = {
    id: lessonId,
    title: parsed.data.title,
    description: parsed.data.description || "",
    content: parsed.data.content || "",
    video_url: parsed.data.video_url || "",
    duration: parsed.data.duration || 0,
    order: parsed.data.order || 0,
    free: parsed.data.free || false,
  };

  const updated = await ModuleModel.findOneAndUpdate(
    { _id: moduleId, course_id: id },
    {
      $push: { lessons: newLesson },
      $set: { updated_at: new Date() },
    },
    { new: true }
  ).lean().select("-__v");

  if (!updated) {
    throw notFound("Módulo no encontrado");
  }

  return c.json({ data: newLesson }, 201);
});

// PUT update lesson
lesson.put("/:id/modules/:moduleId/lessons/:lessonId", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id, moduleId, lessonId } = c.req.param();

  await verifyCourseAccess(id, userId, role);

  const body = await c.req.json();
  const parsed = lessonSchema.safeParse(body);

  if (!parsed.success) {
    throw badRequest(parsed.error.errors[0].message);
  }

  const updateFields: Record<string, unknown> = {};
  if (parsed.data.title !== undefined) updateFields["lessons.$.title"] = parsed.data.title;
  if (parsed.data.description !== undefined) updateFields["lessons.$.description"] = parsed.data.description;
  if (parsed.data.content !== undefined) updateFields["lessons.$.content"] = parsed.data.content;
  if (parsed.data.video_url !== undefined) updateFields["lessons.$.video_url"] = parsed.data.video_url;
  if (parsed.data.duration !== undefined) updateFields["lessons.$.duration"] = parsed.data.duration;
  if (parsed.data.order !== undefined) updateFields["lessons.$.order"] = parsed.data.order;
  if (parsed.data.free !== undefined) updateFields["lessons.$.free"] = parsed.data.free;

  updateFields["updated_at"] = new Date();

  const updated = await ModuleModel.findOneAndUpdate(
    { _id: moduleId, course_id: id, "lessons.id": lessonId },
    { $set: updateFields },
    { new: true }
  ).lean().select("-__v");

  if (!updated) {
    throw notFound("Lección no encontrada");
  }

  const lessonData = (updated.lessons as any[]).find((l: any) => l.id === lessonId);

  return c.json({ data: lessonData });
});

// DELETE lesson
lesson.delete("/:id/modules/:moduleId/lessons/:lessonId", authMiddleware, roleMiddleware("instructor", "admin"), async (c: Context) => {
  const userId = c.get("user_id") as string;
  const role = c.get("role") as string;
  const { id, moduleId, lessonId } = c.req.param();

  await verifyCourseAccess(id, userId, role);

  const updated = await ModuleModel.findOneAndUpdate(
    { _id: moduleId, course_id: id },
    {
      $pull: { lessons: { id: lessonId } },
      $set: { updated_at: new Date() },
    },
    { new: true }
  ).lean().select("-__v");

  if (!updated) {
    throw notFound("Lección no encontrada");
  }

  return c.json({ data: { message: "Lección eliminada exitosamente" } });
});

export default lesson;
