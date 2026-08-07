import {
  BadgeCheck,
  FileCheck2,
  House,
  ScanSearch,
  ShieldCheck,
} from "lucide-react";
import Link from "next/link";

import { Brand } from "@/components/brand";
import { ThemeToggle } from "@/components/theme-toggle";

const highlights = [
  {
    icon: ScanSearch,
    title: "Temukan masalah lebih cepat",
    description: "Periksa halaman dan lihat bagian yang membutuhkan perhatian.",
  },
  {
    icon: FileCheck2,
    title: "Dapatkan langkah perbaikan",
    description:
      "Setiap temuan memiliki penjelasan yang dapat langsung ditindaklanjuti.",
  },
  {
    icon: ShieldCheck,
    title: "Data akun tetap terlindungi",
    description:
      "Sesi dan perubahan password diproses dengan pengamanan berlapis.",
  },
];

export function AuthPageShell({
  eyebrow,
  title,
  description,
  children,
  footer,
  showHomeLink = true,
}: Readonly<{
  eyebrow: string;
  title: string;
  description: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  showHomeLink?: boolean;
}>) {
  return (
    <main className="relative min-h-dvh overflow-hidden bg-slate-50 text-slate-950 dark:bg-slate-950 dark:text-white">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_8%_8%,rgba(13,148,136,0.15),transparent_28%),radial-gradient(circle_at_92%_90%,rgba(168,85,247,0.10),transparent_30%)] dark:bg-[radial-gradient(circle_at_8%_8%,rgba(13,148,136,0.28),transparent_30%),radial-gradient(circle_at_92%_90%,rgba(168,85,247,0.15),transparent_32%)]" />

      <div className="pointer-events-none absolute inset-0 opacity-[0.035] [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:52px_52px] dark:opacity-[0.05]" />

      <header className="relative z-10 border-b border-slate-200/70 bg-white/70 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/60">
        <div className="mx-auto flex h-16 w-full max-w-7xl items-center justify-between gap-4 px-4 sm:h-[4.5rem] sm:px-6 lg:px-8">
          <Brand />

          <div className="flex shrink-0 items-center gap-2">
            <ThemeToggle />

            {showHomeLink && (
              <Link
                href="/"
                className="inline-flex h-10 items-center gap-2 rounded-xl border border-slate-200 bg-white px-3 text-xs font-bold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-100 sm:px-4 sm:text-sm dark:border-white/10 dark:bg-white/5 dark:text-slate-200 dark:hover:border-white/20 dark:hover:bg-white/10"
              >
                <House className="size-4" aria-hidden="true" />

                <span className="hidden xs:inline">Ke Beranda</span>
                <span className="xs:hidden">Beranda</span>
              </Link>
            )}
          </div>
        </div>
      </header>

      <div className="relative z-10 mx-auto grid min-h-[calc(100dvh-4rem)] w-full max-w-7xl items-center gap-10 px-4 py-8 sm:min-h-[calc(100dvh-4.5rem)] sm:px-6 sm:py-10 lg:grid-cols-[minmax(0,0.95fr)_minmax(400px,0.75fr)] lg:gap-16 lg:px-8 lg:py-12 xl:gap-20">
        <section className="hidden min-w-0 lg:block">
          <div className="max-w-xl">
            <span className="inline-flex items-center gap-2 rounded-full border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-bold text-blue-700 shadow-sm dark:border-blue-400/20 dark:bg-blue-400/10 dark:text-blue-300">
              <BadgeCheck className="size-4 shrink-0" aria-hidden="true" />
              Pemeriksaan aksesibilitas yang mudah dipahami
            </span>

            <h1 className="mt-6 text-4xl font-black leading-[1.08] tracking-[-0.04em] text-slate-950 dark:text-white xl:text-5xl">
              Website yang lebih mudah digunakan dimulai dari pemeriksaan yang
              jelas.
            </h1>

            <p className="mt-5 max-w-lg text-base leading-7 text-slate-600 dark:text-slate-300">
              AksesCheck membantu Anda menemukan hambatan pada website, memahami
              dampaknya, lalu menentukan apa yang harus diperbaiki lebih dulu.
            </p>

            <div className="mt-8 grid gap-3">
              {highlights.map((item) => {
                const Icon = item.icon;

                return (
                  <article
                    key={item.title}
                    className="flex items-start gap-4 rounded-2xl border border-slate-200 bg-white/70 p-4 shadow-sm backdrop-blur transition hover:border-blue-200 hover:bg-white dark:border-white/10 dark:bg-white/[0.035] dark:hover:border-blue-400/20 dark:hover:bg-white/[0.055]"
                  >
                    <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-blue-50 text-blue-600 dark:bg-blue-400/10 dark:text-blue-300">
                      <Icon className="size-5" aria-hidden="true" />
                    </span>

                    <div className="min-w-0">
                      <h2 className="text-sm font-black">{item.title}</h2>

                      <p className="mt-1 text-sm leading-6 text-slate-500 dark:text-slate-400">
                        {item.description}
                      </p>
                    </div>
                  </article>
                );
              })}
            </div>

            <p className="mt-7 max-w-lg text-xs leading-5 text-slate-400 dark:text-slate-500">
              Pemeriksaan otomatis membantu menemukan banyak masalah.
              Pemeriksaan manual tetap diperlukan untuk hasil yang lebih
              lengkap.
            </p>
          </div>
        </section>

        <section className="flex min-w-0 justify-center lg:justify-end">
          <div className="w-full max-w-[31rem]">
            <div className="mb-6 lg:hidden">
              <span className="inline-flex items-center gap-2 rounded-full border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-bold text-blue-700 dark:border-blue-400/20 dark:bg-blue-400/10 dark:text-blue-300">
                <BadgeCheck className="size-4" aria-hidden="true" />
                Pemeriksaan aksesibilitas yang mudah dipahami
              </span>
            </div>

            <div className="rounded-[1.75rem] border border-slate-200 bg-white/95 p-5 text-slate-950 shadow-2xl shadow-slate-950/10 backdrop-blur-xl sm:p-7 lg:p-8 dark:border-white/10 dark:bg-slate-900/95 dark:text-white dark:shadow-black/30">
              <p className="text-xs font-black uppercase tracking-[0.16em] text-blue-600 sm:text-sm dark:text-blue-400">
                {eyebrow}
              </p>

              <h2 className="mt-2.5 text-2xl font-black tracking-tight sm:text-3xl">
                {title}
              </h2>

              <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400">
                {description}
              </p>

              <div className="mt-7">{children}</div>

              {footer && (
                <div className="mt-7 border-t border-slate-200 pt-5 text-center text-sm leading-6 text-slate-600 dark:border-white/10 dark:text-slate-400">
                  {footer}
                </div>
              )}
            </div>

            <p className="mx-auto mt-5 max-w-md text-center text-xs leading-5 text-slate-400 dark:text-slate-500 lg:hidden">
              Pemeriksaan otomatis tidak menggantikan audit aksesibilitas manual
              secara menyeluruh.
            </p>
          </div>
        </section>
      </div>
    </main>
  );
}
