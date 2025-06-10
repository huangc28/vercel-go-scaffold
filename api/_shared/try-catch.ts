export type Result<T> = [T | null, Error | null];

export const tryCatch = async <T>(p: T | Promise<T>): Promise<Result<T>> => {
  try {
    return [await p, null];
  } catch (error) {
    console.error("Error in tryCatch:", error);
    return [null, error as Error];
  }
};
