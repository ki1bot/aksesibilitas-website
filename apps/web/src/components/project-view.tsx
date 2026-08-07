"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  ExternalLink,
  Globe2,
  Pencil,
  Play,
  Save,
  Trash2,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
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
  createScan,
  deleteProject,
  getProject,
  listScans,
  updateProject,
} from "@/lib/api/services";
import type { ProjectRequest } from "@/lib/api/types";
import { formatDate, truncate } from "@/lib/format";

const projectSchema = z.object({
  name: z.string().trim().min(2).max(120),
  description: z.string().trim().max(1000),
});

const urlSchema = z.url("URL website tidak valid");

export function ProjectView({ projectId }: { projectId: string }) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const [editing, setEditing] = useState(false);
  const [scanURL, setScanURL] = useState("");

  const projectQuery = useQuery({
    queryKey: ["project", projectId],
    queryFn: () => getProject(projectId),
  });

  const scansQuery = useQuery({
    queryKey: ["scans", "project", projectId],
    queryFn: () => listScans(projectId, 100),
  });

  const updateMutation = useMutation({
    mutationFn: (input: ProjectRequest) => updateProject(projectId, input),
    onSuccess: async () => {
      setEditing(false);

      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["project", projectId],
        }),
        queryClient.invalidateQueries({
          queryKey: ["projects"],
        }),
      ]);

      toast.success("Project berhasil diperbarui");
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : "Gagal memperbarui project",
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteProject(projectId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["projects"],
      });

      router.replace("/dashboard");
      toast.success("Project berhasil dihapus");
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : "Gagal menghapus project",
      );
    },
  });

  const scanMutation = useMutation({
    mutationFn: () => {
      const parsed = urlSchema.safeParse(scanURL.trim());

      if (!parsed.success) {
        throw new Error(
          parsed.error.issues[0]?.message ?? "URL website tidak valid",
        );
      }

      return createScan({
        project_id: projectId,
        url: parsed.data,
      });
    },
    onSuccess: (scan) => {
      router.push(`/scans/${scan.id}`);
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : "Gagal membuat scan",
      );
    },
  });

  function submitProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const formData = new FormData(event.currentTarget);

    const parsed = projectSchema.safeParse({
      name: String(formData.get("name") ?? ""),
      description: String(formData.get("description") ?? ""),
    });

    if (!parsed.success) {
      toast.error(
        parsed.error.issues[0]?.message ?? "Data project tidak valid",
      );
      return;
    }

    updateMutation.mutate(parsed.data);
  }

  function submitScan(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    scanMutation.mutate();
  }

  function handleDelete() {
    const confirmed = window.confirm(
      "Project dan seluruh data scan di dalamnya akan dihapus permanen. Lanjutkan?",
    );

    if (confirmed) {
      deleteMutation.mutate();
    }
  }

  const project = projectQuery.data;

  return (
    <AppShell>
      <Link
        href="/dashboard"
        className="inline-flex min-h-10 items-center gap-2 rounded-xl px-1 text-sm font-bold text-slate-600 transition hover:text-blue-600 dark:text-slate-300 dark:hover:text-blue-400"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        Kembali ke dashboard
      </Link>

      {projectQuery.isPending ? (
        <LoadingState label="Memuat project" />
      ) : projectQuery.error || !project ? (
        <div className="mt-6 sm:mt-8">
          <ErrorState
            message={
              projectQuery.error instanceof Error
                ? projectQuery.error.message
                : "Project tidak ditemukan"
            }
          />
        </div>
      ) : (
        <>
          <section className="mt-5 rounded-[1.75rem] border border-slate-200 bg-white p-5 shadow-sm sm:mt-7 sm:p-7 lg:p-8 dark:border-white/10 dark:bg-white/[0.03]">
            {editing ? (
              <form key={project.updated_at} onSubmit={submitProject}>
                <div className="grid gap-5">
                  <label>
                    <span className="mb-2 block text-sm font-bold">
                      Nama project
                    </span>

                    <input
                      type="text"
                      name="name"
                      defaultValue={project.name}
                      required
                      minLength={2}
                      maxLength={120}
                      className="h-12 w-full rounded-xl border border-slate-200 bg-white px-4 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                    />
                  </label>

                  <label>
                    <span className="mb-2 block text-sm font-bold">
                      Deskripsi
                    </span>

                    <textarea
                      name="description"
                      defaultValue={project.description}
                      maxLength={1000}
                      rows={4}
                      className="w-full resize-y rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                    />
                  </label>
                </div>

                <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:flex-wrap">
                  <button
                    type="submit"
                    disabled={updateMutation.isPending}
                    className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-bold text-white transition hover:bg-blue-700 disabled:opacity-60"
                  >
                    <Save className="size-4" aria-hidden="true" />

                    {updateMutation.isPending
                      ? "Menyimpan..."
                      : "Simpan perubahan"}
                  </button>

                  <button
                    type="button"
                    onClick={() => setEditing(false)}
                    className="min-h-11 rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-bold transition hover:bg-slate-100 dark:border-white/10 dark:hover:bg-white/5"
                  >
                    Batal
                  </button>
                </div>
              </form>
            ) : (
              <div className="flex flex-col gap-6 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <p className="text-xs font-bold uppercase tracking-[0.18em] text-blue-600 sm:text-sm dark:text-blue-400">
                    Project
                  </p>

                  <h1 className="mt-2 break-words text-3xl font-black tracking-tight sm:text-4xl">
                    {project.name}
                  </h1>

                  <p className="mt-3 max-w-3xl text-sm leading-6 text-slate-600 sm:text-base sm:leading-7 dark:text-slate-400">
                    {project.description ||
                      "Project ini belum memiliki deskripsi."}
                  </p>

                  <p className="mt-4 text-xs font-medium text-slate-400">
                    Dibuat {formatDate(project.created_at)}
                  </p>
                </div>

                <div className="flex shrink-0 flex-col gap-2 sm:flex-row sm:flex-wrap">
                  <button
                    type="button"
                    onClick={() => setEditing(true)}
                    className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold transition hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:hover:bg-white/10"
                  >
                    <Pencil className="size-4" aria-hidden="true" />
                    Edit
                  </button>

                  <button
                    type="button"
                    onClick={handleDelete}
                    disabled={deleteMutation.isPending}
                    className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-2.5 text-sm font-bold text-red-700 transition hover:bg-red-100 disabled:opacity-60 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300 dark:hover:bg-red-400/15"
                  >
                    <Trash2 className="size-4" aria-hidden="true" />
                    {deleteMutation.isPending ? "Menghapus..." : "Hapus"}
                  </button>
                </div>
              </div>
            )}
          </section>

          <section className="mt-5 overflow-hidden rounded-[1.75rem] bg-slate-950 p-5 text-white shadow-xl sm:mt-6 sm:p-7 lg:p-8">
            <div className="flex items-start gap-3">
              <span className="grid size-11 shrink-0 place-items-center rounded-xl bg-blue-600 sm:rounded-2xl">
                <Globe2 className="size-5" aria-hidden="true" />
              </span>

              <div className="min-w-0">
                <h2 className="text-lg font-black sm:text-xl">Scan website</h2>

                <p className="mt-1 text-sm leading-6 text-slate-400">
                  Masukkan URL publik dengan protokol HTTP atau HTTPS.
                </p>
              </div>
            </div>

            <form
              onSubmit={submitScan}
              className="mt-6 grid min-w-0 gap-3 sm:grid-cols-[minmax(0,1fr)_auto]"
            >
              <input
                type="url"
                value={scanURL}
                onChange={(event) => setScanURL(event.target.value)}
                required
                maxLength={2048}
                placeholder="https://example.com"
                className="h-12 w-full min-w-0 rounded-xl border border-white/10 bg-white/10 px-4 text-sm outline-none placeholder:text-slate-500 focus:border-blue-400 focus:ring-4 focus:ring-blue-400/15"
              />

              <button
                type="submit"
                disabled={scanMutation.isPending}
                className="inline-flex h-12 items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold transition hover:bg-blue-500 disabled:opacity-60"
              >
                <Play className="size-4" aria-hidden="true" />

                {scanMutation.isPending ? "Mengirim..." : "Mulai scan"}
              </button>
            </form>
          </section>

          <section className="mt-9 sm:mt-10">
            <h2 className="text-xl font-black">Histori project</h2>

            <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
              Seluruh pemindaian yang dilakukan dalam project ini.
            </p>

            <div className="mt-5">
              {scansQuery.isPending ? (
                <LoadingState label="Memuat histori project" />
              ) : scansQuery.error ? (
                <ErrorState
                  message={
                    scansQuery.error instanceof Error
                      ? scansQuery.error.message
                      : "Gagal mengambil histori"
                  }
                />
              ) : !scansQuery.data?.length ? (
                <EmptyState
                  title="Belum ada pemindaian"
                  description="Masukkan URL di atas untuk memulai audit pertama pada project ini."
                />
              ) : (
                <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm sm:rounded-3xl dark:border-white/10 dark:bg-white/[0.03]">
                  <div className="divide-y divide-slate-100 dark:divide-white/5">
                    {scansQuery.data.map((scan) => (
                      <Link
                        key={scan.id}
                        href={`/scans/${scan.id}`}
                        className="grid min-w-0 gap-3 px-4 py-4 transition hover:bg-slate-50 sm:px-5 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto] lg:items-center lg:gap-5 dark:hover:bg-white/[0.03]"
                      >
                        <div className="min-w-0">
                          <p className="truncate font-bold">
                            {scan.page_title || truncate(scan.url, 70)}
                          </p>

                          <p className="mt-1 truncate text-xs text-slate-500 dark:text-slate-400">
                            {scan.url}
                          </p>
                        </div>

                        <StatusBadge status={scan.status} />

                        <div>
                          <strong className="text-lg font-black">
                            {scan.automated_score}
                          </strong>

                          <span className="ml-1 text-xs text-slate-400">
                            /100
                          </span>
                        </div>

                        <span className="inline-flex items-center gap-2 text-xs font-semibold text-slate-400">
                          {formatDate(scan.created_at)}

                          <ExternalLink
                            className="size-3.5"
                            aria-hidden="true"
                          />
                        </span>
                      </Link>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </section>
        </>
      )}
    </AppShell>
  );
}
