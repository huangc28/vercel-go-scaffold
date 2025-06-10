import { Pool } from "pg";
import { env } from "#shared/env.js";

let pool: Pool | null = null;

const getPool = () => {
  if (!pool) {
    let dbDSN =
      `postgresql://${env.DB_USER}:${env.DB_PASSWORD}@${env.DB_HOST}:${env.DB_PORT}/${env.DB_NAME}`;

    const isVercel = env.VERCEL_ENV === "preview" ||
      env.VERCEL_ENV === "production";

    if (isVercel) {
      dbDSN += "?pgbouncer=true";
      console.info("🔄 Using transaction pool mode");
    }

    const poolConfig = isVercel
      ? {
        connectionString: dbDSN,
        max: 1,
        idleTimeoutMillis: 1000, // 1 second
        connectionTimeoutMillis: 5_000,
      }
      : {
        connectionString: dbDSN,
        max: 20,
        idleTimeoutMillis: 30_000, // 30 seconds
        connectionTimeoutMillis: 2_000,
      };

    pool = new Pool(poolConfig);

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
