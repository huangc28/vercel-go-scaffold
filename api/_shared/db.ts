import { Pool } from "pg";
import { env } from "#shared/env.js";

let pool: Pool | null = null;

const getPool = () => {
  if (!pool) {
    const dbDSN =
      `postgresql://${env.DB_USER}:${env.DB_PASSWORD}@${env.DB_HOST}:${env.DB_PORT}/${env.DB_NAME}`;

    pool = new Pool({
      connectionString: dbDSN,
      max: 20,
      idleTimeoutMillis: 30_000, // 30 seconds
      connectionTimeoutMillis: 2_000,
    });

    pool.on("connect", () => {
      console.log("✅ Connected to pg pool");
    });

    pool.on("error", (err) => {
      console.error("❌ Unexpected pg pool error", err);
    });
  }

  return pool;
};

export default getPool();
