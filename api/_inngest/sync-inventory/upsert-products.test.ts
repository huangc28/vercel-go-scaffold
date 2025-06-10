import { describe, expect, it } from "vitest";
import { upsertProducts } from "./upsert-products.ts";

describe("updateProducts", () => {
  it("should update products", async () => {
    const products = await upsertProducts([
      {
        uuid: "123",
        sku: "123",
        name: "123",
        ready_for_sale: "Y",
        stock_count: 1,
        price: 1,
        short_desc: "123",
      },
    ]);

    expect(products).toEqual({ inserted: 1, updated: 0, total: 1 });
  });
});
