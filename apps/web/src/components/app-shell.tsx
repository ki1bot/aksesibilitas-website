"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  FolderKanban,
  History,
  LayoutDashboard,
  LogOut,
  Menu,
  Moon,
  ScanSearch,
  Sun,
  X,
} from "lucide-react";
import { useTheme } from "next-themes";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Brand } from "@/components/brand";
import { LoadingState } from "@/components/ui-kit";
import { ApiError } from "@/lib/api/client";
import { getCurrentUser, logoutAccount } from "@/lib/api/services";
import { cn } from "@/lib/utils";

const navigation = [
  {
    href: "/dashboard",
    label: "Dashboard",
    icon: LayoutDashboard,
  },
  {
    href: "/dashboard#projects",
    label: "Project",
    icon: FolderKanban,
  },
  {
    href: "/dashboard#history",
    label: "Histori scan",
    icon: History,
  },
];

export function AppShell({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const pathname = usePathname();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { resolvedTheme, setTheme } = useTheme();
  const [menuOpen, setMenuOpen] = useState(false);

  const userQuery = useQuery({
    queryKey: ["current-user"],
    queryFn: getCurrentUser,
    retry: false,
  });

  useEffect(() => {
    if (userQuery.error instanceof ApiError && userQuery.error.status === 401) {
      router.replace("/login");
    }
  }, [router, userQuery.error]);

  async function handleLogout() {
    try {
      await logoutAccount();
      queryClient.clear();
      router.replace("/login");
      router.refresh();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Logout gagal");
    }
  }

  if (userQuery.isPending) {
    return (
      <main className="min-h-screen bg-slate-50 dark:bg-slate-950">
        <LoadingState label="Memeriksa sesi pengguna" />
      </main>
    );
  }

  if (!userQuery.data) {
    return null;
  }

  const user = userQuery.data;

  return (
    <div className="min-h-screen bg-slate-50 text-slate-950 dark:bg-slate-950 dark:text-white">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-72 border-r border-slate-200 bg-white px-5 py-6 dark:border-white/10 dark:bg-slate-950 lg:flex lg:flex-col">
        <Brand href="/dashboard" />

        <nav
          className="mt-10 flex flex-1 flex-col gap-2"
          aria-label="Navigasi utama"
        >
          {navigation.map((item) => {
            const Icon = item.icon;
            const active =
              item.href === "/dashboard"
                ? pathname === "/dashboard"
                : pathname.startsWith(item.href.split("#")[0]) &&
                  item.href !== "/dashboard";

            return (
              <Link
                key={item.label}
                href={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-semibold transition",
                  active
                    ? "bg-blue-600 text-white shadow-lg shadow-blue-600/20"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-950 dark:text-slate-300 dark:hover:bg-white/5 dark:hover:text-white",
                )}
              >
                <Icon className="size-5" aria-hidden="true" />
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 dark:border-white/10 dark:bg-white/[0.03]">
          <div className="flex items-center gap-3">
            <span className="grid size-10 place-items-center rounded-xl bg-blue-100 font-bold text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
              {user.name.slice(0, 1).toUpperCase()}
            </span>

            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-bold">{user.name}</p>
              <p className="truncate text-xs text-slate-500 dark:text-slate-400">
                {user.email}
              </p>
            </div>
          </div>

          <button
            type="button"
            onClick={handleLogout}
            className="mt-4 flex w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm font-semibold text-slate-700 transition hover:border-red-200 hover:bg-red-50 hover:text-red-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-red-400/20 dark:hover:bg-red-400/10 dark:hover:text-red-300"
          >
            <LogOut className="size-4" aria-hidden="true" />
            Keluar
          </button>
        </div>
      </aside>

      <header className="sticky top-0 z-30 border-b border-slate-200 bg-white/90 px-4 py-3 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/90 lg:ml-72 lg:px-8">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-4">
          <div className="flex items-center gap-3 lg:hidden">
            <button
              type="button"
              onClick={() => setMenuOpen(true)}
              className="grid size-10 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-white"
              aria-label="Buka navigasi"
            >
              <Menu className="size-5" aria-hidden="true" />
            </button>

            <Brand compact href="/dashboard" />
          </div>

          <div className="hidden lg:block">
            <p className="text-xs font-bold uppercase tracking-[0.18em] text-blue-600 dark:text-blue-400">
              Accessibility workspace
            </p>
            <p className="text-sm text-slate-500 dark:text-slate-400">
              Audit satu halaman dengan WCAG 2.2 A dan AA
            </p>
          </div>

          <button
            type="button"
            onClick={() =>
              setTheme(resolvedTheme === "dark" ? "light" : "dark")
            }
            className="grid size-10 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 transition hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
            aria-label="Ubah tema"
          >
            {resolvedTheme === "dark" ? (
              <Sun className="size-4.5" aria-hidden="true" />
            ) : (
              <Moon className="size-4.5" aria-hidden="true" />
            )}
          </button>
        </div>
      </header>

      {menuOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <button
            type="button"
            className="absolute inset-0 bg-slate-950/50 backdrop-blur-sm"
            onClick={() => setMenuOpen(false)}
            aria-label="Tutup navigasi"
          />

          <aside className="relative h-full w-[min(22rem,88vw)] bg-white p-5 shadow-2xl dark:bg-slate-950">
            <div className="flex items-center justify-between">
              <Brand href="/dashboard" />

              <button
                type="button"
                onClick={() => setMenuOpen(false)}
                className="grid size-10 place-items-center rounded-xl border border-slate-200 dark:border-white/10"
                aria-label="Tutup navigasi"
              >
                <X className="size-5" aria-hidden="true" />
              </button>
            </div>

            <nav className="mt-10 flex flex-col gap-2">
              {navigation.map((item) => {
                const Icon = item.icon;

                return (
                  <Link
                    key={item.label}
                    href={item.href}
                    onClick={() => setMenuOpen(false)}
                    className="flex items-center gap-3 rounded-2xl px-4 py-3 text-sm font-semibold text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-white/5"
                  >
                    <Icon className="size-5" aria-hidden="true" />
                    {item.label}
                  </Link>
                );
              })}
            </nav>

            <button
              type="button"
              onClick={handleLogout}
              className="mt-8 flex w-full items-center justify-center gap-2 rounded-xl bg-red-50 px-4 py-3 text-sm font-bold text-red-700 dark:bg-red-400/10 dark:text-red-300"
            >
              <LogOut className="size-4" aria-hidden="true" />
              Keluar
            </button>
          </aside>
        </div>
      )}

      <main className="lg:ml-72">
        <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-10">
          {children}
        </div>
      </main>

      <div className="sr-only">
        <ScanSearch />
      </div>
    </div>
  );
}
