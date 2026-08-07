"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Eye,
  EyeOff,
  KeyRound,
  LockKeyhole,
  Mail,
  ShieldCheck,
} from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

import { AppShell } from "@/components/app-shell";
import { AuthPageShell } from "@/components/auth-page-shell";
import {
  changePassword,
  requestPasswordReset,
  resetPassword,
} from "@/lib/api/password";

const emailSchema = z.email("Masukkan alamat email yang valid");

const newPasswordSchema = z
  .string()
  .min(8, "Password minimal 8 karakter")
  .max(72, "Password maksimal 72 karakter")
  .regex(/^[A-Za-z0-9]+$/, "Password hanya boleh berisi huruf dan angka");

function PasswordInput({
  label,
  value,
  onChange,
  autoComplete,
  placeholder,
  enforceNewPasswordPolicy = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  autoComplete: string;
  placeholder: string;
  enforceNewPasswordPolicy?: boolean;
}) {
  const [visible, setVisible] = useState(false);

  return (
    <label className="block">
      <span className="mb-2 block text-sm font-bold">{label}</span>

      <span className="relative block">
        <LockKeyhole
          className="pointer-events-none absolute left-4 top-1/2 size-4.5 -translate-y-1/2 text-slate-400"
          aria-hidden="true"
        />

        <input
          type={visible ? "text" : "password"}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          autoComplete={autoComplete}
          required
          minLength={enforceNewPasswordPolicy ? 8 : 1}
          maxLength={72}
          pattern={enforceNewPasswordPolicy ? "[A-Za-z0-9]+" : undefined}
          placeholder={placeholder}
          className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-12 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
        />

        <button
          type="button"
          onClick={() => setVisible((current) => !current)}
          className="absolute right-2 top-1/2 grid size-9 -translate-y-1/2 place-items-center rounded-lg text-slate-500 transition hover:bg-slate-100 dark:hover:bg-white/5"
          aria-label={visible ? "Sembunyikan password" : "Tampilkan password"}
        >
          {visible ? (
            <EyeOff className="size-4.5" aria-hidden="true" />
          ) : (
            <Eye className="size-4.5" aria-hidden="true" />
          )}
        </button>
      </span>
    </label>
  );
}

function FormError({ message }: { message: string }) {
  if (!message) {
    return null;
  }

  return (
    <div
      role="alert"
      className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300"
    >
      {message}
    </div>
  );
}

export function ForgotPasswordForm() {
  const [email, setEmail] = useState("");
  const [formError, setFormError] = useState("");
  const [message, setMessage] = useState("");
  const [debugResetURL, setDebugResetURL] = useState("");

  const mutation = useMutation({
    mutationFn: async () => {
      const parsed = emailSchema.safeParse(email.trim());

      if (!parsed.success) {
        throw new Error(parsed.error.issues[0]?.message);
      }

      return requestPasswordReset({
        email: parsed.data,
      });
    },

    onSuccess: (result) => {
      setFormError("");
      setMessage(result.message);
      setDebugResetURL(result.debug_reset_url ?? "");
    },

    onError: (error) => {
      setMessage("");
      setDebugResetURL("");

      setFormError(
        error instanceof Error
          ? error.message
          : "Permintaan tidak dapat diproses",
      );
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError("");
    mutation.mutate();
  }

  return (
    <AuthPageShell
      eyebrow="Lupa password"
      title="Buat tautan pemulihan"
      description="Masukkan email akun Anda. Kami akan mengirim tautan untuk membuat password baru."
      footer={
        <Link
          href="/login"
          className="inline-flex items-center gap-2 font-bold text-blue-600 transition hover:text-blue-700 dark:text-blue-400"
        >
          <ArrowLeft className="size-4" aria-hidden="true" />
          Kembali ke halaman masuk
        </Link>
      }
    >
      {message ? (
        <div className="space-y-5">
          <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-5 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200">
            <CheckCircle2 className="size-7" aria-hidden="true" />

            <h3 className="mt-4 font-black">Periksa email Anda</h3>

            <p className="mt-2 text-sm leading-6">{message}</p>
          </div>

          {debugResetURL && (
            <div className="rounded-2xl border border-amber-200 bg-amber-50 p-5 dark:border-amber-400/20 dark:bg-amber-400/10">
              <p className="text-sm font-bold text-amber-900 dark:text-amber-200">
                Mode pengembangan
              </p>

              <p className="mt-2 text-sm leading-6 text-amber-800 dark:text-amber-300">
                SMTP belum diperlukan saat pengembangan. Gunakan tautan berikut
                untuk menguji halaman reset password.
              </p>

              <Link
                href={debugResetURL}
                className="mt-4 inline-flex items-center gap-2 text-sm font-bold text-blue-700 hover:underline dark:text-blue-300"
              >
                Buka halaman reset password
                <ArrowRight className="size-4" aria-hidden="true" />
              </Link>
            </div>
          )}

          <button
            type="button"
            onClick={() => {
              setMessage("");
              setDebugResetURL("");
            }}
            className="h-11 w-full rounded-xl border border-slate-200 bg-white px-4 text-sm font-bold text-slate-700 transition hover:bg-slate-50 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
          >
            Gunakan email lain
          </button>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-5" noValidate>
          <label className="block">
            <span className="mb-2 block text-sm font-bold">Email akun</span>

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
                className="h-12 w-full rounded-xl border border-slate-200 bg-white pl-11 pr-4 text-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 dark:border-white/10 dark:bg-white/5"
              />
            </span>
          </label>

          <FormError message={formError} />

          <button
            type="submit"
            disabled={mutation.isPending}
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {mutation.isPending
              ? "Sedang mengirim..."
              : "Kirim tautan pemulihan"}

            {!mutation.isPending && (
              <ArrowRight className="size-4" aria-hidden="true" />
            )}
          </button>
        </form>
      )}
    </AuthPageShell>
  );
}

export function ResetPasswordForm() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token")?.trim() ?? "";

  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [formError, setFormError] = useState("");
  const [completed, setCompleted] = useState(false);

  const mutation = useMutation({
    mutationFn: async () => {
      if (!token) {
        throw new Error("Tautan reset password tidak lengkap");
      }

      const parsedPassword = newPasswordSchema.safeParse(password);

      if (!parsedPassword.success) {
        throw new Error(parsedPassword.error.issues[0]?.message);
      }

      if (password !== confirmation) {
        throw new Error("Konfirmasi password belum sama");
      }

      return resetPassword({
        token,
        password: parsedPassword.data,
        password_confirmation: confirmation,
      });
    },

    onSuccess: () => {
      setFormError("");
      setCompleted(true);
    },

    onError: (error) => {
      setFormError(
        error instanceof Error ? error.message : "Password tidak dapat diubah",
      );
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError("");
    mutation.mutate();
  }

  return (
    <AuthPageShell
      eyebrow="Reset password"
      title={completed ? "Password sudah diperbarui" : "Buat password baru"}
      description={
        completed
          ? "Semua sesi lama sudah dikeluarkan agar akun Anda tetap aman."
          : "Gunakan password baru minimal 8 karakter yang hanya berisi huruf atau angka."
      }
      footer={
        <Link
          href="/login"
          className="inline-flex items-center gap-2 font-bold text-blue-600 transition hover:text-blue-700 dark:text-blue-400"
        >
          <ArrowLeft className="size-4" aria-hidden="true" />
          Kembali ke halaman masuk
        </Link>
      }
    >
      {completed ? (
        <div className="space-y-5">
          <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-5 text-emerald-800 dark:border-emerald-400/20 dark:bg-emerald-400/10 dark:text-emerald-200">
            <CheckCircle2 className="size-8" aria-hidden="true" />

            <p className="mt-4 text-sm leading-6">
              Anda sekarang dapat masuk menggunakan password yang baru.
            </p>
          </div>

          <Link
            href="/login"
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700"
          >
            Masuk ke akun
            <ArrowRight className="size-4" aria-hidden="true" />
          </Link>
        </div>
      ) : token ? (
        <form onSubmit={handleSubmit} className="space-y-5" noValidate>
          <PasswordInput
            label="Password baru"
            value={password}
            onChange={setPassword}
            autoComplete="new-password"
            placeholder="Minimal 8 karakter"
            enforceNewPasswordPolicy
          />

          <PasswordInput
            label="Ulangi password baru"
            value={confirmation}
            onChange={setConfirmation}
            autoComplete="new-password"
            placeholder="Ketik ulang password baru"
            enforceNewPasswordPolicy
          />

          <div className="rounded-xl bg-slate-50 px-4 py-3 text-xs leading-5 text-slate-600 dark:bg-white/5 dark:text-slate-400">
            Password minimal 8 karakter dan hanya boleh berisi huruf atau angka.
            Setelah password diubah, semua perangkat yang masih login akan
            otomatis dikeluarkan.
          </div>

          <FormError message={formError} />

          <button
            type="submit"
            disabled={mutation.isPending}
            className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {mutation.isPending
              ? "Sedang menyimpan..."
              : "Simpan password baru"}

            {!mutation.isPending && (
              <ArrowRight className="size-4" aria-hidden="true" />
            )}
          </button>
        </form>
      ) : (
        <div className="rounded-2xl border border-red-200 bg-red-50 p-5 text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-300">
          <KeyRound className="size-8" aria-hidden="true" />

          <h3 className="mt-4 font-black">Tautan tidak lengkap</h3>

          <p className="mt-2 text-sm leading-6">
            Minta tautan baru agar Anda dapat melanjutkan proses reset password.
          </p>

          <Link
            href="/forgot-password"
            className="mt-5 inline-flex items-center gap-2 text-sm font-bold underline"
          >
            Minta tautan baru
            <ArrowRight className="size-4" aria-hidden="true" />
          </Link>
        </div>
      )}
    </AuthPageShell>
  );
}

export function ChangePasswordView() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [formError, setFormError] = useState("");

  const mutation = useMutation({
    mutationFn: async () => {
      if (!currentPassword) {
        throw new Error("Password saat ini belum diisi");
      }

      const parsedPassword = newPasswordSchema.safeParse(newPassword);

      if (!parsedPassword.success) {
        throw new Error(parsedPassword.error.issues[0]?.message);
      }

      if (newPassword !== confirmation) {
        throw new Error("Konfirmasi password belum sama");
      }

      return changePassword({
        current_password: currentPassword,
        new_password: parsedPassword.data,
        password_confirmation: confirmation,
      });
    },

    onSuccess: () => {
      queryClient.clear();

      toast.success("Password berhasil diubah. Silakan masuk kembali");

      router.replace("/login");
      router.refresh();
    },

    onError: (error) => {
      setFormError(
        error instanceof Error ? error.message : "Password tidak dapat diubah",
      );
    },
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFormError("");
    mutation.mutate();
  }

  return (
    <AppShell>
      <section className="mx-auto max-w-3xl">
        <div className="flex items-start gap-4">
          <span className="grid size-12 shrink-0 place-items-center rounded-2xl bg-blue-50 text-blue-600 dark:bg-blue-400/10 dark:text-blue-300">
            <ShieldCheck className="size-5" aria-hidden="true" />
          </span>

          <div>
            <p className="text-sm font-bold uppercase tracking-[0.16em] text-blue-600 dark:text-blue-400">
              Keamanan akun
            </p>

            <h1 className="mt-2 text-3xl font-black tracking-tight sm:text-4xl">
              Ganti password
            </h1>

            <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-600 dark:text-slate-400 sm:text-base">
              Setelah password diubah, Anda akan dikeluarkan dari semua
              perangkat dan perlu masuk kembali.
            </p>
          </div>
        </div>

        <div className="mt-8 grid gap-6 lg:grid-cols-[1fr_0.7fr]">
          <form
            onSubmit={handleSubmit}
            className="space-y-5 rounded-[2rem] border border-slate-200 bg-white p-6 shadow-sm dark:border-white/10 dark:bg-white/[0.03] sm:p-8"
            noValidate
          >
            <PasswordInput
              label="Password saat ini"
              value={currentPassword}
              onChange={setCurrentPassword}
              autoComplete="current-password"
              placeholder="Masukkan password saat ini"
            />

            <PasswordInput
              label="Password baru"
              value={newPassword}
              onChange={setNewPassword}
              autoComplete="new-password"
              placeholder="Minimal 8 karakter"
              enforceNewPasswordPolicy
            />

            <PasswordInput
              label="Ulangi password baru"
              value={confirmation}
              onChange={setConfirmation}
              autoComplete="new-password"
              placeholder="Ketik ulang password baru"
              enforceNewPasswordPolicy
            />

            <FormError message={formError} />

            <div className="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:justify-end">
              <Link
                href="/profile"
                className="inline-flex h-11 items-center justify-center rounded-xl border border-slate-200 bg-white px-5 text-sm font-bold text-slate-700 transition hover:bg-slate-50 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
              >
                Batal
              </Link>

              <button
                type="submit"
                disabled={mutation.isPending}
                className="inline-flex h-11 items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {mutation.isPending ? "Sedang menyimpan..." : "Ganti password"}

                {!mutation.isPending && (
                  <ArrowRight className="size-4" aria-hidden="true" />
                )}
              </button>
            </div>
          </form>

          <aside className="h-fit rounded-[2rem] border border-slate-200 bg-slate-50 p-6 dark:border-white/10 dark:bg-white/[0.03]">
            <KeyRound
              className="size-6 text-blue-600 dark:text-blue-400"
              aria-hidden="true"
            />

            <h2 className="mt-4 font-black">Ketentuan password</h2>

            <ul className="mt-4 space-y-3 text-sm leading-6 text-slate-600 dark:text-slate-400">
              <li>Gunakan minimal 8 karakter.</li>
              <li>Password hanya boleh berisi huruf atau angka.</li>
              <li>
                Huruf dan angka boleh digunakan sendiri atau dikombinasikan.
              </li>
              <li>Jangan gunakan password yang mudah ditebak.</li>
            </ul>
          </aside>
        </div>
      </section>
    </AppShell>
  );
}
