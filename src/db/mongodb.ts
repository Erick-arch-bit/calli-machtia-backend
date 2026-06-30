import mongoose from "mongoose";
import { config } from "../config";

export async function connectMongoDB(): Promise<void> {
  try {
    await mongoose.connect(config.mongodbUri, {
      dbName: config.mongodbDatabase,
    });
    console.log("MongoDB connected");
  } catch (err) {
    console.error("MongoDB connection error:", err);
  }
}

export async function pingMongoDB(): Promise<boolean> {
  try {
    if (mongoose.connection.readyState !== 1) return false;
    await mongoose.connection.db!.admin().ping();
    return true;
  } catch {
    return false;
  }
}

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
  },
  { _id: false }
);

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
