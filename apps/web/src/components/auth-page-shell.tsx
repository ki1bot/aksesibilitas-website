import { BadgeCheck, FileCheck2, ScanSearch, ShieldCheck } from "lucide-react";
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
}: Readonly<{
  eyebrow: string;
  title: string;
  description: string;
  children: React.ReactNode;
  footer?: React.ReactNode;
}>) {
  return (
    <main className="relative min-h-screen overflow-hidden bg-slate-950 text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_10%_10%,rgba(13,148,136,0.28),transparent_28%),radial-gradient(circle_at_90%_85%,rgba(217,70,239,0.16),transparent_30%)]" />
      <div className="absolute inset-0 opacity-[0.06] [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:52px_52px]" />

      <div className="relative mx-auto grid min-h-screen max-w-7xl lg:grid-cols-[1.05fr_0.95fr]">
        <section className="hidden flex-col justify-between px-12 py-10 lg:flex xl:px-16 xl:py-12">
          <Brand className="[&_span_span:first-child]:text-white" />

          <div className="max-w-xl py-10">
            <span className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-bold text-blue-200 backdrop-blur">
              <BadgeCheck className="size-4" aria-hidden="true" />
              Pemeriksaan aksesibilitas yang mudah dipahami
            </span>

            <h1 className="mt-7 text-5xl font-black leading-[1.06] tracking-[-0.045em] xl:text-6xl">
              Website yang lebih mudah digunakan dimulai dari pemeriksaan yang
              jelas.
            </h1>

            <p className="mt-6 max-w-lg text-lg leading-8 text-slate-300">
              AksesCheck membantu Anda menemukan hambatan pada website, memahami
              dampaknya, lalu memperbaikinya satu per satu.
            </p>

            <div className="mt-10 grid gap-3">
              {highlights.map((item) => {
                const Icon = item.icon;

                return (
                  <article
                    key={item.title}
                    className="flex gap-4 rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur"
                  >
                    <span className="grid size-11 shrink-0 place-items-center rounded-xl bg-blue-400/10 text-blue-200">
                      <Icon className="size-5" aria-hidden="true" />
                    </span>

                    <div>
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

        <section className="flex items-center justify-center px-4 py-6 sm:px-8 sm:py-10 lg:px-10">
          <div className="w-full max-w-lg">
            <div className="mb-6 flex items-center justify-between lg:hidden">
              <Brand className="[&_span_span:first-child]:text-white" />

              <Link
                href="/"
                className="rounded-xl border border-white/10 bg-white/5 px-3 py-2 text-sm font-semibold text-slate-200 transition hover:bg-white/10"
              >
                Halaman utama
              </Link>
            </div>

            <div className="rounded-[2rem] border border-white/10 bg-white p-6 text-slate-950 shadow-2xl shadow-black/30 sm:p-9 dark:bg-slate-900 dark:text-white">
              <p className="text-sm font-bold uppercase tracking-[0.16em] text-blue-600 dark:text-blue-400">
                {eyebrow}
              </p>

              <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
                {title}
              </h2>

              <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400 sm:text-base">
                {description}
              </p>

              <div className="mt-8">{children}</div>

              {footer && (
                <div className="mt-7 border-t border-slate-200 pt-6 text-center text-sm text-slate-600 dark:border-white/10 dark:text-slate-400">
                  {footer}
                </div>
              )}
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
