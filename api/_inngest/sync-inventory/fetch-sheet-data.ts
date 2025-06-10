import { google } from "googleapis";
import { env } from "#shared/env.ts";

export interface ProductRow {
  sku: string;
  name: string;
  uuid: string;
  ready_for_sale: "Y" | "N";
  stock_count: number;
  price: number;
  short_desc: string;
}

// Google Sheets API client setup
const getGoogleSheetsClient = () => {
  const auth = new google.auth.GoogleAuth({
    credentials: {
      client_email: env.GOOGLE_SERVICE_ACCOUNT_EMAIL,
      private_key: env.GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY?.replace(
        /\\n/g,
        "\n",
      ),
    },
    scopes: ["https://www.googleapis.com/auth/spreadsheets.readonly"],
  });

  return google.sheets({ version: "v4", auth });
};

// Function to fetch data from Google Sheets
export const fetchSheetData = async (): Promise<ProductRow[]> => {
  try {
    const sheets = getGoogleSheetsClient();
    const spreadsheetId = process.env.GOOGLE_SHEET_ID;
    const range = process.env.GOOGLE_SHEET_RANGE || "Sheet1!A2:H1000"; // Skip header row

    const response = await sheets.spreadsheets.values.get({
      spreadsheetId,
      range,
    });

    const rows = response.data.values || [];

    return rows
      .map(transformRowToProduct)
      .filter((product) => filterEmptyUUID(product));
  } catch (error) {
    console.error("Error fetching sheet data:", error);
    throw new Error(`Failed to fetch sheet data: ${error.message}`);
  }
};

function transformRowToProduct(row: string[]): ProductRow {
  return {
    sku: row[0] || "",
    uuid: row[1] || "",
    name: row[2] || "",
    ready_for_sale: row[3] === "Y" ? "Y" : "N",
    stock_count: parseInt(row[4] || "0"),
    price: parseFloat(row[5] || "0"),
    short_desc: row[6] || "",
  };
}

function filterEmptyUUID(products: ProductRow) {
  return products.uuid !== "";
}
