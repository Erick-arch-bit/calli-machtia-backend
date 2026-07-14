/**
 * Handler de integración con Mux (video).
 * Genera URLs de subida para lecciones y procesa webhooks
 * cuando los videos están listos para reproducción.
 */
import { Hono } from "hono";
import Mux from "@mux/mux-node";
import { eq } from "drizzle-orm";
import { db } from "../db/postgres";
import { courses } from "../schema/courses";
import { connectMongoDB, ModuleModel } from "../db/mongodb";
import { authMiddleware } from "../middleware/auth";
import { badRequest, notFound, internal } from "../lib/errors";
import { config } from "../config";
import type { Context } from "hono";

/** Inicializa el cliente de Mux verificando que las credenciales existan */
function getMux() {
  if (!config.muxTokenId || !config.muxTokenSecret) {
    throw internal("Mux no está configurado");
  }
  return new Mux({
    tokenId: config.muxTokenId,
    tokenSecret: config.muxTokenSecret,
    webhookSecret: config.muxWebhookSecret,
  });
}

const mux = new Hono();

/** POST /upload-url — Genera una URL de subida firmada de Mux para una lección */
mux.post("/upload-url", authMiddleware, async (c: Context) => {
  const { courseId, lessonId } = await c.req.json();

  if (!courseId || !lessonId) {
    throw badRequest("courseId y lessonId son requeridos");
  }

  const [course] = await db
    .select({ id: courses.id })
    .from(courses)
    .where(eq(courses.id, courseId))
    .limit(1);

  if (!course) {
    throw notFound("Curso no encontrado");
  }

  await connectMongoDB();
  const modulo = await ModuleModel.findOne(
    { "lessons.id": lessonId },
    { "lessons.$": 1 }
  );

  if (!modulo) {
    throw notFound("Lección no encontrada");
  }

  try {
    const upload = await getMux().video.uploads.create({
      timeout: 3600,
      cors_origin: config.frontendUrl,
      new_asset_settings: {
        playback_policy: ["public"],
        passthrough: JSON.stringify({ courseId, lessonId }),
      },
    });

    return c.json({ data: { url: upload.url, upload_id: upload.id } });
  } catch (err) {
    console.error("Mux upload error:", err);
    throw internal("Error al crear upload en Mux");
  }
});

/** POST /webhook — Recibe eventos de Mux y actualiza la lección con playback_id cuando el video está listo */
mux.post("/webhook", async (c: Context) => {
  try {
    const body = await c.req.text();
    const headers = Object.fromEntries(c.req.raw.headers.entries());

    const event = await getMux().webhooks.unwrap(body, headers as any);

    if (event.type !== "video.asset.ready") {
      return c.json({ received: true });
    }

    const data = event.data as any;
    const assetId = data.id;
    const playbackId = data.playback_ids?.[0]?.id;

    let courseId: string;
    let lessonId: string;
    try {
      const parsed = JSON.parse(data.passthrough ?? "{}");
      courseId = parsed.courseId;
      lessonId = parsed.lessonId;
    } catch {
      return c.json({ error: "passthrough inválido" }, 400);
    }

    if (!playbackId || !lessonId || !courseId) {
      console.warn("Mux webhook: datos incompletos");
      return c.json({ received: true });
    }

    await connectMongoDB();
    await ModuleModel.updateOne(
      { course_id: courseId },
      {
        $set: {
          "lessons.$[elem].mux_playback_id": playbackId,
          "lessons.$[elem].mux_asset_id": assetId,
        },
      },
      { arrayFilters: [{ "elem.id": lessonId }] }
    );

    return c.json({ received: true });
  } catch (err) {
    console.error("Mux webhook error:", err);
    return c.json({ received: true }, 200);
  }
});

export default mux;
