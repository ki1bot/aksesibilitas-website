import createClient from "openapi-fetch";

import type { paths } from "@/lib/api/schema";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://127.0.0.1:8080/api/v1";

export const CSRF_COOKIE_NAME =
  process.env.NEXT_PUBLIC_CSRF_COOKIE_NAME ??
  "aksesibilitaswebsite_session_csrf";

export const api = createClient<paths>({
  baseUrl: API_BASE_URL,
  credentials: "include",
});

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(message: string, status = 500, code = "unknown_error") {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function getCSRFToken(): string {
  if (typeof document === "undefined") {
    return "";
  }

  const prefix = `${encodeURIComponent(CSRF_COOKIE_NAME)}=`;

  const cookie = document.cookie
    .split(";")
    .map((value) => value.trim())
    .find((value) => value.startsWith(prefix));

  if (!cookie) {
    return "";
  }

  return decodeURIComponent(cookie.slice(prefix.length));
}

export function csrfHeaders(): Record<string, string> {
  const token = getCSRFToken();

  if (!token) {
    return {};
  }

  return {
    "X-CSRF-Token": token,
  };
}

export function throwApiError(
  error: unknown,
  response: Response,
  fallback: string,
): never {
  const payload =
    typeof error === "object" && error !== null
      ? (error as { code?: unknown; message?: unknown })
      : null;

  const code =
    typeof payload?.code === "string" ? payload.code : "request_failed";

  const message =
    typeof payload?.message === "string" ? payload.message : fallback;

  throw new ApiError(message, response.status, code);
}
