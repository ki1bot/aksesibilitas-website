import {
  API_BASE_URL,
  api,
  csrfHeaders,
  throwApiError,
} from "@/lib/api/client";
import type {
  AuthResponse,
  ManualReviewItem,
  ManualReviewResponse,
  Project,
  ProjectRequest,
  Report,
  ReportFormat,
  ReviewStatus,
  Scan,
  User,
  Violation,
  ViolationDetail,
} from "@/lib/api/types";

export async function registerAccount(input: {
  name: string;
  email: string;
  password: string;
}): Promise<AuthResponse> {
  const result = await api.POST("/auth/register", {
    body: input,
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Pendaftaran gagal");
  }

  return result.data;
}

export async function loginAccount(input: {
  email: string;
  password: string;
}): Promise<AuthResponse> {
  const result = await api.POST("/auth/login", {
    body: input,
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Login gagal");
  }

  return result.data;
}

export async function logoutAccount(): Promise<void> {
  const result = await api.POST("/auth/logout", {
    headers: csrfHeaders(),
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Logout gagal");
  }
}

export async function getCurrentUser(): Promise<User> {
  const result = await api.GET("/auth/me");

  if (result.error) {
    throwApiError(
      result.error,
      result.response,
      "Sesi tidak valid atau sudah berakhir",
    );
  }

  return result.data;
}

export async function listProjects(): Promise<Project[]> {
  const result = await api.GET("/projects");

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal mengambil project");
  }

  return result.data;
}

export async function createProject(input: ProjectRequest): Promise<Project> {
  const result = await api.POST("/projects", {
    headers: csrfHeaders(),
    body: input,
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal membuat project");
  }

  return result.data;
}

export async function getProject(projectId: string): Promise<Project> {
  const result = await api.GET("/projects/{projectId}/", {
    params: {
      path: {
        projectId,
      },
    },
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Project tidak ditemukan");
  }

  return result.data;
}

export async function updateProject(
  projectId: string,
  input: ProjectRequest,
): Promise<Project> {
  const result = await api.PATCH("/projects/{projectId}/", {
    params: {
      path: {
        projectId,
      },
    },
    headers: csrfHeaders(),
    body: input,
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal memperbarui project");
  }

  return result.data;
}

export async function deleteProject(projectId: string): Promise<void> {
  const result = await api.DELETE("/projects/{projectId}/", {
    params: {
      path: {
        projectId,
      },
    },
    headers: csrfHeaders(),
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal menghapus project");
  }
}

export async function listScans(
  projectId?: string,
  limit = 50,
): Promise<Scan[]> {
  const result = await api.GET("/scans", {
    params: {
      query: {
        project_id: projectId,
        limit,
      },
    },
  });

  if (result.error) {
    throwApiError(
      result.error,
      result.response,
      "Gagal mengambil histori scan",
    );
  }

  return result.data;
}

export async function createScan(input: {
  project_id: string;
  url: string;
}): Promise<Scan> {
  const result = await api.POST("/scans", {
    headers: csrfHeaders(),
    body: input,
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal membuat scan");
  }

  return result.data;
}

export async function getScan(scanId: string): Promise<Scan> {
  const result = await api.GET("/scans/{scanId}/", {
    params: {
      path: {
        scanId,
      },
    },
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Scan tidak ditemukan");
  }

  return result.data;
}

export async function cancelScan(scanId: string): Promise<Scan> {
  const result = await api.POST("/scans/{scanId}/cancel", {
    params: {
      path: {
        scanId,
      },
    },
    headers: csrfHeaders(),
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal membatalkan scan");
  }

  return result.data;
}

export async function retryScan(scanId: string): Promise<Scan> {
  const result = await api.POST("/scans/{scanId}/retry", {
    params: {
      path: {
        scanId,
      },
    },
    headers: csrfHeaders(),
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal mengulangi scan");
  }

  return result.data;
}

export async function deleteScan(scanId: string): Promise<void> {
  const result = await api.DELETE("/scans/{scanId}/", {
    params: {
      path: {
        scanId,
      },
    },
    headers: csrfHeaders(),
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal menghapus scan");
  }
}

export async function listViolations(scanId: string): Promise<Violation[]> {
  const result = await api.GET("/scans/{scanId}/violations", {
    params: {
      path: {
        scanId,
      },
    },
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal mengambil violation");
  }

  return result.data;
}

export async function getViolation(
  violationId: string,
): Promise<ViolationDetail> {
  const result = await api.GET("/violations/{violationId}/", {
    params: {
      path: {
        violationId,
      },
    },
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Violation tidak ditemukan");
  }

  return result.data;
}

export async function updateViolationReview(
  violationId: string,
  input: {
    status: ReviewStatus;
    notes: string;
  },
): Promise<Violation> {
  const result = await api.PATCH("/violations/{violationId}/", {
    params: {
      path: {
        violationId,
      },
    },
    headers: csrfHeaders(),
    body: input,
  });

  if (result.error) {
    throwApiError(
      result.error,
      result.response,
      "Gagal menyimpan review violation",
    );
  }

  return result.data;
}

export async function getManualReview(
  scanId: string,
): Promise<ManualReviewResponse> {
  const result = await api.GET("/scans/{scanId}/manual-review", {
    params: {
      path: {
        scanId,
      },
    },
  });

  if (result.error) {
    throwApiError(
      result.error,
      result.response,
      "Gagal mengambil pemeriksaan manual",
    );
  }

  return result.data;
}

export async function updateManualReviewItem(
  itemId: string,
  input: {
    status: ReviewStatus;
    notes: string;
  },
): Promise<ManualReviewItem> {
  const result = await api.PATCH("/manual-review/items/{itemId}", {
    params: {
      path: {
        itemId,
      },
    },
    headers: csrfHeaders(),
    body: input,
  });

  if (result.error) {
    throwApiError(
      result.error,
      result.response,
      "Gagal menyimpan pemeriksaan manual",
    );
  }

  return result.data;
}

export async function createReport(
  scanId: string,
  format: ReportFormat,
): Promise<Report> {
  const result = await api.POST("/scans/{scanId}/reports", {
    params: {
      path: {
        scanId,
      },
    },
    headers: csrfHeaders(),
    body: {
      format,
    },
  });

  if (result.error) {
    throwApiError(result.error, result.response, "Gagal membuat laporan");
  }

  return result.data;
}

export async function downloadReportFile(
  reportId: string,
  filename: string,
): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/reports/${encodeURIComponent(reportId)}/download`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    let message = "Gagal mengunduh laporan";

    try {
      const payload = (await response.json()) as {
        message?: string;
      };

      if (payload.message) {
        message = payload.message;
      }
    } catch {
      message = "Gagal mengunduh laporan";
    }

    throw new Error(message);
  }

  const blob = await response.blob();
  const objectURL = URL.createObjectURL(blob);
  const anchor = document.createElement("a");

  anchor.href = objectURL;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();

  URL.revokeObjectURL(objectURL);
}
