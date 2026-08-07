"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  FolderKanban,
  History,
  KeyRound,
  LayoutDashboard,
  LogOut,
  Moon,
  Sun,
} from "lucide-react";
import { useTheme } from "next-themes";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useSyncExternalStore } from "react";
import { toast } from "sonner";

import { Brand } from "@/components/brand";
import { ErrorState, LoadingState } from "@/components/ui-kit";
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
    label: "Riwayat",
    description: "Buka hasil pemeriksaan sebelumnya",
    icon: History,
    hash: "#history",
  },
  {
    href: "/change-password",
    label: "Password",
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

  const currentHash = useSyncExternalStore(
    subscribeToHashChange,
    getHashSnapshot,
    getServerHashSnapshot,
  );

  const returnPath =
    pathname === "/dashboard" && currentHash
      ? `${pathname}${currentHash}`
      : pathname;

  const userQuery = useQuery({
    queryKey: ["current-user"],
    queryFn: getCurrentUser,
    retry: false,
    staleTime: 0,
    refetchOnMount: "always",
    refetchOnWindowFocus: true,
    refetchInterval: 60_000,
    refetchIntervalInBackground: false,
  });

  useEffect(() => {
    if (
      !(userQuery.error instanceof ApiError) ||
      userQuery.error.status !== 401
    ) {
      return;
    }

    queryClient.removeQueries({
      predicate: (query) => query.queryKey[0] !== "current-user",
    });

    router.replace(`/login?next=${encodeURIComponent(returnPath)}`);
  }, [queryClient, returnPath, router, userQuery.error]);

  const logoutMutation = useMutation({
    mutationFn: logoutAccount,
    onSuccess: () => {
      queryClient.removeQueries();

      toast.success("Berhasil keluar dari akun");

      router.replace("/login");
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 401) {
        queryClient.removeQueries();

        toast.info("Sesi Anda sudah berakhir");

        router.replace("/login");

        return;
      }

      toast.error(
        error instanceof Error ? error.message : "Gagal keluar dari akun",
      );
    },
  });

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

  function handleLogout() {
    if (logoutMutation.isPending) {
      return;
    }

    logoutMutation.mutate();
  }

  function handleThemeToggle() {
    setTheme(resolvedTheme === "dark" ? "light" : "dark");
  }

  if (userQuery.isPending) {
    return (
      <main className="grid min-h-dvh place-items-center bg-slate-50 px-4 dark:bg-slate-950">
        <LoadingState label="Memeriksa sesi akun" />
      </main>
    );
  }

  if (userQuery.error instanceof ApiError && userQuery.error.status === 401) {
    return (
      <main className="grid min-h-dvh place-items-center bg-slate-50 px-4 dark:bg-slate-950">
        <LoadingState label="Sesi berakhir, mengalihkan ke halaman masuk" />
      </main>
    );
  }

  if (userQuery.error || !userQuery.data) {
    return (
      <main className="min-h-dvh bg-slate-50 px-4 py-10 dark:bg-slate-950 sm:px-6">
        <div className="mx-auto max-w-xl">
          <ErrorState
            message={
              userQuery.error instanceof Error
                ? userQuery.error.message
                : "Sesi akun tidak dapat diperiksa"
            }
          />

          <button
            type="button"
            onClick={() => userQuery.refetch()}
            disabled={userQuery.isFetching}
            className="mt-4 flex min-h-11 w-full items-center justify-center rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-bold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {userQuery.isFetching ? "Memeriksa kembali..." : "Coba lagi"}
          </button>
        </div>
      </main>
    );
  }

  const user = userQuery.data;

  return (
    <div className="min-h-dvh bg-slate-50 text-slate-950 dark:bg-slate-950 dark:text-white">
      <header className="sticky top-0 z-40 border-b border-slate-200 bg-white/95 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/95">
        <div className="mx-auto flex h-16 w-full max-w-[1440px] items-center gap-3 px-4 sm:h-[4.5rem] sm:px-6 lg:px-8 xl:px-10">
          <div className="shrink-0">
            <Brand href="/dashboard" />
          </div>

          <nav
            className="ml-6 hidden min-w-0 flex-1 items-center justify-center gap-1 lg:flex"
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
                    "inline-flex h-10 items-center gap-2 rounded-xl px-3.5 text-sm font-bold transition xl:px-4",
                    active
                      ? "bg-blue-600 text-white shadow-lg shadow-blue-600/15"
                      : "text-slate-600 hover:bg-slate-100 hover:text-slate-950 dark:text-slate-300 dark:hover:bg-white/5 dark:hover:text-white",
                  )}
                  aria-current={active ? "page" : undefined}
                >
                  <Icon className="size-4" aria-hidden="true" />

                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>

          <div className="ml-auto flex shrink-0 items-center gap-2">
            <div className="hidden min-w-0 items-center gap-2.5 rounded-xl border border-slate-200 bg-slate-50 py-1.5 pl-1.5 pr-3 dark:border-white/10 dark:bg-white/[0.04] sm:flex">
              <span className="grid size-8 shrink-0 place-items-center rounded-lg bg-blue-100 text-xs font-black text-blue-700 dark:bg-blue-400/10 dark:text-blue-300">
                {user.name.slice(0, 1).toUpperCase()}
              </span>

              <div className="hidden min-w-0 xl:block">
                <p className="max-w-36 truncate text-xs font-bold">
                  {user.name}
                </p>

                <p className="mt-0.5 max-w-36 truncate text-[11px] text-slate-500 dark:text-slate-400">
                  {user.email}
                </p>
              </div>
            </div>

            <button
              type="button"
              onClick={handleThemeToggle}
              className="grid size-10 place-items-center rounded-xl border border-slate-200 bg-white text-slate-700 shadow-sm transition hover:bg-slate-100 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
              aria-label="Ganti tema"
              title="Ganti tema"
            >
              <Moon className="size-4.5 dark:hidden" aria-hidden="true" />

              <Sun className="hidden size-4.5 dark:block" aria-hidden="true" />
            </button>

            <button
              type="button"
              onClick={handleLogout}
              disabled={logoutMutation.isPending}
              className="inline-flex h-10 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-3 text-sm font-bold text-slate-700 shadow-sm transition hover:border-red-200 hover:bg-red-50 hover:text-red-700 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-red-400/20 dark:hover:bg-red-400/10 dark:hover:text-red-300 sm:px-3.5"
              aria-label={
                logoutMutation.isPending
                  ? "Sedang keluar dari akun"
                  : "Keluar dari akun"
              }
            >
              <LogOut className="size-4" aria-hidden="true" />

              <span className="hidden md:inline">
                {logoutMutation.isPending ? "Keluar..." : "Keluar"}
              </span>
            </button>
          </div>
        </div>

        <div className="border-t border-slate-100 dark:border-white/5 lg:hidden">
          <nav
            className="mx-auto flex w-full max-w-[1440px] gap-2 overflow-x-auto px-4 py-2.5 sm:px-6"
            aria-label="Navigasi mobile"
          >
            {navigation.map((item) => {
              const Icon = item.icon;
              const active = navigationIsActive(item);

              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "inline-flex h-10 shrink-0 items-center gap-2 whitespace-nowrap rounded-xl px-3.5 text-sm font-bold transition",
                    active
                      ? "bg-blue-600 text-white shadow-md shadow-blue-600/15"
                      : "bg-slate-100 text-slate-600 hover:bg-slate-200 hover:text-slate-950 dark:bg-white/5 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white",
                  )}
                  aria-current={active ? "page" : undefined}
                >
                  <Icon className="size-4" aria-hidden="true" />

                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>
      </header>

      <main className="min-w-0">
        <div className="mx-auto w-full max-w-[1440px] px-4 py-6 sm:px-6 sm:py-8 lg:px-8 lg:py-10 xl:px-10">
          {children}
        </div>
      </main>
    </div>
  );
}
