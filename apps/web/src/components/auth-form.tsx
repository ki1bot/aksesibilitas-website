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

import { AuthPageShell } from "@/components/auth-page-shell";
import { loginAccount, registerAccount } from "@/lib/api/services";

const loginSchema = z.object({
  email: z.email("Masukkan alamat email yang valid"),
  password: z.string().min(1, "Password belum diisi"),
});

const registerSchema = z.object({
  name: z
    .string()
    .trim()
    .min(2, "Nama minimal 2 karakter")
    .max(100, "Nama maksimal 100 karakter"),
  email: z.email("Masukkan alamat email yang valid"),
  password: z
    .string()
    .min(10, "Password minimal 10 karakter")
    .max(72, "Password maksimal 72 karakter"),
});

export function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const registerMode = mode === "register";

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [formError, setFormError] = useState("");

  const mutation = useMutation({
    mutationFn: async () => {
      if (registerMode) {
        const parsed = registerSchema.safeParse({
          name,
          email,
          password,
        });

        if (!parsed.success) {
          throw new Error(
            parsed.error.issues[0]?.message ?? "Data pendaftaran belum benar",
          );
        }

        return registerAccount(parsed.data);
      }

      const parsed = loginSchema.safeParse({
        email,
        password,
      });

      if (!parsed.success) {
        throw new Error(
          parsed.error.issues[0]?.message ?? "Email atau password belum benar",
        );
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
        error instanceof Error
          ? error.message
          : "Permintaan tidak dapat diproses";

      setFormError(message);
      toast.error(message);
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError("");
    mutation.mutate();
  }

  return (
    <AuthPageShell
      eyebrow={registerMode ? "Buat akun" : "Masuk ke akun"}
      title={
        registerMode ? "Mulai periksa website Anda" : "Selamat datang kembali"
      }
      description={
        registerMode
          ? "Buat akun untuk menyimpan project, riwayat pemeriksaan, dan laporan Anda."
          : "Masukkan email dan password untuk melanjutkan pekerjaan Anda."
      }
      footer={
        registerMode ? (
          <>
            Sudah punya akun?{" "}
            <Link
              href="/login"
              className="font-bold text-blue-600 transition hover:text-blue-700 dark:text-blue-400"
            >
              Masuk sekarang
            </Link>
          </>
        ) : (
          <>
            Belum punya akun?{" "}
            <Link
              href="/register"
              className="font-bold text-blue-600 transition hover:text-blue-700 dark:text-blue-400"
            >
              Daftar gratis
            </Link>
          </>
        )
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5" noValidate>
        {registerMode && (
          <label className="block">
            <span className="mb-2 block text-sm font-bold">Nama lengkap</span>

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
                placeholder="Rifqi"
                className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
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
              placeholder="Rifqi@email.com"
              className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
            />
          </span>
        </label>

        <label className="block">
          <span className="mb-2 flex items-center justify-between gap-3">
            <span className="text-sm font-bold">Password</span>

            {!registerMode && (
              <Link
                href="/forgot-password"
                className="text-sm font-semibold text-blue-600 transition hover:text-blue-700 dark:text-blue-400"
              >
                Lupa password?
              </Link>
            )}
          </span>

          <span className="relative block">
            <LockKeyhole
              className="pointer-events-none absolute left-4 top-1/2 size-4.5 -translate-y-1/2 text-slate-400"
              aria-hidden="true"
            />

            <input
              type={passwordVisible ? "text" : "password"}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete={registerMode ? "new-password" : "current-password"}
              required
              minLength={registerMode ? 10 : 1}
              maxLength={72}
              placeholder={
                registerMode ? "Minimal 10 karakter" : "Masukkan password"
              }
              className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-12 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
            />

            <button
              type="button"
              onClick={() => setPasswordVisible((value) => !value)}
              className="absolute right-2 top-1/2 grid size-9 -translate-y-1/2 place-items-center rounded-lg text-slate-500 transition hover:bg-slate-100 dark:hover:bg-white/5"
              aria-label={
                passwordVisible ? "Sembunyikan password" : "Tampilkan password"
              }
            >
              {passwordVisible ? (
                <EyeOff className="size-4.5" aria-hidden="true" />
              ) : (
                <Eye className="size-4.5" aria-hidden="true" />
              )}
            </button>
          </span>
        </label>

        {registerMode && (
          <p className="rounded-xl bg-slate-50 px-4 py-3 text-xs leading-5 text-slate-600 dark:bg-white/5 dark:text-slate-400">
            Gunakan minimal 10 karakter. Hindari memakai password yang sama
            dengan akun lain.
          </p>
        )}

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
            ? "Sedang memproses..."
            : registerMode
              ? "Buat akun"
              : "Masuk"}

          {!mutation.isPending && (
            <ArrowRight className="size-4" aria-hidden="true" />
          )}
        </button>
      </form>
    </AuthPageShell>
  );
}
