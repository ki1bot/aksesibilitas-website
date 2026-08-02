import type {
  ReviewStatus,
  ScanStatus,
  ViolationImpact,
} from "@/lib/api/types";

export function formatDate(value: string): string {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function formatDuration(milliseconds: number): string {
  if (milliseconds < 1000) {
    return `${milliseconds} ms`;
  }

  return `${(milliseconds / 1000).toFixed(1)} detik`;
}

export function truncate(value: string, length = 48): string {
  if (value.length <= length) {
    return value;
  }

  return `${value.slice(0, length - 1)}…`;
}

export function scanStatusLabel(status: ScanStatus): string {
  const labels: Record<ScanStatus, string> = {
    queued: "Dalam antrean",
    running: "Sedang dipindai",
    completed: "Selesai",
    failed: "Gagal",
    cancelled: "Dibatalkan",
  };

  return labels[status];
}

export function impactLabel(impact: ViolationImpact): string {
  const labels: Record<ViolationImpact, string> = {
    critical: "Kritis",
    serious: "Serius",
    moderate: "Sedang",
    minor: "Ringan",
  };

  return labels[impact];
}

export function reviewStatusLabel(status: ReviewStatus): string {
  const labels: Record<ReviewStatus, string> = {
    pending: "Belum diperiksa",
    passed: "Lulus",
    failed: "Bermasalah",
    not_applicable: "Tidak berlaku",
  };

  return labels[status];
}
