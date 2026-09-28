// Falls back to the local backend port so `npm run dev` works without a .env file.
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";
