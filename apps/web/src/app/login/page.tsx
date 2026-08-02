import type { Metadata } from "next";

import { AuthForm } from "@/components/auth-form";

export const metadata: Metadata = {
  title: "Masuk",
  description: "Masuk ke workspace AksesCheck ID.",
};

export default function LoginPage() {
  return <AuthForm mode="login" />;
}
