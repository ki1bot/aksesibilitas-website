import type { Metadata } from "next";

import { ChangePasswordView } from "@/components/password-forms";

export const metadata: Metadata = {
  title: "Ganti Password",
  description: "Perbarui password akun AksesCheck Anda.",
};

export default function ChangePasswordPage() {
  return <ChangePasswordView />;
}
