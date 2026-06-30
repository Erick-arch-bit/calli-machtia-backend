import { drizzle } from "drizzle-orm/node-postgres";
import pg from "pg";
import { config } from "../config";
import * as schema from "../schema";

const { Pool } = pg;

const pool = new Pool({
  connectionString: config.databaseUrl,
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 10000,
});

pool.on("error", (err) => {
  console.error("PostgreSQL pool error:", err);
});

export const db = drizzle(pool, { schema });

export async function pingPostgres(): Promise<boolean> {
  try {
    await pool.query("SELECT 1");
    return true;
  } catch {
    return false;
  }
}
