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

  function showMutationError(error: unknown) {
    toast.error(error instanceof Error ? error.message : "Permintaan gagal");
  }

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
        className="inline-flex items-center gap-2 text-sm font-bold text-slate-600 hover:text-blue-600 dark:text-slate-300 dark:hover:text-blue-400"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        Kembali ke dashboard
      </Link>

      {scanQuery.isPending ? (
        <LoadingState label="Memuat data scan" />
      ) : scanQuery.error || !scan ? (
        <div className="mt-8">
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
          <section className="mt-7 overflow-hidden rounded-[2rem] border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/[0.03]">
            <div className="grid gap-8 p-6 sm:p-8 xl:grid-cols-[1fr_auto] xl:items-center">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-3">
                  <StatusBadge status={scan.status} />

                  <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-400">
                    <Clock3 className="size-3.5" aria-hidden="true" />
                    {formatDate(scan.created_at)}
                  </span>
                </div>

                <h1 className="mt-5 text-3xl font-black tracking-tight sm:text-4xl">
                  {scan.page_title || "Pemeriksaan aksesibilitas"}
                </h1>

                <a
                  href={scan.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="mt-3 block break-all text-sm font-semibold text-blue-600 hover:underline dark:text-blue-400"
                >
                  {scan.url}
                </a>

                <div className="mt-6 flex flex-wrap gap-3">
                  {(scan.status === "queued" || scan.status === "running") && (
                    <button
                      type="button"
                      onClick={() => cancelMutation.mutate()}
                      disabled={cancelMutation.isPending}
                      className="inline-flex h-11 items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 text-sm font-bold text-amber-800 disabled:cursor-not-allowed disabled:opacity-60 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-300"
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
                      className="inline-flex h-11 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-bold text-white disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <RefreshCcw className="size-4" aria-hidden="true" />
                      {retryMutation.isPending ? "Mengirim..." : "Ulangi scan"}
                    </button>
                  )}

                  <button
                    type="button"
                    onClick={handleDelete}
                    disabled={deleteMutation.isPending}
                    className="inline-flex h-11 items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 text-sm font-bold text-red-700 disabled:cursor-not-allowed disabled:opacity-60 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300"
                  >
                    <Trash2 className="size-4" aria-hidden="true" />
                    {deleteMutation.isPending ? "Menghapus..." : "Hapus"}
                  </button>
                </div>

                {scan.error_message && (
                  <div
                    role="alert"
                    className="mt-6 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm leading-6 text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300"
                  >
                    {scan.error_message}
                  </div>
                )}
              </div>

              <ScoreGauge score={scan.automated_score} />
            </div>

            {(scan.status === "queued" || scan.status === "running") && (
              <div className="border-t border-slate-200 bg-slate-50 px-6 py-5 dark:border-white/10 dark:bg-white/[0.02] sm:px-8">
                <div className="flex flex-col gap-2 text-sm sm:flex-row sm:items-center sm:justify-between">
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
              <section className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
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
                    className="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.03]"
                  >
                    <strong className="text-3xl font-black">
                      {item.value}
                    </strong>

                    <p className="mt-2 text-sm font-semibold text-slate-500 dark:text-slate-400">
                      Dampak {item.label}
                    </p>
                  </article>
                ))}
              </section>

              <section className="mt-6 grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
                <article className="rounded-[2rem] border border-slate-200 bg-white p-6 shadow-sm dark:border-white/10 dark:bg-white/[0.03]">
                  <h2 className="text-xl font-black">Distribusi pelanggaran</h2>

                  <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
                    Berdasarkan tingkat dampak axe-core
                  </p>

                  <div className="mt-6 h-64">
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

                <article className="rounded-[2rem] bg-slate-950 p-6 text-white shadow-xl">
                  <p className="text-sm font-bold uppercase tracking-[0.18em] text-blue-300">
                    Laporan audit
                  </p>

                  <h2 className="mt-3 text-2xl font-black">
                    Ekspor hasil pemeriksaan
                  </h2>

                  <p className="mt-3 max-w-xl text-sm leading-6 text-slate-300">
                    Laporan memuat skor otomatis, jumlah pelanggaran, DOM
                    snippet, selector, saran perbaikan, dan checklist manual.
                  </p>

                  <dl className="mt-6 grid grid-cols-2 gap-4">
                    <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                      <dt className="text-xs text-slate-400">Durasi scan</dt>

                      <dd className="mt-1 font-bold">
                        {formatDuration(scan.duration_ms)}
                      </dd>
                    </div>

                    <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                      <dt className="text-xs text-slate-400">
                        Total violation
                      </dt>

                      <dd className="mt-1 font-bold">{violations.length}</dd>
                    </div>
                  </dl>

                  <div className="mt-6 flex flex-wrap gap-3">
                    <button
                      type="button"
                      onClick={() => reportMutation.mutate("pdf")}
                      disabled={reportMutation.isPending}
                      className="inline-flex h-11 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-bold disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <FileText className="size-4" aria-hidden="true" />
                      Unduh PDF
                    </button>

                    <button
                      type="button"
                      onClick={() => reportMutation.mutate("json")}
                      disabled={reportMutation.isPending}
                      className="inline-flex h-11 items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 text-sm font-bold disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      <FileJson className="size-4" aria-hidden="true" />
                      Unduh JSON
                    </button>
                  </div>
                </article>
              </section>

              <section className="mt-10">
                <div>
                  <h2 className="text-xl font-black">Pelanggaran otomatis</h2>

                  <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
                    Tinjau setiap aturan, selector, DOM snippet, dan saran
                    perbaikannya.
                  </p>
                </div>

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
                    <div className="rounded-3xl border border-emerald-200 bg-emerald-50 p-6 dark:border-emerald-400/20 dark:bg-emerald-400/10">
                      <div className="flex items-start gap-4">
                        <CheckCircle2
                          className="mt-0.5 size-6 shrink-0 text-emerald-600 dark:text-emerald-400"
                          aria-hidden="true"
                        />

                        <div>
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

              <section className="mt-10">
                <h2 className="text-xl font-black">Pemeriksaan manual</h2>

                <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
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

              <section className="mt-10 rounded-3xl border border-amber-200 bg-amber-50 p-6 dark:border-amber-400/20 dark:bg-amber-400/10">
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
    <article className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/[0.03]">
      <button
        type="button"
        onClick={() => setExpanded((value) => !value)}
        className="flex w-full items-start justify-between gap-5 p-5 text-left sm:p-6"
        aria-expanded={expanded}
      >
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <ImpactBadge impact={violation.impact} />
            <ReviewBadge status={violation.review_status} />
          </div>

          <h3 className="mt-4 font-black">{violation.help}</h3>

          <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-400">
            {violation.description}
          </p>

          <code className="mt-3 inline-flex rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-700 dark:bg-white/5 dark:text-slate-300">
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
        <div className="border-t border-slate-200 bg-slate-50/70 p-5 dark:border-white/10 dark:bg-white/[0.02] sm:p-6">
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
                    className="rounded-2xl border border-slate-200 bg-white p-4 dark:border-white/10 dark:bg-slate-950"
                  >
                    <p className="text-xs font-bold uppercase tracking-[0.16em] text-slate-400">
                      Node {index + 1}
                    </p>

                    <p className="mt-3 break-all text-sm font-semibold text-blue-600 dark:text-blue-400">
                      {node.target.join(" ")}
                    </p>

                    <pre className="mt-3 overflow-x-auto whitespace-pre-wrap break-all rounded-xl bg-slate-950 p-4 text-xs leading-6 text-slate-200">
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

              <div className="grid gap-4 sm:grid-cols-2">
                <label>
                  <span className="mb-2 block text-sm font-bold">
                    Status review
                  </span>

                  <select
                    value={status}
                    onChange={(event) =>
                      setStatus(event.target.value as ReviewStatus)
                    }
                    className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none focus:border-blue-500 dark:border-white/10 dark:bg-slate-950"
                  >
                    {reviewStatuses.map((value) => (
                      <option key={value} value={value}>
                        {reviewStatusLabel(value)}
                      </option>
                    ))}
                  </select>
                </label>

                <label>
                  <span className="mb-2 block text-sm font-bold">Catatan</span>

                  <textarea
                    value={notes}
                    onChange={(event) => setNotes(event.target.value)}
                    maxLength={5000}
                    rows={3}
                    className="w-full resize-none rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-blue-500 dark:border-white/10 dark:bg-slate-950"
                    placeholder="Catatan hasil verifikasi..."
                  />
                </label>
              </div>

              <div className="flex flex-wrap items-center gap-3">
                <button
                  type="button"
                  onClick={() => updateMutation.mutate()}
                  disabled={updateMutation.isPending}
                  className="inline-flex h-11 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-bold text-white disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <Save className="size-4" aria-hidden="true" />

                  {updateMutation.isPending ? "Menyimpan..." : "Simpan review"}
                </button>

                {violation.help_url && (
                  <a
                    href={violation.help_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm font-bold text-blue-600 hover:underline dark:text-blue-400"
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
    <article className="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.03] sm:p-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-600 dark:text-blue-400">
            Pemeriksaan {item.position}
          </p>

          <h3 className="mt-2 font-black">{item.criterion}</h3>

          <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-600 dark:text-slate-400">
            {item.instruction}
          </p>
        </div>

        <ReviewBadge status={item.status} />
      </div>

      <div className="mt-5 grid gap-4 md:grid-cols-[0.4fr_1fr_auto] md:items-end">
        <label>
          <span className="mb-2 block text-sm font-bold">Status</span>

          <select
            value={status}
            onChange={(event) => setStatus(event.target.value as ReviewStatus)}
            className="h-11 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm outline-none focus:border-blue-500 dark:border-white/10 dark:bg-slate-950"
          >
            {reviewStatuses.map((value) => (
              <option key={value} value={value}>
                {reviewStatusLabel(value)}
              </option>
            ))}
          </select>
        </label>

        <label>
          <span className="mb-2 block text-sm font-bold">Catatan</span>

          <textarea
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            maxLength={5000}
            rows={2}
            placeholder="Tuliskan hasil pemeriksaan..."
            className="w-full resize-none rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-blue-500 dark:border-white/10 dark:bg-slate-950"
          />
        </label>

        <button
          type="button"
          onClick={() => mutation.mutate()}
          disabled={mutation.isPending}
          className="inline-flex h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-bold text-white disabled:cursor-not-allowed disabled:opacity-60"
        >
          <Save className="size-4" aria-hidden="true" />

          {mutation.isPending ? "Menyimpan..." : "Simpan"}
        </button>
      </div>
    </article>
  );
}
