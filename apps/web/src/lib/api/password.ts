import { API_BASE_URL, ApiError, csrfHeaders } from "@/lib/api/client";

type ErrorPayload = {
  code?: string;
  message?: string;
};

export type ForgotPasswordResponse = {
  message: string;
};

async function readPayload(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";

  if (!contentType.includes("application/json")) {
    return null;
  }

  try {
    return await response.json();
  } catch {
    return null;
  }
}

function toApiError(
  response: Response,
  payload: unknown,
  fallback: string,
): ApiError {
  const errorPayload =
    typeof payload === "object" && payload !== null
      ? (payload as ErrorPayload)
      : null;

  return new ApiError(
    typeof errorPayload?.message === "string" ? errorPayload.message : fallback,
    response.status,
    typeof errorPayload?.code === "string"
      ? errorPayload.code
      : "request_failed",
  );
}

export async function requestPasswordReset(input: {
  email: string;
}): Promise<ForgotPasswordResponse> {
  const response = await fetch(`${API_BASE_URL}/auth/forgot-password`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  const payload = await readPayload(response);

  if (!response.ok) {
    throw toApiError(response, payload, "Permintaan reset password gagal");
  }

  return payload as ForgotPasswordResponse;
}

export async function resetPassword(input: {
  token: string;
  password: string;
  password_confirmation: string;
}): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });

  const payload = await readPayload(response);

  if (!response.ok) {
    throw toApiError(response, payload, "Password gagal diubah");
  }
}

export async function changePassword(input: {
  current_password: string;
  new_password: string;
  password_confirmation: string;
}): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/auth/change-password`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...csrfHeaders(),
    },
    body: JSON.stringify(input),
  });

  const payload = await readPayload(response);

  if (!response.ok) {
    throw toApiError(response, payload, "Password gagal diubah");
  }
}
