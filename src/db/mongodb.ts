/**
 * Conexión a MongoDB mediante Mongoose.
 * Define los schemas de módulos y lecciones para el contenido estructurado
 * de los cursos (almacenados en MongoDB por su naturaleza anidada).
 */
import mongoose from "mongoose";
import { config } from "../config";

/** Conecta a MongoDB con timeout de 5 segundos */
export async function connectMongoDB(): Promise<void> {
  try {
    await mongoose.connect(config.mongodbUri, {
      dbName: config.mongodbDatabase,
      serverSelectionTimeoutMS: 5000,
      connectTimeoutMS: 5000,
    });
    console.log("MongoDB connected");
  } catch (err) {
    console.error("MongoDB connection error:", err);
  }
}

/** Verifica que MongoDB responde correctamente mediante ping admin */
export async function pingMongoDB(): Promise<boolean> {
  try {
    if (mongoose.connection.readyState !== 1) return false;
    await mongoose.connection.db!.admin().ping();
    return true;
  } catch {
    return false;
  }
}

/** Schema de una lección individual dentro de un módulo */
const lessonSchema = new mongoose.Schema(
  {
    id: { type: String },
    title: { type: String, required: true },
    description: { type: String, default: "" },
    content: { type: String, default: "" },
    video_url: { type: String, default: "" },
    duration: { type: Number, default: 0 },
    order: { type: Number, default: 0 },
    free: { type: Boolean, default: false },
    mux_playback_id: { type: String, default: "" },
    mux_asset_id: { type: String, default: "" },
  },
  { _id: false }
);

/** Schema de un módulo que contiene un arreglo de lecciones */
const moduleSchema = new mongoose.Schema(
  {
    _id: { type: String },
    course_id: { type: String, required: true, index: true },
    title: { type: String, required: true },
    description: { type: String, default: "" },
    order: { type: Number, default: 0 },
    lessons: [lessonSchema],
    created_at: { type: Date, default: Date.now },
    updated_at: { type: Date, default: Date.now },
  },
  { _id: false }
);

export const ModuleModel = mongoose.model("Module", moduleSchema);
