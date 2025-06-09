import { google } from "googleapis";

// Types for the sheet data
export interface ProductRow {
  name: string;
  // description: string;
  // price: number;
  // quantity: number;
  // sku: string;
  // category: string;
  // imageUrl: string;
  // isActive: boolean;
  // rowIndex: number;
}

// Google Sheets API client setup
const getGoogleSheetsClient = () => {
  const auth = new google.auth.GoogleAuth({
    credentials: {
      client_email: process.env.GOOGLE_SERVICE_ACCOUNT_EMAIL,
      private_key: process.env.GOOGLE_SERVICE_ACCOUNT_PRIVATE_KEY?.replace(
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

    return rows.map((row) => ({
      name: row[2] || "",
    })).filter((product) => product.name); // Filter out empty rows
  } catch (error) {
    console.error("Error fetching sheet data:", error);
    throw new Error(`Failed to fetch sheet data: ${error.message}`);
  }
};
