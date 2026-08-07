"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  Ban,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Clock3,
  FileJson,
  FileText,
  RefreshCcw,
  Save,
  Trash2,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { toast } from "sonner";

import { AppShell } from "@/components/app-shell";
import {
  EmptyState,
  ErrorState,
  ImpactBadge,
  LoadingState,
  ReviewBadge,
  ScoreGauge,
  StatusBadge,
} from "@/components/ui-kit";
import {
  cancelScan,
  createReport,
  deleteScan,
  downloadReportFile,
  getManualReview,
  getScan,
  getViolation,
  listViolations,
  retryScan,
  updateManualReviewItem,
  updateViolationReview,
} from "@/lib/api/services";
import type {
  ManualReviewItem,
  ReportFormat,
  ReviewStatus,
  Violation,
} from "@/lib/api/types";
import { formatDate, formatDuration, reviewStatusLabel } from "@/lib/format";

const reviewStatuses: ReviewStatus[] = [
  "pending",
  "passed",
  "failed",
  "not_applicable",
];

export function ScanView({ scanId }: { scanId: string }) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const scanQuery = useQuery({
    queryKey: ["scan", scanId],
    queryFn: () => getScan(scanId),
    refetchInterval: (query) => {
      const status = query.state.data?.status;

      if (status === "queued" || status === "running") {
        return 2_000;
      }

      return false;
    },
  });

  const completed = scanQuery.data?.status === "completed";

  const violationsQuery = useQuery({
    queryKey: ["violations", scanId],
    queryFn: () => listViolations(scanId),
    enabled: completed,
  });

  const manualReviewQuery = useQuery({
    queryKey: ["manual-review", scanId],
    queryFn: () => getManualReview(scanId),
    enabled: completed,
  });

  function showMutationError(error: unknown) {
    toast.error(error instanceof Error ? error.message : "Permintaan gagal");
  }

  const cancelMutation = useMutation({
    mutationFn: () => cancelScan(scanId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["scan", scanId],
      });

      toast.success("Scan berhasil dibatalkan");
    },
    onError: showMutationError,
  });

  const retryMutation = useMutation({
    mutationFn: () => retryScan(scanId),
    onSuccess: async () => {
      queryClient.removeQueries({
        queryKey: ["violations", scanId],
      });

      queryClient.removeQueries({
        queryKey: ["manual-review", scanId],
      });

      await queryClient.invalidateQueries({
        queryKey: ["scan", scanId],
      });

      toast.success("Scan dimasukkan kembali ke antrean");
    },
    onError: showMutationError,
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteScan(scanId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["scans"],
      });

      router.replace("/dashboard");
      toast.success("Scan berhasil dihapus");
    },
    onError: showMutationError,
  });

  const reportMutation = useMutation({
    mutationFn: async (format: ReportFormat) => {
      const report = await createReport(scanId, format);

      await downloadReportFile(report.id, report.filename);

      return report;
    },
    onSuccess: (report) => {
      toast.success(`Laporan ${report.format.toUpperCase()} berhasil dibuat`);
    },
    onError: showMutationError,
  });

  const scan = scanQuery.data;
  const violations = violationsQuery.data ?? [];

  const chartData = useMemo(() => {
    if (!scan) {
      return [];
    }

    return [
      {
        name: "Kritis",
        total: scan.critical_count,
      },
      {
        name: "Serius",
        total: scan.serious_count,
      },
      {
        name: "Sedang",
        total: scan.moderate_count,
      },
      {
        name: "Ringan",
        total: scan.minor_count,
      },
    ];
  }, [scan]);

  function handleDelete() {
    const confirmed = window.confirm(
      "Scan, violation, pemeriksaan manual, dan laporan terkait akan dihapus permanen. Lanjutkan?",
    );

    if (confirmed) {
      deleteMutation.mutate();
    }
  }

  return (
    <AppShell>
      <Link
        href="/dashboard"
        className="inline-flex min-h-10 items-center gap-2 rounded-xl px-1 text-sm font-bold text-slate-600 transition hover:text-blue-600 dark:text-slate-300 dark:hover:text-blue-400"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        Kembali ke dashboard
      </Link>

      {scanQuery.isPending ? (
        <LoadingState label="Memuat data scan" />
      ) : scanQuery.error || !scan ? (
        <div className="mt-6 sm:mt-8">
          <ErrorState
            message={
              scanQuery.error instanceof Error
                ? scanQuery.error.message
                : "Scan tidak ditemukan"
            }
          />
        </div>
      ) : (
        <>
          <section className="mt-5 overflow-hidden rounded-[1.75rem] border border-slate-200 bg-white shadow-sm sm:mt-7 dark:border-white/10 dark:bg-white/[0.03]">
            <div className="grid min-w-0 gap-7 p-5 sm:p-7 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center lg:gap-10 lg:p-8">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2.5 sm:gap-3">
                  <StatusBadge status={scan.status} />

                  <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-400">
                    <Clock3 className="size-3.5" aria-hidden="true" />
                    {formatDate(scan.created_at)}
                  </span>
                </div>

                <h1 className="mt-4 break-words text-2xl font-black tracking-tight sm:mt-5 sm:text-3xl lg:text-4xl">
                  {scan.page_title || "Pemeriksaan aksesibilitas"}
                </h1>

                <a
                  href={scan.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-3 block break-all text-sm font-semibold leading-6 text-blue-600 hover:underline dark:text-blue-400"
                >
                  {scan.url}
                </a>

                <div className="mt-5 flex flex-col gap-2.5 sm:mt-6 sm:flex-row sm:flex-wrap sm:gap-3">
                  {(scan.status === "queued" || scan.status === "running") && (
                    <button
                      type="button"
                      onClick={() => cancelMutation.mutate()}
                      disabled={cancelMutation.isPending}
                      className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm font-bold text-amber-800 transition hover:bg-amber-100 disabled:cursor-not-allowed disabled:opacity-60 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-300 dark:hover:bg-amber-400/15"
                    >
                      <Ban className="size-4" aria-hidden="true" />
                      {cancelMutation.isPending ? "Membatalkan..." : "Batalkan"}
                    </button>
                  )}

                  {(scan.status === "failed" ||
                    scan.status === "cancelled") && (
                    <button
                      type="button"
                      onClick={() => retryMutation.mutate()}
                      disabled={retryMutation.isPending}
                      className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-bold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <RefreshCcw className="size-4" aria-hidden="true" />
                      {retryMutation.isPending ? "Mengirim..." : "Ulangi scan"}
                    </button>
                  )}

                  <button
                    type="button"
                    onClick={handleDelete}
                    disabled={deleteMutation.isPending}
                    className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-2.5 text-sm font-bold text-red-700 transition hover:bg-red-100 disabled:cursor-not-allowed disabled:opacity-60 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300 dark:hover:bg-red-400/15"
                  >
                    <Trash2 className="size-4" aria-hidden="true" />
                    {deleteMutation.isPending ? "Menghapus..." : "Hapus"}
                  </button>
                </div>

                {scan.error_message && (
                  <div
                    role="alert"
                    className="mt-5 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm leading-6 text-red-700 sm:mt-6 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300"
                  >
                    {scan.error_message}
                  </div>
                )}
              </div>

              <div className="flex justify-center lg:justify-end">
                <ScoreGauge score={scan.automated_score} />
              </div>
            </div>

            {(scan.status === "queued" || scan.status === "running") && (
              <div className="border-t border-slate-200 bg-slate-50 px-5 py-5 dark:border-white/10 dark:bg-white/[0.02] sm:px-7 lg:px-8">
                <div className="flex flex-col gap-1.5 text-sm sm:flex-row sm:items-center sm:justify-between sm:gap-4">
                  <span className="font-bold">
                    {scan.status === "queued"
                      ? "Menunggu worker tersedia"
                      : "Chromium sedang memeriksa halaman"}
                  </span>

                  <span className="text-slate-500 dark:text-slate-400">
                    Status diperbarui setiap 2 detik
                  </span>
                </div>

                <div className="mt-3 h-2 overflow-hidden rounded-full bg-slate-200 dark:bg-white/10">
                  <div className="h-full w-1/3 animate-[scan-progress_1.5s_ease-in-out_infinite] rounded-full bg-blue-600" />
                </div>
              </div>
            )}
          </section>

          {completed && (
            <>
              <section className="mt-5 grid grid-cols-2 gap-3 sm:mt-6 sm:gap-4 lg:grid-cols-4">
                {[
                  {
                    label: "Kritis",
                    value: scan.critical_count,
                  },
                  {
                    label: "Serius",
                    value: scan.serious_count,
                  },
                  {
                    label: "Sedang",
                    value: scan.moderate_count,
                  },
                  {
                    label: "Ringan",
                    value: scan.minor_count,
                  },
                ].map((item) => (
                  <article
                    key={item.label}
                    className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:rounded-3xl sm:p-5 dark:border-white/10 dark:bg-white/[0.03]"
                  >
                    <strong className="text-2xl font-black sm:text-3xl">
                      {item.value}
                    </strong>

                    <p className="mt-2 text-xs font-semibold leading-5 text-slate-500 sm:text-sm dark:text-slate-400">
                      Dampak {item.label}
                    </p>
                  </article>
                ))}
              </section>

              <section className="mt-5 grid gap-5 sm:mt-6 sm:gap-6 lg:grid-cols-2">
                <article className="min-w-0 rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm sm:p-6 dark:border-white/10 dark:bg-white/[0.03]">
                  <h2 className="text-lg font-black sm:text-xl">
                    Distribusi pelanggaran
                  </h2>

                  <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
                    Berdasarkan tingkat dampak axe-core
                  </p>

                  <div className="mt-5 h-64 min-w-0 sm:mt-6 sm:h-72">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={chartData}>
                        <CartesianGrid
                          strokeDasharray="4 4"
                          vertical={false}
                          stroke="var(--chart-grid)"
                        />

                        <XAxis
                          dataKey="name"
                          axisLine={false}
                          tickLine={false}
                          tick={{
                            fill: "var(--chart-label)",
                            fontSize: 12,
                          }}
                        />

                        <YAxis
                          allowDecimals={false}
                          axisLine={false}
                          tickLine={false}
                          width={32}
                          tick={{
                            fill: "var(--chart-label)",
                            fontSize: 12,
                          }}
                        />

                        <Tooltip
                          cursor={{
                            fill: "var(--chart-hover)",
                          }}
                          contentStyle={{
                            borderRadius: 14,
                            border: "1px solid var(--chart-grid)",
                            background: "var(--chart-tooltip)",
                          }}
                        />

                        <Bar
                          dataKey="total"
                          name="Temuan"
                          fill="var(--chart-accent)"
                          radius={[8, 8, 0, 0]}
                        />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </article>

                <article className="rounded-[1.75rem] bg-slate-950 p-5 text-white shadow-xl sm:p-6">
                  <p className="text-xs font-bold uppercase tracking-[0.18em] text-blue-300 sm:text-sm">
                    Laporan audit
                  </p>

                  <h2 className="mt-3 text-2xl font-black">
                    Ekspor hasil pemeriksaan
                  </h2>

                  <p className="mt-3 max-w-xl text-sm leading-6 text-slate-300">
                    Laporan memuat skor otomatis, jumlah pelanggaran, DOM
                    snippet, selector, saran perbaikan, dan checklist manual.
                  </p>

                  <dl className="mt-5 grid grid-cols-2 gap-3 sm:mt-6 sm:gap-4">
                    <div className="rounded-2xl border border-white/10 bg-white/5 p-3.5 sm:p-4">
                      <dt className="text-xs text-slate-400">Durasi scan</dt>

                      <dd className="mt-1 break-words font-bold">
                        {formatDuration(scan.duration_ms)}
                      </dd>
                    </div>

                    <div className="rounded-2xl border border-white/10 bg-white/5 p-3.5 sm:p-4">
                      <dt className="text-xs text-slate-400">
                        Total violation
                      </dt>

                      <dd className="mt-1 font-bold">{violations.length}</dd>
                    </div>
                  </dl>

                  <div className="mt-5 grid gap-2.5 sm:mt-6 sm:grid-cols-2 sm:gap-3">
                    <button
                      type="button"
                      onClick={() => reportMutation.mutate("pdf")}
                      disabled={reportMutation.isPending}
                      className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-bold transition hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <FileText className="size-4" aria-hidden="true" />
                      Unduh PDF
                    </button>

                    <button
                      type="button"
                      onClick={() => reportMutation.mutate("json")}
                      disabled={reportMutation.isPending}
                      className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-bold transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <FileJson className="size-4" aria-hidden="true" />
                      Unduh JSON
                    </button>
                  </div>
                </article>
              </section>

              <section className="mt-9 sm:mt-10">
                <h2 className="text-xl font-black">Pelanggaran otomatis</h2>

                <p className="mt-1 max-w-3xl text-sm leading-6 text-slate-500 dark:text-slate-400">
                  Tinjau setiap aturan, selector, DOM snippet, dan saran
                  perbaikannya.
                </p>

                <div className="mt-5">
                  {violationsQuery.isPending ? (
                    <LoadingState label="Memuat hasil axe-core" />
                  ) : violationsQuery.error ? (
                    <ErrorState
                      message={
                        violationsQuery.error instanceof Error
                          ? violationsQuery.error.message
                          : "Gagal mengambil violation"
                      }
                    />
                  ) : violations.length === 0 ? (
                    <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-5 sm:rounded-3xl sm:p-6 dark:border-emerald-400/20 dark:bg-emerald-400/10">
                      <div className="flex items-start gap-4">
                        <CheckCircle2
                          className="mt-0.5 size-6 shrink-0 text-emerald-600 dark:text-emerald-400"
                          aria-hidden="true"
                        />

                        <div className="min-w-0">
                          <h3 className="font-black text-emerald-950 dark:text-emerald-100">
                            Tidak ditemukan pelanggaran otomatis pada aturan
                            yang diperiksa.
                          </h3>

                          <p className="mt-2 text-sm leading-6 text-emerald-700 dark:text-emerald-300">
                            Pemeriksaan manual tetap diperlukan untuk menentukan
                            kesesuaian WCAG.
                          </p>
                        </div>
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {violations.map((violation) => (
                        <ViolationCard
                          key={`${violation.id}-${violation.updated_at}`}
                          violation={violation}
                          scanId={scanId}
                        />
                      ))}
                    </div>
                  )}
                </div>
              </section>

              <section className="mt-9 sm:mt-10">
                <h2 className="text-xl font-black">Pemeriksaan manual</h2>

                <p className="mt-1 max-w-3xl text-sm leading-6 text-slate-500 dark:text-slate-400">
                  Otomatisasi tidak dapat menilai seluruh pengalaman pengguna.
                  Selesaikan checklist berikut secara manual.
                </p>

                <div className="mt-5">
                  {manualReviewQuery.isPending ? (
                    <LoadingState label="Memuat checklist manual" />
                  ) : manualReviewQuery.error ? (
                    <ErrorState
                      message={
                        manualReviewQuery.error instanceof Error
                          ? manualReviewQuery.error.message
                          : "Gagal mengambil checklist manual"
                      }
                    />
                  ) : !manualReviewQuery.data?.items.length ? (
                    <EmptyState
                      title="Checklist tidak tersedia"
                      description="Belum ada item pemeriksaan manual untuk scan ini."
                    />
                  ) : (
                    <div className="space-y-4">
                      {manualReviewQuery.data.items.map((item) => (
                        <ManualReviewCard
                          key={`${item.id}-${item.updated_at}`}
                          item={item}
                          scanId={scanId}
                        />
                      ))}
                    </div>
                  )}
                </div>
              </section>

              <section className="mt-9 rounded-2xl border border-amber-200 bg-amber-50 p-5 sm:mt-10 sm:rounded-3xl sm:p-6 dark:border-amber-400/20 dark:bg-amber-400/10">
                <h2 className="font-black text-amber-950 dark:text-amber-100">
                  Batas hasil otomatis
                </h2>

                <p className="mt-2 text-sm leading-6 text-amber-800 dark:text-amber-300">
                  Skor otomatis adalah indikator teknis berdasarkan aturan yang
                  dapat diperiksa mesin. Hasil ini tidak boleh diklaim sebagai
                  bukti bahwa website telah 100% sesuai WCAG.
                </p>
              </section>
            </>
          )}
        </>
      )}
    </AppShell>
  );
}

function ViolationCard({
  violation,
  scanId,
}: {
  violation: Violation;
  scanId: string;
}) {
  const queryClient = useQueryClient();

  const [expanded, setExpanded] = useState(false);
  const [status, setStatus] = useState<ReviewStatus>(violation.review_status);
  const [notes, setNotes] = useState(violation.notes);

  const detailQuery = useQuery({
    queryKey: ["violation-detail", violation.id],
    queryFn: () => getViolation(violation.id),
    enabled: expanded,
  });

  const updateMutation = useMutation({
    mutationFn: () =>
      updateViolationReview(violation.id, {
        status,
        notes,
      }),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["violations", scanId],
        }),
        queryClient.invalidateQueries({
          queryKey: ["violation-detail", violation.id],
        }),
      ]);

      toast.success("Review violation disimpan");
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : "Gagal menyimpan review",
      );
    },
  });

  return (
    <article className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm sm:rounded-3xl dark:border-white/10 dark:bg-white/[0.03]">
      <button
        type="button"
        onClick={() => setExpanded((value) => !value)}
        className="flex w-full items-start justify-between gap-4 p-4 text-left transition hover:bg-slate-50 sm:gap-5 sm:p-6 dark:hover:bg-white/[0.02]"
        aria-expanded={expanded}
      >
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <ImpactBadge impact={violation.impact} />
            <ReviewBadge status={violation.review_status} />
          </div>

          <h3 className="mt-4 break-words font-black">{violation.help}</h3>

          <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-400">
            {violation.description}
          </p>

          <code className="mt-3 inline-flex max-w-full break-all rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-700 dark:bg-white/5 dark:text-slate-300">
            {violation.rule_id}
          </code>
        </div>

        {expanded ? (
          <ChevronUp
            className="mt-1 size-5 shrink-0 text-slate-400"
            aria-hidden="true"
          />
        ) : (
          <ChevronDown
            className="mt-1 size-5 shrink-0 text-slate-400"
            aria-hidden="true"
          />
        )}
      </button>

      {expanded && (
        <div className="border-t border-slate-200 bg-slate-50/70 p-4 sm:p-6 dark:border-white/10 dark:bg-white/[0.02]">
          {detailQuery.isPending ? (
            <LoadingState label="Memuat DOM snippet" />
          ) : detailQuery.error ? (
            <ErrorState
              message={
                detailQuery.error instanceof Error
                  ? detailQuery.error.message
                  : "Detail tidak dapat dimuat"
              }
            />
          ) : (
            <div className="space-y-5">
              {detailQuery.data?.nodes.length ? (
                detailQuery.data.nodes.map((node, index) => (
                  <div
                    key={node.id}
                    className="min-w-0 rounded-2xl border border-slate-200 bg-white p-4 dark:border-white/10 dark:bg-slate-950"
                  >
                    <p className="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">
                      Node {index + 1}
                    </p>

                    <p className="mt-3 break-all text-sm font-semibold leading-6 text-blue-600 dark:text-blue-400">
                      {node.target.join(" ")}
                    </p>

                    <pre className="mt-3 max-w-full overflow-x-auto rounded-xl bg-slate-950 p-4 text-xs leading-6 text-slate-200">
                      <code>{node.html}</code>
                    </pre>

                    {node.failure_summary && (
                      <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400">
                        {node.failure_summary}
                      </p>
                    )}
                  </div>
                ))
              ) : (
                <EmptyState
                  title="DOM node tidak tersedia"
                  description="Violation ini tidak memiliki data DOM node yang dapat ditampilkan."
                />
              )}

              <div className="grid gap-4 lg:grid-cols-[minmax(180px,0.4fr)_minmax(0,1fr)]">
                <label className="min-w-0">
                  <span className="mb-2 block text-sm font-bold">
                    Status review
                  </span>

                  <select
                    value={status}
                    onChange={(event) =>
                      setStatus(event.target.value as ReviewStatus)
                    }
                    className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-slate-950"
                  >
                    {reviewStatuses.map((value) => (
                      <option key={value} value={value}>
                        {reviewStatusLabel(value)}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="min-w-0">
                  <span className="mb-2 block text-sm font-bold">Catatan</span>

                  <textarea
                    value={notes}
                    onChange={(event) => setNotes(event.target.value)}
                    maxLength={5000}
                    rows={3}
                    className="w-full resize-y rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-slate-950"
                    placeholder="Catatan hasil verifikasi..."
                  />
                </label>
              </div>

              <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
                <button
                  type="button"
                  onClick={() => updateMutation.mutate()}
                  disabled={updateMutation.isPending}
                  className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-bold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <Save className="size-4" aria-hidden="true" />

                  {updateMutation.isPending ? "Menyimpan..." : "Simpan review"}
                </button>

                {violation.help_url && (
                  <a
                    href={violation.help_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex min-h-11 items-center justify-center rounded-xl px-2 text-sm font-bold text-blue-600 hover:underline sm:justify-start dark:text-blue-400"
                  >
                    Buka referensi aturan
                  </a>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </article>
  );
}

function ManualReviewCard({
  item,
  scanId,
}: {
  item: ManualReviewItem;
  scanId: string;
}) {
  const queryClient = useQueryClient();

  const [status, setStatus] = useState<ReviewStatus>(item.status);
  const [notes, setNotes] = useState(item.notes);

  const mutation = useMutation({
    mutationFn: () =>
      updateManualReviewItem(item.id, {
        status,
        notes,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["manual-review", scanId],
      });

      toast.success("Pemeriksaan manual disimpan");
    },
    onError: (error) => {
      toast.error(
        error instanceof Error
          ? error.message
          : "Gagal menyimpan pemeriksaan manual",
      );
    },
  });

  return (
    <article className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:rounded-3xl sm:p-6 dark:border-white/10 dark:bg-white/[0.03]">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between sm:gap-4">
        <div className="min-w-0">
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-600 dark:text-blue-400">
            Pemeriksaan {item.position}
          </p>

          <h3 className="mt-2 break-words font-black">{item.criterion}</h3>

          <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-600 dark:text-slate-400">
            {item.instruction}
          </p>
        </div>

        <ReviewBadge status={item.status} />
      </div>

      <div className="mt-5 grid min-w-0 gap-4 lg:grid-cols-[minmax(180px,0.35fr)_minmax(0,1fr)_auto] lg:items-end">
        <label className="min-w-0">
          <span className="mb-2 block text-sm font-bold">Status</span>

          <select
            value={status}
            onChange={(event) => setStatus(event.target.value as ReviewStatus)}
            className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-slate-950"
          >
            {reviewStatuses.map((value) => (
              <option key={value} value={value}>
                {reviewStatusLabel(value)}
              </option>
            ))}
          </select>
        </label>

        <label className="min-w-0">
          <span className="mb-2 block text-sm font-bold">Catatan</span>

          <textarea
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            maxLength={5000}
            rows={2}
            placeholder="Tuliskan hasil pemeriksaan..."
            className="w-full resize-y rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-slate-950"
          />
        </label>

        <button
          type="button"
          onClick={() => mutation.mutate()}
          disabled={mutation.isPending}
          className="inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-bold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60 lg:w-auto"
        >
          <Save className="size-4" aria-hidden="true" />

          {mutation.isPending ? "Menyimpan..." : "Simpan"}
        </button>
      </div>
    </article>
  );
}
