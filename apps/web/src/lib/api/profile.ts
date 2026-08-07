import { API_BASE_URL, csrfHeaders, throwApiError } from "@/lib/api/client";
import type { User } from "@/lib/api/types";

export type UpdateProfileInput = {
  name: string;
  email: string;
  current_password: string;
};

export async function updateCurrentUser(
  input: UpdateProfileInput,
): Promise<User> {
  const response = await fetch(`${API_BASE_URL}/auth/me`, {
    method: "PATCH",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...csrfHeaders(),
    },
    body: JSON.stringify(input),
  });

  const payload = await response.json().catch(() => null);

  if (!response.ok) {
    throwApiError(payload, response, "Profil tidak dapat diperbarui");
  }

  return payload as User;
}
