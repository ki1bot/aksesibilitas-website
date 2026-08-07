"use client";

import { useQuery } from "@tanstack/react-query";
import {
  ArrowRight,
  KeyRound,
  Mail,
  ShieldCheck,
  UserRound,
} from "lucide-react";
import Link from "next/link";

import { AppShell } from "@/components/app-shell";
import { ErrorState, LoadingState } from "@/components/ui-kit";
import { getCurrentUser } from "@/lib/api/services";

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
        <ProfileContent
          name={userQuery.data.name}
          email={userQuery.data.email}
        />
      )}
    </AppShell>
  );
}

function ProfileContent({ name, email }: { name: string; email: string }) {
  const initial = name.trim().slice(0, 1).toUpperCase() || "U";

  return (
    <section className="mx-auto w-full max-w-5xl">
      <div className="max-w-2xl">
        <p className="text-sm font-bold text-blue-600 dark:text-blue-400">
          Akun
        </p>

        <h1 className="mt-2 text-3xl font-black tracking-tight sm:text-4xl">
          Profil Anda
        </h1>

        <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400 sm:text-base sm:leading-7">
          Lihat informasi akun dan kelola keamanan akun Anda.
        </p>
      </div>

      <div className="mt-8 grid gap-6 lg:grid-cols-[minmax(0,1.1fr)_minmax(320px,0.9fr)]">
        <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm dark:border-white/10 dark:bg-white/[0.03]">
          <div className="border-b border-slate-100 px-5 py-5 dark:border-white/5 sm:px-6">
            <h2 className="text-lg font-black">Informasi profil</h2>

            <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
              Informasi utama yang terhubung dengan akun Anda.
            </p>
          </div>

          <div className="p-5 sm:p-6">
            <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
              <span className="grid size-20 shrink-0 place-items-center rounded-3xl bg-blue-100 text-2xl font-black text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
                {initial}
              </span>

              <div className="min-w-0">
                <h3 className="truncate text-xl font-black sm:text-2xl">
                  {name}
                </h3>

                <p className="mt-1 break-all text-sm text-slate-500 dark:text-slate-400">
                  {email}
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
                    {name}
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
                    {email}
                  </p>
                </div>
              </div>
            </div>
          </div>
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
