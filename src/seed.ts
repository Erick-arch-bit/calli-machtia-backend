import pg from "pg";
import { config } from "./config";

const { Pool } = pg;

async function seed() {
  const pool = new Pool({ connectionString: config.databaseUrl });
  const client = await pool.connect();

  try {
    await client.query(`INSERT INTO users (name, email, password_hash, role) VALUES
      ('Admin',      'admin@callimachtia.com', '$2a$10$swmPi2vWTbkFaAq7FO/gje/oa4f3QSj2.z2KhproTl58rrVTBI00G', 'admin'),
      ('Instructor', 'instructor@test.com',     '$2a$10$swmPi2vWTbkFaAq7FO/gjerf8u6TUHp1HYK0MrZ57EVxIdARYP1oG', 'instructor'),
      ('Alumno',     'alumno@test.com',         '$2a$10$swmPi2vWTbkFaAq7FO/gjerf8u6TUHp1HYK0MrZ57EVxIdARYP1oG', 'alumno')
    ON CONFLICT (email) DO NOTHING`);

    console.log("Seed completed");
    console.log("  admin@callimachtia.com / admin123");
    console.log("  instructor@test.com   / password123");
    console.log("  alumno@test.com       / password123");
  } finally {
    client.release();
    await pool.end();
  }
}

seed().catch((err) => {
  console.error("Seed failed:", err);
  process.exit(1);
});
