import type { Metadata } from "next";

import { ForgotPasswordForm } from "@/components/password-forms";

export const metadata: Metadata = {
  title: "Lupa Password",
  description: "Minta tautan untuk membuat password baru.",
};

export default function ForgotPasswordPage() {
  return <ForgotPasswordForm />;
}
