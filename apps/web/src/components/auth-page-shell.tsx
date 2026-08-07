import {
  BadgeCheck,
  FileCheck2,
  House,
  ScanSearch,
  ShieldCheck,
} from "lucide-react";
import Link from "next/link";

import { Brand } from "@/components/brand";

const highlights = [
  {
    icon: ScanSearch,
    title: "Temukan masalah lebih cepat",
    description:
      "Periksa satu halaman website dan lihat bagian yang perlu diperbaiki.",
  },
  {
    icon: FileCheck2,
    title: "Dapatkan langkah perbaikan",
    description:
      "Setiap temuan dilengkapi penjelasan yang dapat langsung ditindaklanjuti.",
  },
  {
    icon: ShieldCheck,
    title: "Data akun tetap terlindungi",
    description:
      "Sesi, token, dan perubahan password diproses dengan pengamanan berlapis.",
  },
];

export function AuthPageShell({
  eyebrow,
  title,
  description,
  children,
  footer,
  showHomeLink = false,
}: Readonly<{
  eyebrow: string;
  title: string;
  description: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
  showHomeLink?: boolean;
}>) {
  return (
    <main className="relative min-h-dvh overflow-hidden bg-slate-950 text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_8%_8%,rgba(13,148,136,0.3),transparent_30%),radial-gradient(circle_at_92%_88%,rgba(217,70,239,0.14),transparent_30%)]" />

      <div className="absolute inset-0 opacity-[0.05] [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:52px_52px]" />

      <div className="relative mx-auto grid min-h-dvh w-full max-w-[1440px] lg:grid-cols-[minmax(0,1.05fr)_minmax(440px,0.95fr)]">
        <section className="hidden min-w-0 flex-col justify-between px-10 py-10 lg:flex xl:px-16 xl:py-12">
          <div className="flex items-center justify-between gap-4">
            <Brand className="[&_span_span:first-child]:text-white" />

            {showHomeLink && (
              <Link
                href="/"
                className="inline-flex h-10 shrink-0 items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 text-sm font-bold text-slate-200 backdrop-blur transition hover:border-white/20 hover:bg-white/10 hover:text-white"
              >
                <House className="size-4" aria-hidden="true" />
                Ke Beranda
              </Link>
            )}
          </div>

          <div className="my-auto max-w-2xl py-12">
            <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-bold text-blue-200 backdrop-blur">
              <BadgeCheck className="size-4" aria-hidden="true" />
              Pemeriksaan aksesibilitas yang mudah dipahami
            </span>

            <h1 className="mt-7 max-w-2xl text-4xl font-black leading-[1.08] tracking-[-0.04em] xl:text-5xl 2xl:text-6xl">
              Website yang lebih mudah digunakan dimulai dari pemeriksaan yang
              jelas.
            </h1>

            <p className="mt-6 max-w-xl text-base leading-7 text-slate-300 xl:text-lg xl:leading-8">
              AksesCheck membantu Anda menemukan hambatan pada website, memahami
              dampaknya, lalu memperbaikinya satu per satu.
            </p>

            <div className="mt-9 grid max-w-xl gap-3">
              {highlights.map((item) => {
                const Icon = item.icon;

                return (
                  <article
                    key={item.title}
                    className="flex gap-4 rounded-2xl border border-white/10 bg-white/[0.04] p-4 backdrop-blur transition hover:bg-white/[0.07]"
                  >
                    <span className="grid size-11 shrink-0 place-items-center rounded-xl bg-blue-400/10 text-blue-200">
                      <Icon className="size-5" aria-hidden="true" />
                    </span>

                    <div className="min-w-0">
                      <h2 className="text-sm font-bold">{item.title}</h2>

                      <p className="mt-1 text-sm leading-6 text-slate-400">
                        {item.description}
                      </p>
                    </div>
                  </article>
                );
              })}
            </div>
          </div>

          <p className="max-w-xl text-xs leading-5 text-slate-500">
            Pemeriksaan otomatis membantu menemukan banyak masalah, tetapi
            pemeriksaan manual tetap diperlukan untuk hasil yang lengkap.
          </p>
        </section>

        <section className="flex min-w-0 items-center justify-center px-4 py-5 sm:px-6 sm:py-8 lg:border-l lg:border-white/5 lg:bg-black/10 lg:px-8 xl:px-12">
          <div className="w-full max-w-lg">
            <div className="mb-5 flex min-w-0 items-center justify-between gap-3 lg:hidden">
              <div className="min-w-0">
                <Brand className="[&_span_span:first-child]:text-white" />
              </div>

              {showHomeLink && (
                <Link
                  href="/"
                  className="inline-flex h-10 shrink-0 items-center gap-2 rounded-xl border border-white/10 bg-white/5 px-3 text-xs font-bold text-slate-200 transition hover:bg-white/10 sm:px-4 sm:text-sm"
                >
                  <House className="size-4" aria-hidden="true" />
                  Ke Beranda
                </Link>
              )}
            </div>

            <div className="rounded-[1.75rem] border border-white/10 bg-white p-5 text-slate-950 shadow-2xl shadow-black/30 sm:p-8 lg:p-9 dark:bg-slate-900 dark:text-white">
              <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-600 sm:text-sm dark:text-blue-400">
                {eyebrow}
              </p>

              <h2 className="mt-3 text-2xl font-black tracking-tight sm:text-3xl lg:text-4xl">
                {title}
              </h2>

              <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400 sm:text-base sm:leading-7">
                {description}
              </p>

              <div className="mt-7 sm:mt-8">{children}</div>

              {footer && (
                <div className="mt-7 border-t border-slate-200 pt-6 text-center text-sm leading-6 text-slate-600 dark:border-white/10 dark:text-slate-400">
                  {footer}
                </div>
              )}
            </div>

            <p className="mx-auto mt-5 max-w-md text-center text-xs leading-5 text-slate-500 lg:hidden">
              Pemeriksaan otomatis tidak menggantikan audit aksesibilitas manual
              secara menyeluruh.
            </p>
          </div>
        </section>
      </div>
    </main>
  );
}
