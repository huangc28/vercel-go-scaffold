import { Inngest } from "inngest";
import { updateProducts } from "./upsert-products.js";

const syncInventory = (inngest: Inngest) => {
  return inngest.createFunction(
    {
      id: "sync-inventory",
      retries: 3,
    },
    { cron: "*/15 * * * *" }, // every 15 minutes
    async ({ event }) => {
    },
  );
};

export { syncInventory };
