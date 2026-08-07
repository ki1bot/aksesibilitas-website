"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  FolderKanban,
  History,
  KeyRound,
  LayoutDashboard,
  LogOut,
  Menu,
  Moon,
  Sun,
  X,
} from "lucide-react";
import { useTheme } from "next-themes";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState, useSyncExternalStore } from "react";
import { toast } from "sonner";

import { Brand } from "@/components/brand";
import { LoadingState } from "@/components/ui-kit";
import { ApiError } from "@/lib/api/client";
import { getCurrentUser, logoutAccount } from "@/lib/api/services";
import { cn } from "@/lib/utils";

const navigation = [
  {
    href: "/dashboard",
    label: "Ringkasan",
    description: "Lihat kondisi workspace",
    icon: LayoutDashboard,
  },
  {
    href: "/dashboard#projects",
    label: "Project",
    description: "Kelola website yang diperiksa",
    icon: FolderKanban,
    hash: "#projects",
  },
  {
    href: "/dashboard#history",
    label: "Riwayat pemeriksaan",
    description: "Buka hasil pemeriksaan sebelumnya",
    icon: History,
    hash: "#history",
  },
  {
    href: "/change-password",
    label: "Ganti password",
    description: "Perbarui keamanan akun",
    icon: KeyRound,
  },
];

function subscribeToHashChange(callback: () => void) {
  window.addEventListener("hashchange", callback);
  window.addEventListener("popstate", callback);

  return () => {
    window.removeEventListener("hashchange", callback);
    window.removeEventListener("popstate", callback);
  };
}

function getHashSnapshot() {
  return window.location.hash;
}

function getServerHashSnapshot() {
  return "";
}

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

  const currentHash = useSyncExternalStore(
    subscribeToHashChange,
    getHashSnapshot,
    getServerHashSnapshot,
  );

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

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    const previousOverflow = document.body.style.overflow;

    document.body.style.overflow = "hidden";

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setMenuOpen(false);
      }
    }

    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [menuOpen]);

  function navigationIsActive(item: (typeof navigation)[number]) {
    if (item.hash === "#projects") {
      return (
        (pathname === "/dashboard" && currentHash === "#projects") ||
        pathname.startsWith("/projects/")
      );
    }

    if (item.hash === "#history") {
      return (
        (pathname === "/dashboard" && currentHash === "#history") ||
        pathname.startsWith("/scans/")
      );
    }

    if (item.href === "/dashboard") {
      return (
        pathname === "/dashboard" &&
        currentHash !== "#projects" &&
        currentHash !== "#history"
      );
    }

    return pathname === item.href;
  }

  const activeNavigation = navigation.find(navigationIsActive) ?? navigation[0];

  async function handleLogout() {
    try {
      await logoutAccount();

      queryClient.clear();

      router.replace("/login");
      router.refresh();
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "Gagal keluar dari akun",
      );
    }
  }

  if (userQuery.isPending) {
    return (
      <main className="grid min-h-dvh place-items-center bg-slate-50 px-4 dark:bg-slate-950">
        <LoadingState label="Memeriksa sesi akun" />
      </main>
    );
  }

  if (!userQuery.data) {
    return null;
  }

  const user = userQuery.data;

  return (
    <div className="min-h-dvh bg-slate-50 text-slate-950 dark:bg-slate-950 dark:text-white">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-64 border-r border-slate-200 bg-white dark:border-white/10 dark:bg-slate-950 lg:flex lg:flex-col xl:w-72">
        <div className="flex h-20 shrink-0 items-center border-b border-slate-100 px-5 dark:border-white/5 xl:px-6">
          <Brand href="/dashboard" />
        </div>

        <nav
          className="flex flex-1 flex-col gap-1.5 overflow-y-auto px-3 py-6 xl:px-4"
          aria-label="Navigasi utama"
        >
          {navigation.map((item) => {
            const Icon = item.icon;
            const active = navigationIsActive(item);

            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "group flex items-start gap-3 rounded-2xl px-3.5 py-3 transition xl:px-4",
                  active
                    ? "bg-blue-600 text-white shadow-lg shadow-blue-600/15"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-950 dark:text-slate-300 dark:hover:bg-white/5 dark:hover:text-white",
                )}
                aria-current={active ? "page" : undefined}
              >
                <span
                  className={cn(
                    "grid size-9 shrink-0 place-items-center rounded-xl transition",
                    active
                      ? "bg-white/15"
                      : "bg-slate-100 text-slate-500 group-hover:bg-white dark:bg-white/5 dark:text-slate-400 dark:group-hover:bg-white/10",
                  )}
                >
                  <Icon className="size-4.5" aria-hidden="true" />
                </span>

                <span className="min-w-0 pt-0.5">
                  <span className="block text-sm font-bold">{item.label}</span>

                  <span
                    className={cn(
                      "mt-0.5 block text-xs leading-5",
                      active
                        ? "text-white/75"
                        : "text-slate-400 dark:text-slate-500",
                    )}
                  >
                    {item.description}
                  </span>
                </span>
              </Link>
            );
          })}
        </nav>

        <div className="shrink-0 border-t border-slate-100 p-4 dark:border-white/5 xl:p-5">
          <div className="rounded-2xl border border-slate-200 bg-slate-50 p-3.5 dark:border-white/10 dark:bg-white/[0.03]">
            <div className="flex min-w-0 items-center gap-3">
              <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-blue-100 text-sm font-black text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
                {user.name.slice(0, 1).toUpperCase()}
              </span>

              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-bold">{user.name}</p>

                <p className="mt-0.5 truncate text-xs text-slate-500 dark:text-slate-400">
                  {user.email}
                </p>
              </div>
            </div>

            <button
              type="button"
              onClick={handleLogout}
              className="mt-3 flex min-h-10 w-full items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-red-200 hover:bg-red-50 hover:text-red-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-red-400/20 dark:hover:bg-red-400/10 dark:hover:text-red-300"
            >
              <LogOut className="size-4" aria-hidden="true" />
              Keluar
            </button>
          </div>
        </div>
      </aside>

      <div className="min-w-0 lg:pl-64 xl:pl-72">
        <header className="sticky top-0 z-30 border-b border-slate-200 bg-white/90 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/90">
          <div className="mx-auto flex h-16 max-w-[1440px] items-center justify-between gap-3 px-4 sm:px-6 lg:px-8 xl:px-10">
            <div className="flex min-w-0 items-center gap-3">
              <button
                type="button"
                onClick={() => setMenuOpen(true)}
                className="grid size-10 shrink-0 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 shadow-sm transition hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10 lg:hidden"
                aria-label="Buka navigasi"
                aria-expanded={menuOpen}
                aria-controls="mobile-navigation"
              >
                <Menu className="size-5" aria-hidden="true" />
              </button>

              <div className="min-w-0">
                <p className="truncate text-sm font-black sm:text-base">
                  {activeNavigation.label}
                </p>

                <p className="hidden truncate text-xs text-slate-500 dark:text-slate-400 sm:block">
                  {activeNavigation.description}
                </p>
              </div>
            </div>

            <div className="flex shrink-0 items-center gap-2">
              <span className="hidden max-w-48 truncate rounded-xl bg-slate-100 px-3 py-2 text-xs font-semibold text-slate-600 dark:bg-white/5 dark:text-slate-300 md:block">
                {user.name}
              </span>

              <button
                type="button"
                onClick={() =>
                  setTheme(resolvedTheme === "dark" ? "light" : "dark")
                }
                className="grid size-10 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 shadow-sm transition hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
                aria-label={
                  resolvedTheme === "dark"
                    ? "Gunakan tema terang"
                    : "Gunakan tema gelap"
                }
              >
                {resolvedTheme === "dark" ? (
                  <Sun className="size-4.5" aria-hidden="true" />
                ) : (
                  <Moon className="size-4.5" aria-hidden="true" />
                )}
              </button>
            </div>
          </div>
        </header>

        <main className="min-w-0">
          <div className="mx-auto w-full max-w-[1440px] px-4 py-6 sm:px-6 sm:py-8 lg:px-8 lg:py-10 xl:px-10">
            {children}
          </div>
        </main>
      </div>

      {menuOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <button
            type="button"
            className="absolute inset-0 bg-slate-950/60 backdrop-blur-sm"
            onClick={() => setMenuOpen(false)}
            aria-label="Tutup navigasi"
          />

          <aside
            id="mobile-navigation"
            className="relative flex h-full w-[min(22rem,calc(100vw-2rem))] flex-col bg-white shadow-2xl dark:bg-slate-950"
          >
            <div className="flex h-18 shrink-0 items-center justify-between border-b border-slate-100 px-4 dark:border-white/5">
              <Brand href="/dashboard" />

              <button
                type="button"
                onClick={() => setMenuOpen(false)}
                className="grid size-10 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 transition hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
                aria-label="Tutup navigasi"
              >
                <X className="size-5" aria-hidden="true" />
              </button>
            </div>

            <nav
              className="flex flex-1 flex-col gap-1.5 overflow-y-auto p-4"
              aria-label="Navigasi mobile"
            >
              {navigation.map((item) => {
                const Icon = item.icon;
                const active = navigationIsActive(item);

                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={() => setMenuOpen(false)}
                    className={cn(
                      "flex items-start gap-3 rounded-2xl px-3.5 py-3 transition",
                      active
                        ? "bg-blue-600 text-white shadow-lg shadow-blue-600/15"
                        : "text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-white/5",
                    )}
                    aria-current={active ? "page" : undefined}
                  >
                    <span
                      className={cn(
                        "grid size-9 shrink-0 place-items-center rounded-xl",
                        active ? "bg-white/15" : "bg-slate-100 dark:bg-white/5",
                      )}
                    >
                      <Icon className="size-4.5" aria-hidden="true" />
                    </span>

                    <span className="min-w-0 pt-0.5">
                      <span className="block text-sm font-bold">
                        {item.label}
                      </span>

                      <span
                        className={cn(
                          "mt-0.5 block text-xs leading-5",
                          active
                            ? "text-white/75"
                            : "text-slate-400 dark:text-slate-500",
                        )}
                      >
                        {item.description}
                      </span>
                    </span>
                  </Link>
                );
              })}
            </nav>

            <div className="shrink-0 border-t border-slate-100 p-4 dark:border-white/5">
              <div className="rounded-2xl bg-slate-50 p-4 dark:bg-white/[0.03]">
                <div className="flex min-w-0 items-center gap-3">
                  <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-blue-100 text-sm font-black text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
                    {user.name.slice(0, 1).toUpperCase()}
                  </span>

                  <div className="min-w-0">
                    <p className="truncate text-sm font-bold">{user.name}</p>

                    <p className="mt-0.5 truncate text-xs text-slate-500 dark:text-slate-400">
                      {user.email}
                    </p>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={handleLogout}
                  className="mt-4 flex min-h-11 w-full items-center justify-center gap-2 rounded-xl bg-red-50 px-4 py-2.5 text-sm font-bold text-red-700 transition hover:bg-red-100 dark:bg-red-400/10 dark:text-red-300 dark:hover:bg-red-400/15"
                >
                  <LogOut className="size-4" aria-hidden="true" />
                  Keluar dari akun
                </button>
              </div>
            </div>
          </aside>
        </div>
      )}
    </div>
  );
}
