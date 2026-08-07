"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  ArrowRight,
  Eye,
  EyeOff,
  KeyRound,
  LockKeyhole,
  Mail,
  Pencil,
  Save,
  ShieldCheck,
  UserRound,
  X,
} from "lucide-react";
import Link from "next/link";
import { FormEvent, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

import { AppShell } from "@/components/app-shell";
import { ErrorState, LoadingState } from "@/components/ui-kit";
import { updateCurrentUser } from "@/lib/api/profile";
import { getCurrentUser } from "@/lib/api/services";
import type { User } from "@/lib/api/types";

const profileSchema = z.object({
  name: z
    .string()
    .trim()
    .min(2, "Nama minimal 2 karakter")
    .max(100, "Nama maksimal 100 karakter"),
  email: z.email("Masukkan alamat email yang valid"),
});

export function ProfileView() {
  const userQuery = useQuery({
    queryKey: ["current-user"],
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 15_000,
  });

  return (
    <AppShell>
      {userQuery.isPending ? (
        <LoadingState label="Memuat profil" />
      ) : userQuery.error || !userQuery.data ? (
        <ErrorState
          message={
            userQuery.error instanceof Error
              ? userQuery.error.message
              : "Profil tidak dapat dimuat"
          }
        />
      ) : (
        <ProfileContent user={userQuery.data} />
      )}
    </AppShell>
  );
}

function ProfileContent({ user }: { user: User }) {
  const queryClient = useQueryClient();

  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(user.name);
  const [email, setEmail] = useState(user.email);
  const [currentPassword, setCurrentPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [formError, setFormError] = useState("");

  const normalizedEmail = email.trim().toLowerCase();

  const emailChanged = normalizedEmail !== user.email.trim().toLowerCase();

  const mutation = useMutation({
    mutationFn: async () => {
      const parsed = profileSchema.safeParse({
        name,
        email: normalizedEmail,
      });

      if (!parsed.success) {
        throw new Error(
          parsed.error.issues[0]?.message ?? "Data profil tidak valid",
        );
      }

      if (emailChanged && !currentPassword) {
        throw new Error("Masukkan password saat ini untuk mengganti email");
      }

      return updateCurrentUser({
        name: parsed.data.name,
        email: parsed.data.email,
        current_password: emailChanged ? currentPassword : "",
      });
    },

    onSuccess: (updatedUser) => {
      queryClient.setQueryData(["current-user"], updatedUser);

      setName(updatedUser.name);
      setEmail(updatedUser.email);
      setCurrentPassword("");
      setFormError("");
      setEditing(false);

      toast.success("Profil berhasil diperbarui");
    },

    onError: (error) => {
      const message =
        error instanceof Error
          ? error.message
          : "Profil tidak dapat diperbarui";

      setFormError(message);
      toast.error(message);
    },
  });

  const initial = user.name.trim().slice(0, 1).toUpperCase() || "U";

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (mutation.isPending) {
      return;
    }

    setFormError("");
    mutation.mutate();
  }

  function startEditing() {
    setName(user.name);
    setEmail(user.email);
    setCurrentPassword("");
    setFormError("");
    setEditing(true);
  }

  function cancelEditing() {
    setName(user.name);
    setEmail(user.email);
    setCurrentPassword("");
    setFormError("");
    setEditing(false);
  }

  return (
    <section className="mx-auto w-full max-w-5xl">
      <Link
        href="/dashboard"
        className="group inline-flex min-h-10 items-center gap-2 rounded-xl px-1 text-sm font-bold text-slate-600 transition hover:text-blue-600 dark:text-slate-400 dark:hover:text-blue-400"
      >
        <ArrowLeft
          className="size-4 transition-transform group-hover:-translate-x-1"
          aria-hidden="true"
        />
        Kembali ke dashboard
      </Link>

      <div className="mt-5 flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div className="max-w-2xl">
          <p className="text-sm font-bold text-blue-600 dark:text-blue-400">
            Akun
          </p>

          <h1 className="mt-2 text-3xl font-black tracking-tight sm:text-4xl">
            Profil Anda
          </h1>

          <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400 sm:text-base sm:leading-7">
            Kelola nama, alamat email, dan keamanan akun Anda.
          </p>
        </div>

        {!editing && (
          <button
            type="button"
            onClick={startEditing}
            className="inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 shadow-sm transition hover:border-blue-200 hover:bg-blue-50 hover:text-blue-700 sm:w-auto dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:border-blue-400/20 dark:hover:bg-blue-400/10 dark:hover:text-blue-300"
          >
            <Pencil className="size-4" aria-hidden="true" />
            Edit profil
          </button>
        )}
      </div>

      <div className="mt-8 grid gap-6 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
        <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/[0.03]">
          <div className="border-b border-slate-100 px-5 py-5 dark:border-white/5 sm:px-6">
            <h2 className="text-lg font-black">Informasi profil</h2>

            <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
              Informasi utama yang terhubung dengan akun Anda.
            </p>
          </div>

          {editing ? (
            <form onSubmit={handleSubmit} className="p-5 sm:p-6" noValidate>
              <div className="flex items-center gap-4">
                <span className="grid size-16 shrink-0 place-items-center rounded-2xl bg-blue-100 text-xl font-black text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
                  {name.trim().slice(0, 1).toUpperCase() || initial}
                </span>

                <div>
                  <h3 className="font-black">Edit informasi akun</h3>

                  <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
                    Simpan perubahan setelah data sudah benar.
                  </p>
                </div>
              </div>

              <div className="mt-7 grid gap-5">
                <label className="block">
                  <span className="mb-2 block text-sm font-bold">
                    Nama lengkap
                  </span>

                  <span className="relative block">
                    <UserRound
                      className="pointer-events-none absolute left-4 top-1/2 size-4.5 -translate-y-1/2 text-slate-400"
                      aria-hidden="true"
                    />

                    <input
                      type="text"
                      value={name}
                      onChange={(event) => setName(event.target.value)}
                      required
                      minLength={2}
                      maxLength={100}
                      disabled={mutation.isPending}
                      autoComplete="name"
                      className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
                    />
                  </span>
                </label>

                <label className="block">
                  <span className="mb-2 block text-sm font-bold">Email</span>

                  <span className="relative block">
                    <Mail
                      className="pointer-events-none absolute left-4 top-1/2 size-4.5 -translate-y-1/2 text-slate-400"
                      aria-hidden="true"
                    />

                    <input
                      type="email"
                      value={email}
                      onChange={(event) => setEmail(event.target.value)}
                      required
                      maxLength={255}
                      disabled={mutation.isPending}
                      autoComplete="email"
                      className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
                    />
                  </span>
                </label>

                {emailChanged && (
                  <label className="block">
                    <span className="mb-2 block text-sm font-bold">
                      Password saat ini
                    </span>

                    <span className="relative block">
                      <LockKeyhole
                        className="pointer-events-none absolute left-4 top-1/2 size-4.5 -translate-y-1/2 text-slate-400"
                        aria-hidden="true"
                      />

                      <input
                        type={passwordVisible ? "text" : "password"}
                        value={currentPassword}
                        onChange={(event) =>
                          setCurrentPassword(event.target.value)
                        }
                        required
                        maxLength={72}
                        disabled={mutation.isPending}
                        autoComplete="current-password"
                        placeholder="Masukkan password untuk mengganti email"
                        className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-12 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
                      />

                      <button
                        type="button"
                        onClick={() => setPasswordVisible((value) => !value)}
                        disabled={mutation.isPending}
                        className="absolute right-2 top-1/2 grid size-9 -translate-y-1/2 place-items-center rounded-lg text-slate-500 transition hover:bg-slate-100 disabled:opacity-50 dark:hover:bg-white/5"
                        aria-label={
                          passwordVisible
                            ? "Sembunyikan password"
                            : "Tampilkan password"
                        }
                      >
                        {passwordVisible ? (
                          <EyeOff className="size-4.5" aria-hidden="true" />
                        ) : (
                          <Eye className="size-4.5" aria-hidden="true" />
                        )}
                      </button>
                    </span>

                    <p className="mt-2 text-xs leading-5 text-slate-500 dark:text-slate-400">
                      Password hanya diperlukan jika alamat email diubah.
                    </p>
                  </label>
                )}
              </div>

              {formError && (
                <div
                  role="alert"
                  className="mt-5 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300"
                >
                  {formError}
                </div>
              )}

              <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                <button
                  type="button"
                  onClick={cancelEditing}
                  disabled={mutation.isPending}
                  className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50 disabled:opacity-60 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
                >
                  <X className="size-4" aria-hidden="true" />
                  Batal
                </button>

                <button
                  type="submit"
                  disabled={mutation.isPending}
                  className="inline-flex min-h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-2.5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  <Save className="size-4" aria-hidden="true" />

                  {mutation.isPending ? "Menyimpan..." : "Simpan perubahan"}
                </button>
              </div>
            </form>
          ) : (
            <div className="p-5 sm:p-6">
              <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
                <span className="grid size-20 shrink-0 place-items-center rounded-3xl bg-blue-100 text-2xl font-black text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
                  {initial}
                </span>

                <div className="min-w-0">
                  <h3 className="truncate text-xl font-black sm:text-2xl">
                    {user.name}
                  </h3>

                  <p className="mt-1 break-all text-sm text-slate-500 dark:text-slate-400">
                    {user.email}
                  </p>
                </div>
              </div>

              <div className="mt-8 grid gap-3">
                <div className="flex items-start gap-4 rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-white/10 dark:bg-white/[0.03]">
                  <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-white text-slate-500 shadow-sm dark:bg-white/5 dark:text-slate-300">
                    <UserRound className="size-5" aria-hidden="true" />
                  </span>

                  <div className="min-w-0">
                    <p className="text-xs font-bold uppercase tracking-[0.12em] text-slate-400">
                      Nama
                    </p>

                    <p className="mt-1 break-words text-sm font-bold text-slate-900 dark:text-white">
                      {user.name}
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-4 rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-white/10 dark:bg-white/[0.03]">
                  <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-white text-slate-500 shadow-sm dark:bg-white/5 dark:text-slate-300">
                    <Mail className="size-5" aria-hidden="true" />
                  </span>

                  <div className="min-w-0">
                    <p className="text-xs font-bold uppercase tracking-[0.12em] text-slate-400">
                      Email
                    </p>

                    <p className="mt-1 break-all text-sm font-bold text-slate-900 dark:text-white">
                      {user.email}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="self-start rounded-3xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.03] sm:p-6">
          <span className="grid size-12 place-items-center rounded-2xl bg-blue-50 text-blue-600 dark:bg-blue-400/10 dark:text-blue-300">
            <ShieldCheck className="size-6" aria-hidden="true" />
          </span>

          <h2 className="mt-5 text-xl font-black">Keamanan akun</h2>

          <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-400">
            Kelola password untuk menjaga keamanan dan akses ke akun Anda.
          </p>

          <Link
            href="/change-password"
            className="mt-6 flex min-h-12 w-full items-center justify-between gap-4 rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-left transition hover:border-blue-300 hover:bg-blue-50 dark:border-white/10 dark:bg-white/[0.03] dark:hover:border-blue-400/30 dark:hover:bg-blue-400/10"
          >
            <span className="flex min-w-0 items-center gap-3">
              <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-white text-blue-600 shadow-sm dark:bg-white/5 dark:text-blue-300">
                <KeyRound className="size-5" aria-hidden="true" />
              </span>

              <span className="min-w-0">
                <span className="block text-sm font-black">Ganti Password</span>

                <span className="mt-0.5 block text-xs leading-5 text-slate-500 dark:text-slate-400">
                  Perbarui password akun
                </span>
              </span>
            </span>

            <ArrowRight
              className="size-4 shrink-0 text-slate-400"
              aria-hidden="true"
            />
          </Link>
        </div>
      </div>
    </section>
  );
}
