import type { Metadata } from "next";
import { Suspense } from "react";

import { ResetPasswordForm } from "@/components/password-forms";

export const metadata: Metadata = {
  title: "Reset Password",
  description: "Buat password baru untuk akun AksesCheck.",
};

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <main className="grid min-h-screen place-items-center bg-slate-950 px-4 text-white">
          <p className="text-sm font-semibold text-slate-300">
            Memuat halaman...
          </p>
        </main>
      }
    >
      <ResetPasswordForm />
    </Suspense>
  );
}
