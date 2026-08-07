import type { Metadata } from "next";

import { ProfileView } from "@/components/profile-view";

export const metadata: Metadata = {
  title: "Profil",
  description: "Lihat informasi dan keamanan akun AksesCheck.",
};

export default function ProfilePage() {
  return <ProfileView />;
}
