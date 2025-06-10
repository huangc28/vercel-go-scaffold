import { Inngest, NonRetriableError } from "inngest";
import { tryCatch } from "#shared/try-catch.js";
import { upsertProducts } from "./upsert-products.js";
import { fetchSheetData } from "./fetch-sheet-data.js";

export const syncFunc = async ({ step }) => {
  const [products, fetchError] = await tryCatch(
    step.run("fetch-sheet-data", async () => {
      const products = await fetchSheetData();

      console.log("products", products);

      return products;
    }),
  );

  if (fetchError) {
    console.error("Error fetching sheet data:", fetchError);
    throw fetchError;
  }

  if (!products || products.length === 0) {
    console.info("No products to update");
    throw new NonRetriableError("No products to update");
  }

  const [result, upsertError] = await tryCatch(
    step.run("upsert-products", async () => upsertProducts(products)),
  );

  console.info("result synced", result);

  if (upsertError) {
    console.error("Error upserting products:", upsertError);
    throw upsertError;
  }

  return {
    inserted: result.inserted,
    updated: result.updated,
    total: result.total,
  };
};

export const syncInventory = (inngest: Inngest) => {
  return inngest.createFunction(
    {
      id: "sync-inventory",
      retries: 3,
    },
    { cron: "*/15 * * * *" }, // every 15 minutes
    syncFunc,
  );
};
