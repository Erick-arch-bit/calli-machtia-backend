/**
 * Handler de subida de imágenes a Cloudinary.
 * Los usuarios autenticados pueden subir imágenes para cursos o perfiles.
 */
import { Hono } from "hono";
import { v2 as cloudinary } from "cloudinary";
import { authMiddleware } from "../middleware/auth";
import { badRequest, internal } from "../lib/errors";
import { config } from "../config";
import type { Context } from "hono";

const upload = new Hono();

/** POST /image — Sube una imagen a Cloudinary y retorna la URL pública y el public_id */
upload.post("/image", authMiddleware, async (c: Context) => {
  if (!config.cloudinaryUrl) {
    throw internal("Cloudinary no está configurado");
  }

  const body = await c.req.parseBody();
  const file = body.file as File | undefined;

  if (!file) {
    throw badRequest("Archivo requerido");
  }

  cloudinary.config({ url: config.cloudinaryUrl });

  const buffer = Buffer.from(await file.arrayBuffer());
  const result = await new Promise<{ secure_url: string; public_id: string }>((resolve, reject) => {
    const stream = cloudinary.uploader.upload_stream(
      { folder: "calli-machtia", resource_type: "image" },
      (error, result) => {
        if (error || !result) reject(error || new Error("Upload failed"));
        else resolve({ secure_url: result.secure_url, public_id: result.public_id });
      }
    );
    stream.end(buffer);
  });

  return c.json({ data: { url: result.secure_url, public_id: result.public_id } });
});

export default upload;
