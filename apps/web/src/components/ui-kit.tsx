import {
  CircleAlert,
  CircleCheck,
  CircleDashed,
  CircleMinus,
  CircleX,
  Inbox,
  LoaderCircle,
} from "lucide-react";

import type {
  ReviewStatus,
  ScanStatus,
  ViolationImpact,
} from "@/lib/api/types";
import { impactLabel, reviewStatusLabel, scanStatusLabel } from "@/lib/format";
import { cn } from "@/lib/utils";

export function StatusBadge({ status }: { status: ScanStatus }) {
  const styles: Record<ScanStatus, string> = {
    queued:
      "border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-300",
    running:
      "border-blue-200 bg-blue-50 text-blue-800 dark:border-blue-400/20 dark:bg-blue-400/10 dark:text-blue-300",
    completed:
      "border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-300",
    failed:
      "border-red-200 bg-red-50 text-red-800 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300",
    cancelled:
      "border-slate-200 bg-slate-100 text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-300",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold",
        styles[status],
      )}
    >
      {scanStatusLabel(status)}
    </span>
  );
}

export function ImpactBadge({ impact }: { impact: ViolationImpact }) {
  const styles: Record<ViolationImpact, string> = {
    critical:
      "border-red-200 bg-red-50 text-red-800 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300",
    serious:
      "border-orange-200 bg-orange-50 text-orange-800 dark:border-orange-400/20 dark:bg-orange-400/10 dark:text-orange-300",
    moderate:
      "border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-300",
    minor:
      "border-blue-200 bg-blue-50 text-blue-800 dark:border-blue-400/20 dark:bg-blue-400/10 dark:text-blue-300",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold",
        styles[impact],
      )}
    >
      {impactLabel(impact)}
    </span>
  );
}

export function ReviewBadge({ status }: { status: ReviewStatus }) {
  const styles: Record<ReviewStatus, string> = {
    pending:
      "border-slate-200 bg-slate-100 text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-300",
    passed:
      "border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-300",
    failed:
      "border-red-200 bg-red-50 text-red-800 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300",
    not_applicable:
      "border-violet-200 bg-violet-50 text-violet-800 dark:border-violet-400/20 dark:bg-violet-400/10 dark:text-violet-300",
  };

  const icons: Record<ReviewStatus, React.ReactNode> = {
    pending: <CircleDashed className="size-3.5" aria-hidden="true" />,
    passed: <CircleCheck className="size-3.5" aria-hidden="true" />,
    failed: <CircleX className="size-3.5" aria-hidden="true" />,
    not_applicable: <CircleMinus className="size-3.5" aria-hidden="true" />,
  };

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-semibold",
        styles[status],
      )}
    >
      {icons[status]}
      {reviewStatusLabel(status)}
    </span>
  );
}

export function ScoreGauge({
  score,
  size = "large",
}: {
  score: number;
  size?: "small" | "large";
}) {
  const normalizedScore = Math.max(0, Math.min(100, score));

  return (
    <div
      className={cn(
        "relative grid place-items-center rounded-full",
        size === "large" ? "size-44" : "size-24",
      )}
      style={{
        background: `conic-gradient(var(--score-color) ${normalizedScore}%, var(--score-track) ${normalizedScore}% 100%)`,
      }}
      aria-label={`Skor otomatis ${normalizedScore} dari 100`}
    >
      <div
        className={cn(
          "grid place-items-center rounded-full bg-white shadow-inner dark:bg-slate-950",
          size === "large" ? "size-36" : "size-20",
        )}
      >
        <div className="text-center">
          <strong
            className={cn(
              "block font-black tracking-tight text-slate-950 dark:text-white",
              size === "large" ? "text-5xl" : "text-2xl",
            )}
          >
            {normalizedScore}
          </strong>
          <span className="text-xs font-semibold text-slate-500 dark:text-slate-400">
            dari 100
          </span>
        </div>
      </div>
    </div>
  );
}

export function LoadingState({ label = "Memuat data" }: { label?: string }) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center gap-3 text-center">
      <LoaderCircle
        className="size-8 animate-spin text-blue-600"
        aria-hidden="true"
      />
      <p className="text-sm font-medium text-slate-600 dark:text-slate-300">
        {label}
      </p>
    </div>
  );
}

export function EmptyState({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="flex min-h-56 flex-col items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-slate-50/60 px-6 text-center dark:border-white/10 dark:bg-white/[0.02]">
      <span className="mb-4 flex size-12 items-center justify-center rounded-2xl bg-white text-slate-500 shadow-sm ring-1 ring-slate-200 dark:bg-white/5 dark:text-slate-300 dark:ring-white/10">
        <Inbox className="size-5" aria-hidden="true" />
      </span>
      <h3 className="font-bold text-slate-950 dark:text-white">{title}</h3>
      <p className="mt-2 max-w-md text-sm leading-6 text-slate-600 dark:text-slate-400">
        {description}
      </p>
    </div>
  );
}

export function ErrorState({ message }: { message: string }) {
  return (
    <div className="flex min-h-56 flex-col items-center justify-center rounded-3xl border border-red-200 bg-red-50 px-6 text-center dark:border-red-400/20 dark:bg-red-400/10">
      <CircleAlert
        className="mb-4 size-8 text-red-600 dark:text-red-400"
        aria-hidden="true"
      />
      <h3 className="font-bold text-red-950 dark:text-red-100">
        Data tidak dapat dimuat
      </h3>
      <p className="mt-2 max-w-lg text-sm leading-6 text-red-700 dark:text-red-300">
        {message}
      </p>
    </div>
  );
}
