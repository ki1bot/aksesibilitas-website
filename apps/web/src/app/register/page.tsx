import type { Metadata } from "next";

import { AuthForm } from "@/components/auth-form";

export const metadata: Metadata = {
  title: "Daftar",
  description: "Buat akun dan workspace AksesCheck ID.",
};

export default function RegisterPage() {
  return <AuthForm mode="register" />;
}
