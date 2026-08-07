"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowRight,
  Eye,
  EyeOff,
  LoaderCircle,
  LockKeyhole,
  Mail,
  UserRound,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

import { AuthPageShell } from "@/components/auth-page-shell";
import {
  getCurrentUser,
  loginAccount,
  registerAccount,
} from "@/lib/api/services";

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

function getSafeReturnPath() {
  if (typeof window === "undefined") {
    return "/dashboard";
  }

  const searchParams = new URLSearchParams(window.location.search);

  const candidate = searchParams.get("next")?.trim() ?? "";

  if (
    candidate === "/dashboard" ||
    candidate.startsWith("/dashboard#") ||
    candidate === "/profile" ||
    candidate === "/change-password" ||
    candidate.startsWith("/projects/") ||
    candidate.startsWith("/scans/")
  ) {
    return candidate;
  }

  return "/dashboard";
}

export function AuthForm({ mode }: { mode: "login" | "register" }) {
  const router = useRouter();

  const queryClient = useQueryClient();

  const registerMode = mode === "register";

  const [name, setName] = useState("");

  const [email, setEmail] = useState("");

  const [password, setPassword] = useState("");

  const [passwordVisible, setPasswordVisible] = useState(false);

  const [formError, setFormError] = useState("");

  const sessionQuery = useQuery({
    queryKey: ["auth-session-check"],
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 0,
    refetchOnMount: "always",
    refetchOnWindowFocus: false,
  });

  useEffect(() => {
    if (!sessionQuery.data) {
      return;
    }

    router.replace(getSafeReturnPath());
  }, [router, sessionQuery.data]);

  const mutation = useMutation({
    mutationFn: async () => {
      const normalizedEmail = email.trim().toLowerCase();

      if (registerMode) {
        const parsed = registerSchema.safeParse({
          name: name.trim(),
          email: normalizedEmail,
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
        email: normalizedEmail,
        password,
      });

      if (!parsed.success) {
        throw new Error(
          parsed.error.issues[0]?.message ?? "Email atau password belum benar",
        );
      }

      return loginAccount(parsed.data);
    },

    onSuccess: (data) => {
      const destination = getSafeReturnPath();

      queryClient.removeQueries();

      queryClient.setQueryData(["current-user"], data.user);

      toast.success(
        registerMode ? "Akun berhasil dibuat" : "Berhasil masuk ke akun",
      );

      router.replace(destination);
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

    if (mutation.isPending) {
      return;
    }

    setFormError("");

    mutation.mutate();
  }

  if (sessionQuery.isPending || sessionQuery.data) {
    return (
      <AuthPageShell
        eyebrow={registerMode ? "Buat akun" : "Masuk ke akun"}
        title="Memeriksa sesi"
        description="Kami sedang memeriksa apakah akun Anda masih memiliki sesi aktif."
        showHomeLink
      >
        <div className="flex min-h-40 flex-col items-center justify-center gap-3 text-center">
          <LoaderCircle
            className="size-8 animate-spin text-blue-600"
            aria-hidden="true"
          />

          <p className="text-sm font-medium text-slate-600 dark:text-slate-300">
            {sessionQuery.data
              ? "Membuka dashboard..."
              : "Memeriksa sesi akun..."}
          </p>
        </div>
      </AuthPageShell>
    );
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
      showHomeLink
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
                disabled={mutation.isPending}
                placeholder="Rifqi"
                className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
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
              disabled={mutation.isPending}
              placeholder="rifqi@email.com"
              className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
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
              disabled={mutation.isPending}
              placeholder={
                registerMode ? "Minimal 10 karakter" : "Masukkan password"
              }
              className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-12 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5"
            />

            <button
              type="button"
              onClick={() => setPasswordVisible((value) => !value)}
              disabled={mutation.isPending}
              className="absolute right-2 top-1/2 grid size-9 -translate-y-1/2 place-items-center rounded-lg text-slate-500 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-white/5"
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
          {mutation.isPending ? (
            <>
              <LoaderCircle
                className="size-4 animate-spin"
                aria-hidden="true"
              />

              {registerMode ? "Membuat akun..." : "Sedang masuk..."}
            </>
          ) : (
            <>
              {registerMode ? "Buat akun" : "Masuk"}

              <ArrowRight className="size-4" aria-hidden="true" />
            </>
          )}
        </button>
      </form>
    </AuthPageShell>
  );
}
