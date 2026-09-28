import { API_BASE_URL } from "./config";
import type { ApiErrorResponse } from "./types";

export class ApiError extends Error {}

/** Posts multipart form data and unwraps the API's `{ error }` body on failure. */
export async function postForm<T>(path: string, body: FormData): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: "POST",
      body,
    });
  } catch {
    throw new ApiError(
      `Could not reach the API at ${API_BASE_URL}. Is the backend running?`
    );
  }

  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try {
      const data = (await response.json()) as ApiErrorResponse;
      if (data.error) message = data.error;
    } catch {
      // response body wasn't JSON; keep the generic status message
    }
    throw new ApiError(message);
  }

  return (await response.json()) as T;
}
