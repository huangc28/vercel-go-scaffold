import { Inngest } from "inngest";

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
