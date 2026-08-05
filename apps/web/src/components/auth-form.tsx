"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ArrowRight,
  Eye,
  EyeOff,
  LockKeyhole,
  Mail,
  UserRound,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

import { Brand } from "@/components/brand";
import { loginAccount, registerAccount } from "@/lib/api/services";

const loginSchema = z.object({
  email: z.email("Alamat email tidak valid"),
  password: z.string().min(1, "Password wajib diisi"),
});

const registerSchema = z.object({
  name: z
    .string()
    .trim()
    .min(2, "Nama minimal 2 karakter")
    .max(100, "Nama maksimal 100 karakter"),
  email: z.email("Alamat email tidak valid"),
  password: z
    .string()
    .min(10, "Password minimal 10 karakter")
    .max(72, "Password maksimal 72 karakter"),
});

export function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [visible, setVisible] = useState(false);
  const [formError, setFormError] = useState("");

  const mutation = useMutation({
    mutationFn: async () => {
      if (mode === "register") {
        const parsed = registerSchema.safeParse({
          name,
          email,
          password,
        });

        if (!parsed.success) {
          throw new Error(parsed.error.issues[0]?.message);
        }

        return registerAccount(parsed.data);
      }

      const parsed = loginSchema.safeParse({
        email,
        password,
      });

      if (!parsed.success) {
        throw new Error(parsed.error.issues[0]?.message);
      }

      return loginAccount(parsed.data);
    },
    onSuccess: async (data) => {
      queryClient.setQueryData(["current-user"], data.user);
      await queryClient.invalidateQueries({
        queryKey: ["projects"],
      });
      router.replace("/dashboard");
      router.refresh();
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "Permintaan gagal";

      setFormError(message);
      toast.error(message);
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError("");
    mutation.mutate();
  }

  const registerMode = mode === "register";

  return (
    <main className="relative min-h-screen overflow-hidden bg-slate-950 text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_15%_15%,rgba(13,148,136,0.34),transparent_28%),radial-gradient(circle_at_85%_80%,rgba(217,70,239,0.24),transparent_30%)]" />

      <div className="relative mx-auto grid min-h-screen max-w-7xl lg:grid-cols-[1.05fr_0.95fr]">
        <section className="hidden flex-col justify-between px-12 py-10 lg:flex">
          <Brand className="[&_span_span:first-child]:text-white" />

          <div className="max-w-xl">
            <p className="mb-5 text-sm font-bold uppercase tracking-[0.2em] text-blue-300">
              Audit aksesibilitas yang dapat ditindaklanjuti
            </p>

            <h1 className="text-5xl font-black leading-[1.08] tracking-[-0.04em]">
              Bangun website yang dapat digunakan lebih banyak orang.
            </h1>

            <p className="mt-6 max-w-lg text-lg leading-8 text-slate-300">
              Periksa aturan WCAG otomatis, tinjau DOM yang bermasalah, lengkapi
              pemeriksaan manual, dan ekspor laporan dalam satu workspace.
            </p>

            <div className="mt-10 grid grid-cols-3 gap-4">
              {[
                ["WCAG 2.2", "Level A dan AA"],
                ["axe-core", "Automated audit"],
                ["Go worker", "Isolated scanner"],
              ].map(([title, description]) => (
                <div
                  key={title}
                  className="rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur"
                >
                  <strong className="block text-sm">{title}</strong>
                  <span className="mt-1 block text-xs leading-5 text-slate-400">
                    {description}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <p className="text-xs leading-5 text-slate-500">
            Hasil otomatis tidak membuktikan kepatuhan WCAG secara penuh.
            Pemeriksaan manual tetap diperlukan.
          </p>
        </section>

        <section className="flex items-center justify-center px-4 py-8 sm:px-8">
          <div className="w-full max-w-md rounded-[2rem] border border-white/10 bg-white p-6 text-slate-950 shadow-2xl shadow-black/30 sm:p-9 dark:bg-slate-900 dark:text-white">
            <div className="mb-8 lg:hidden">
              <Brand />
            </div>

            <p className="text-sm font-bold uppercase tracking-[0.18em] text-blue-600 dark:text-blue-400">
              {registerMode ? "Buat workspace" : "Selamat datang kembali"}
            </p>

            <h2 className="mt-3 text-3xl font-black tracking-tight">
              {registerMode ? "Daftar ke AksesCheck" : "Masuk ke akun Anda"}
            </h2>

            <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400">
              {registerMode
                ? "Buat akun untuk menyimpan project, histori scan, dan laporan."
                : "Lanjutkan pemeriksaan aksesibilitas website Anda."}
            </p>

            <form onSubmit={handleSubmit} className="mt-8 space-y-5">
              {registerMode && (
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
                      autoComplete="name"
                      required
                      minLength={2}
                      maxLength={100}
                      placeholder="Rifqi Susanto"
                      className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                    />
                  </span>
                </label>
              )}

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
                    autoComplete="email"
                    required
                    placeholder="nama@email.com"
                    className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                  />
                </span>
              </label>

              <label className="block">
                <span className="mb-2 block text-sm font-bold">Password</span>

                <span className="relative block">
                  <LockKeyhole
                    className="pointer-events-none absolute left-4 top-1/2 size-4.5 -translate-y-1/2 text-slate-400"
                    aria-hidden="true"
                  />

                  <input
                    type={visible ? "text" : "password"}
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                    autoComplete={
                      registerMode ? "new-password" : "current-password"
                    }
                    required
                    minLength={registerMode ? 10 : 1}
                    maxLength={72}
                    placeholder={
                      registerMode ? "Minimal 10 karakter" : "Masukkan password"
                    }
                    className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-12 text-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
                  />

                  <button
                    type="button"
                    onClick={() => setVisible((value) => !value)}
                    className="absolute right-2 top-1/2 grid size-9 -translate-y-1/2 place-items-center rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-white/5"
                    aria-label={
                      visible ? "Sembunyikan password" : "Tampilkan password"
                    }
                  >
                    {visible ? (
                      <EyeOff className="size-4.5" aria-hidden="true" />
                    ) : (
                      <Eye className="size-4.5" aria-hidden="true" />
                    )}
                  </button>
                </span>
              </label>

              {formError && (
                <div
                  role="alert"
                  className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300"
                >
                  {formError}
                </div>
              )}

              <button
                type="submit"
                disabled={mutation.isPending}
                className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {mutation.isPending
                  ? "Memproses..."
                  : registerMode
                    ? "Buat akun"
                    : "Masuk"}

                {!mutation.isPending && (
                  <ArrowRight className="size-4" aria-hidden="true" />
                )}
              </button>
            </form>

            <p className="mt-7 text-center text-sm text-slate-600 dark:text-slate-400">
              {registerMode ? "Sudah memiliki akun?" : "Belum memiliki akun?"}{" "}
              <Link
                href={registerMode ? "/login" : "/register"}
                className="font-bold text-blue-600 hover:text-blue-700 dark:text-blue-400"
              >
                {registerMode ? "Masuk" : "Daftar sekarang"}
              </Link>
            </p>
          </div>
        </section>
      </div>
    </main>
  );
}
