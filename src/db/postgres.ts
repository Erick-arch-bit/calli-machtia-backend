/**
 * Conexión a PostgreSQL mediante Drizzle ORM.
 * Configura un pool de conexiones y exporta la instancia de base de datos
 * tipada con el schema completo de la aplicación.
 */
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

/** Verifica que PostgreSQL responde correctamente (SELECT 1) */
export async function pingPostgres(): Promise<boolean> {
  try {
    await pool.query("SELECT 1");
    return true;
  } catch {
    return false;
  }
}
