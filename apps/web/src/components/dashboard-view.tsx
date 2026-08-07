"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Activity,
  ArrowRight,
  CircleAlert,
  FolderKanban,
  Globe2,
  History,
  Plus,
  ScanSearch,
  X,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useMemo, useState } from "react";
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
import { z } from "zod";

import { AppShell } from "@/components/app-shell";
import {
  EmptyState,
  ErrorState,
  LoadingState,
  StatusBadge,
} from "@/components/ui-kit";
import {
  createProject,
  createScan,
  listProjects,
  listScans,
} from "@/lib/api/services";
import { formatDate, truncate } from "@/lib/format";

const projectSchema = z.object({
  name: z
    .string()
    .trim()
    .min(2, "Nama project minimal 2 karakter")
    .max(120, "Nama project maksimal 120 karakter"),
  description: z.string().trim().max(1000, "Deskripsi maksimal 1000 karakter"),
});

const scanSchema = z.object({
  project_id: z.string().uuid("Pilih project terlebih dahulu"),
  url: z.url("Masukkan alamat website yang valid"),
});

export function DashboardView() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const [projectDialogOpen, setProjectDialogOpen] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [projectDescription, setProjectDescription] = useState("");
  const [selectedProject, setSelectedProject] = useState("");
  const [scanURL, setScanURL] = useState("");

  const projectsQuery = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
  });

  const scansQuery = useQuery({
    queryKey: ["scans", "recent"],
    queryFn: () => listScans(undefined, 30),
  });

  const projects = useMemo(
    () => projectsQuery.data ?? [],
    [projectsQuery.data],
  );

  const scans = useMemo(() => scansQuery.data ?? [], [scansQuery.data]);

  const activeProjectID = selectedProject || projects[0]?.id || "";

  const summary = useMemo(() => {
    return scans.reduce(
      (result, scan) => {
        result.total += 1;
        result.critical += scan.critical_count;
        result.serious += scan.serious_count;
        result.moderate += scan.moderate_count;
        result.minor += scan.minor_count;

        if (scan.status === "completed") {
          result.completed += 1;
        }

        return result;
      },
      {
        total: 0,
        completed: 0,
        critical: 0,
        serious: 0,
        moderate: 0,
        minor: 0,
      },
    );
  }, [scans]);

  const chartData = [
    {
      name: "Kritis",
      total: summary.critical,
    },
    {
      name: "Serius",
      total: summary.serious,
    },
    {
      name: "Sedang",
      total: summary.moderate,
    },
    {
      name: "Ringan",
      total: summary.minor,
    },
  ];

  const projectMutation = useMutation({
    mutationFn: () => {
      const parsed = projectSchema.safeParse({
        name: projectName,
        description: projectDescription,
      });

      if (!parsed.success) {
        throw new Error(
          parsed.error.issues[0]?.message ?? "Data project belum benar",
        );
      }

      return createProject(parsed.data);
    },
    onSuccess: async (project) => {
      setProjectName("");
      setProjectDescription("");
      setProjectDialogOpen(false);
      setSelectedProject(project.id);

      await queryClient.invalidateQueries({
        queryKey: ["projects"],
      });

      toast.success("Project berhasil dibuat");
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : "Project tidak dapat dibuat",
      );
    },
  });

  const scanMutation = useMutation({
    mutationFn: () => {
      const parsed = scanSchema.safeParse({
        project_id: activeProjectID,
        url: scanURL.trim(),
      });

      if (!parsed.success) {
        throw new Error(
          parsed.error.issues[0]?.message ?? "Data pemeriksaan belum benar",
        );
      }

      return createScan(parsed.data);
    },
    onSuccess: (scan) => {
      router.push(`/scans/${scan.id}`);
    },
    onError: (error) => {
      toast.error(
        error instanceof Error
          ? error.message
          : "Pemeriksaan tidak dapat dimulai",
      );
    },
  });

  function submitProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    projectMutation.mutate();
  }

  function submitScan(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    scanMutation.mutate();
  }

  return (
    <AppShell>
      <section className="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-600 sm:text-sm dark:text-blue-400">
            Ringkasan akun
          </p>

          <h1 className="mt-2 text-3xl font-black tracking-tight sm:text-4xl">
            Pemeriksaan website
          </h1>

          <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-600 sm:text-base sm:leading-7 dark:text-slate-400">
            Tambahkan website, mulai pemeriksaan, lalu lanjutkan perbaikan dari
            masalah yang paling penting.
          </p>
        </div>

        <button
          type="button"
          onClick={() => setProjectDialogOpen(true)}
          className="inline-flex min-h-11 w-full shrink-0 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-800 shadow-sm transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700 sm:w-auto dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:border-blue-400/20 dark:hover:bg-blue-400/10 dark:hover:text-blue-300"
        >
          <Plus className="size-4" aria-hidden="true" />
          Tambah project
        </button>
      </section>

      <section className="mt-7 grid grid-cols-2 gap-3 sm:mt-8 sm:gap-4 xl:grid-cols-4">
        {[
          {
            label: "Project tersimpan",
            value: projects.length,
            icon: FolderKanban,
          },
          {
            label: "Total pemeriksaan",
            value: summary.total,
            icon: ScanSearch,
          },
          {
            label: "Pemeriksaan selesai",
            value: summary.completed,
            icon: Activity,
          },
          {
            label: "Masalah kritis",
            value: summary.critical,
            icon: CircleAlert,
          },
        ].map((item) => {
          const Icon = item.icon;

          return (
            <article
              key={item.label}
              className="min-w-0 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:rounded-3xl sm:p-5 dark:border-white/10 dark:bg-white/[0.03]"
            >
              <div className="flex items-start justify-between gap-3">
                <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-blue-50 text-blue-600 sm:size-11 sm:rounded-2xl dark:bg-blue-400/10 dark:text-blue-300">
                  <Icon className="size-4.5 sm:size-5" aria-hidden="true" />
                </span>

                <strong className="text-2xl font-black sm:text-3xl">
                  {item.value}
                </strong>
              </div>

              <p className="mt-4 text-xs font-semibold leading-5 text-slate-600 sm:mt-5 sm:text-sm dark:text-slate-400">
                {item.label}
              </p>
            </article>
          );
        })}
      </section>

      <section className="mt-5 grid gap-5 sm:mt-6 sm:gap-6 xl:grid-cols-[minmax(0,1.12fr)_minmax(360px,0.88fr)]">
        <article className="relative min-w-0 overflow-hidden rounded-[1.75rem] bg-slate-950 p-5 text-white shadow-xl sm:p-7 lg:p-8">
          <div className="absolute -right-20 -top-20 size-72 rounded-full bg-blue-600/30 blur-3xl" />
          <div className="absolute -bottom-24 left-20 size-56 rounded-full bg-violet-600/20 blur-3xl" />

          <div className="relative">
            <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-bold text-blue-200">
              <Globe2 className="size-3.5" aria-hidden="true" />
              Periksa satu halaman
            </span>

            <h2 className="mt-5 text-2xl font-black tracking-tight">
              Mulai pemeriksaan baru
            </h2>

            <p className="mt-2 max-w-xl text-sm leading-6 text-slate-300">
              Pilih project dan masukkan alamat halaman publik yang ingin
              diperiksa.
            </p>

            <form
              onSubmit={submitScan}
              className="mt-7 grid min-w-0 gap-4 md:grid-cols-2"
            >
              <label className="min-w-0">
                <span className="mb-2 block text-xs font-bold uppercase tracking-[0.12em] text-slate-400">
                  Project
                </span>

                <select
                  value={activeProjectID}
                  onChange={(event) => setSelectedProject(event.target.value)}
                  required
                  className="h-12 w-full min-w-0 rounded-xl border border-white/10 bg-white/10 px-4 text-sm font-semibold text-white outline-none focus:border-blue-400 focus:ring-4 focus:ring-blue-400/15"
                >
                  <option value="" className="text-slate-950">
                    Pilih project
                  </option>

                  {projects.map((project) => (
                    <option
                      key={project.id}
                      value={project.id}
                      className="text-slate-950"
                    >
                      {project.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="min-w-0">
                <span className="mb-2 block text-xs font-bold uppercase tracking-[0.12em] text-slate-400">
                  Alamat halaman
                </span>

                <input
                  type="url"
                  value={scanURL}
                  onChange={(event) => setScanURL(event.target.value)}
                  required
                  maxLength={2048}
                  placeholder="https://websiteanda.id/halaman"
                  className="h-12 w-full min-w-0 rounded-xl border border-white/10 bg-white/10 px-4 text-sm text-white outline-none placeholder:text-slate-500 focus:border-blue-400 focus:ring-4 focus:ring-blue-400/15"
                />
              </label>

              <div className="md:col-span-2">
                <button
                  type="submit"
                  disabled={scanMutation.isPending || projects.length === 0}
                  className="inline-flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold text-white transition hover:bg-blue-500 disabled:cursor-not-allowed disabled:opacity-50 sm:w-auto"
                >
                  {scanMutation.isPending ? "Memulai..." : "Mulai periksa"}

                  {!scanMutation.isPending && (
                    <ArrowRight className="size-4" aria-hidden="true" />
                  )}
                </button>
              </div>
            </form>

            {projects.length === 0 && !projectsQuery.isPending && (
              <p className="mt-4 text-sm font-medium leading-6 text-amber-300">
                Tambahkan project terlebih dahulu sebelum memulai pemeriksaan.
              </p>
            )}
          </div>
        </article>

        <article className="min-w-0 rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm sm:p-6 dark:border-white/10 dark:bg-white/[0.03]">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <h2 className="font-black">Masalah berdasarkan dampak</h2>

              <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
                Diambil dari 30 pemeriksaan terbaru
              </p>
            </div>

            <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-blue-50 text-blue-600 dark:bg-blue-400/10 dark:text-blue-300">
              <Activity className="size-5" aria-hidden="true" />
            </span>
          </div>

          <div className="mt-5 h-64 min-w-0 sm:h-72">
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
      </section>

      <section id="projects" className="mt-9 scroll-mt-24 sm:mt-10">
        <div>
          <h2 className="text-xl font-black">Project Anda</h2>

          <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
            Pisahkan pemeriksaan berdasarkan website atau aplikasi.
          </p>
        </div>

        <div className="mt-5">
          {projectsQuery.isPending ? (
            <LoadingState label="Memuat project" />
          ) : projectsQuery.error ? (
            <ErrorState
              message={
                projectsQuery.error instanceof Error
                  ? projectsQuery.error.message
                  : "Project tidak dapat dimuat"
              }
            />
          ) : projects.length === 0 ? (
            <EmptyState
              title="Belum ada project"
              description="Tambahkan project pertama agar hasil pemeriksaan tersusun dengan rapi."
            />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
              {projects.map((project) => (
                <Link
                  key={project.id}
                  href={`/projects/${project.id}`}
                  className="group min-w-0 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-1 hover:border-blue-300 hover:shadow-xl hover:shadow-blue-600/5 sm:rounded-3xl dark:border-white/10 dark:bg-white/[0.03] dark:hover:border-blue-400/30"
                >
                  <div className="flex items-start justify-between gap-4">
                    <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-blue-50 text-blue-600 dark:bg-blue-400/10 dark:text-blue-300">
                      <FolderKanban className="size-5" aria-hidden="true" />
                    </span>

                    <ArrowRight
                      className="size-4 shrink-0 text-slate-400 transition group-hover:translate-x-1 group-hover:text-blue-600"
                      aria-hidden="true"
                    />
                  </div>

                  <h3 className="mt-5 break-words font-black">
                    {project.name}
                  </h3>

                  <p className="mt-2 min-h-12 text-sm leading-6 text-slate-600 dark:text-slate-400">
                    {project.description || "Belum ada deskripsi."}
                  </p>

                  <p className="mt-5 text-xs font-medium text-slate-400">
                    Diperbarui {formatDate(project.updated_at)}
                  </p>
                </Link>
              ))}
            </div>
          )}
        </div>
      </section>

      <section id="history" className="mt-9 scroll-mt-24 sm:mt-10">
        <div className="flex items-start gap-3">
          <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-slate-100 text-slate-600 dark:bg-white/5 dark:text-slate-300">
            <History className="size-5" aria-hidden="true" />
          </span>

          <div className="min-w-0">
            <h2 className="text-xl font-black">Riwayat terbaru</h2>

            <p className="mt-0.5 text-sm leading-6 text-slate-500 dark:text-slate-400">
              Menampilkan maksimal 30 pemeriksaan terakhir
            </p>
          </div>
        </div>

        <div className="mt-5 overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm sm:rounded-3xl dark:border-white/10 dark:bg-white/[0.03]">
          {scansQuery.isPending ? (
            <LoadingState label="Memuat riwayat pemeriksaan" />
          ) : scansQuery.error ? (
            <div className="p-4 sm:p-5">
              <ErrorState
                message={
                  scansQuery.error instanceof Error
                    ? scansQuery.error.message
                    : "Riwayat tidak dapat dimuat"
                }
              />
            </div>
          ) : scans.length === 0 ? (
            <div className="p-4 sm:p-5">
              <EmptyState
                title="Belum ada riwayat"
                description="Mulai pemeriksaan pertama untuk melihat hasilnya di sini."
              />
            </div>
          ) : (
            <div className="divide-y divide-slate-100 dark:divide-white/5">
              {scans.map((scan) => (
                <Link
                  key={scan.id}
                  href={`/scans/${scan.id}`}
                  className="grid min-w-0 gap-3 px-4 py-4 transition hover:bg-slate-50 sm:px-5 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-center md:gap-5 dark:hover:bg-white/[0.03]"
                >
                  <div className="min-w-0">
                    <p className="truncate font-bold">
                      {scan.page_title || truncate(scan.url, 72)}
                    </p>

                    <p className="mt-1 truncate text-xs text-slate-500 dark:text-slate-400">
                      {scan.url}
                    </p>
                  </div>

                  <StatusBadge status={scan.status} />

                  <div className="flex items-end justify-between gap-4 md:block md:min-w-24 md:text-right">
                    <strong className="block text-lg font-black">
                      {scan.automated_score}
                      <span className="ml-1 text-xs font-semibold text-slate-400">
                        /100
                      </span>
                    </strong>

                    <span className="text-xs text-slate-400">
                      {formatDate(scan.created_at)}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </section>

      {projectDialogOpen && (
        <div className="fixed inset-0 z-[60] overflow-y-auto p-4">
          <div className="grid min-h-full place-items-center">
            <button
              type="button"
              className="fixed inset-0 bg-slate-950/60 backdrop-blur-sm"
              onClick={() => setProjectDialogOpen(false)}
              aria-label="Tutup dialog"
            />

            <div
              role="dialog"
              aria-modal="true"
              aria-labelledby="project-dialog-title"
              className="relative w-full max-w-lg rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-2xl sm:p-7 lg:p-8 dark:border-white/10 dark:bg-slate-900"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-600 sm:text-sm dark:text-blue-400">
                    Project baru
                  </p>

                  <h2
                    id="project-dialog-title"
                    className="mt-2 text-2xl font-black"
                  >
                    Tambahkan website atau aplikasi
                  </h2>
                </div>

                <button
                  type="button"
                  onClick={() => setProjectDialogOpen(false)}
                  className="grid size-10 shrink-0 place-items-center rounded-xl border border-slate-200 transition hover:bg-slate-100 dark:border-white/10 dark:hover:bg-white/5"
                  aria-label="Tutup dialog"
                >
                  <X className="size-5" aria-hidden="true" />
                </button>
              </div>

              <form onSubmit={submitProject} className="mt-7 space-y-5">
                <label className="block">
                  <span className="mb-2 block text-sm font-bold">
                    Nama project
                  </span>

                  <input
                    type="text"
                    value={projectName}
                    onChange={(event) => setProjectName(event.target.value)}
                    minLength={2}
                    maxLength={120}
                    required
                    placeholder="Contoh: Website portofolio"
                    className="h-12 w-full rounded-xl border border-slate-200 bg-white px-4 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                  />
                </label>

                <label className="block">
                  <span className="mb-2 block text-sm font-bold">
                    Deskripsi
                  </span>

                  <textarea
                    value={projectDescription}
                    onChange={(event) =>
                      setProjectDescription(event.target.value)
                    }
                    maxLength={1000}
                    rows={4}
                    placeholder="Jelaskan website atau tujuan pemeriksaannya."
                    className="w-full resize-y rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                  />
                </label>

                <button
                  type="submit"
                  disabled={projectMutation.isPending}
                  className="flex min-h-12 w-full items-center justify-center rounded-xl bg-blue-600 px-5 py-3 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {projectMutation.isPending
                    ? "Sedang menyimpan..."
                    : "Simpan project"}
                </button>
              </form>
            </div>
          </div>
        </div>
      )}
    </AppShell>
  );
}
