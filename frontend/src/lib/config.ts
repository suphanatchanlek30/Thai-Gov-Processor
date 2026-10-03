// An empty value means same-origin: the Ingress serves "/" and "/api" from one
// host, so the browser calls relative URLs and one image works in every
// environment. Only when the variable is unset (plain `npm run dev` without a
// .env file) does it fall back to the local backend port.
export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
