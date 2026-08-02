import {
  ArrowRight,
  Braces,
  CheckCircle2,
  FileText,
  Keyboard,
  Layers3,
  LockKeyhole,
  ScanSearch,
  ServerCog,
  ShieldCheck,
} from "lucide-react";
import Link from "next/link";

import { Brand } from "@/components/brand";

const features = [
  {
    icon: ScanSearch,
    title: "Automated accessibility audit",
    description:
      "Jalankan axe-core melalui Chromium untuk memeriksa aturan WCAG 2.2 Level A dan AA.",
  },
  {
    icon: Braces,
    title: "DOM snippet dan selector",
    description:
      "Temukan elemen yang bermasalah melalui potongan HTML, CSS selector, dan failure summary.",
  },
  {
    icon: Keyboard,
    title: "Pemeriksaan manual",
    description:
      "Lengkapi audit otomatis dengan checklist keyboard, fokus, zoom, warna, dan pembaca layar.",
  },
  {
    icon: FileText,
    title: "Laporan JSON dan PDF",
    description:
      "Ekspor hasil teknis, saran perbaikan, review manual, dan ringkasan dampak.",
  },
  {
    icon: LockKeyhole,
    title: "Proteksi SSRF",
    description:
      "URL private, localhost, loopback, link-local, metadata cloud, dan protokol berbahaya ditolak.",
  },
  {
    icon: ServerCog,
    title: "Worker terisolasi",
    description:
      "API dan scanner worker berjalan sebagai proses terpisah dengan Redis dan Asynq.",
  },
];

const steps = [
  "Masukkan URL website publik",
  "API memvalidasi URL dan alamat IP",
  "Job dikirim ke antrean Redis",
  "Worker menjalankan Chromium dan axe-core",
  "Hasil dinormalisasi dan disimpan",
  "Tinjau masalah dan ekspor laporan",
];

export default function Home() {
  return (
    <main className="min-h-screen overflow-hidden bg-white text-slate-950 dark:bg-slate-950 dark:text-white">
      <header className="relative z-20 border-b border-slate-200/80 bg-white/80 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/80">
        <div className="mx-auto flex h-20 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
          <Brand />

          <nav
            className="hidden items-center gap-8 md:flex"
            aria-label="Navigasi halaman utama"
          >
            <a
              href="#features"
              className="text-sm font-semibold text-slate-600 transition hover:text-blue-600 dark:text-slate-300"
            >
              Fitur
            </a>
            <a
              href="#workflow"
              className="text-sm font-semibold text-slate-600 transition hover:text-blue-600 dark:text-slate-300"
            >
              Cara kerja
            </a>
            <a
              href="#security"
              className="text-sm font-semibold text-slate-600 transition hover:text-blue-600 dark:text-slate-300"
            >
              Keamanan
            </a>
          </nav>

          <div className="flex items-center gap-2">
            <Link
              href="/login"
              className="hidden h-10 items-center rounded-xl px-4 text-sm font-bold text-slate-700 transition hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-white/5 sm:inline-flex"
            >
              Masuk
            </Link>

            <Link
              href="/register"
              className="inline-flex h-10 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700"
            >
              Mulai audit
              <ArrowRight className="size-4" aria-hidden="true" />
            </Link>
          </div>
        </div>
      </header>

      <section className="relative">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_15%_15%,rgba(37,99,235,0.14),transparent_26%),radial-gradient(circle_at_85%_70%,rgba(124,58,237,0.12),transparent_28%)]" />
        <div className="absolute inset-0 opacity-[0.035] [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:44px_44px]" />

        <div className="relative mx-auto grid max-w-7xl gap-14 px-4 py-20 sm:px-6 sm:py-28 lg:grid-cols-[1.08fr_0.92fr] lg:items-center lg:px-8 lg:py-36">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-bold text-blue-700 dark:border-blue-400/20 dark:bg-blue-400/10 dark:text-blue-300">
              <ShieldCheck className="size-3.5" aria-hidden="true" />
              WCAG 2.2 Level A dan AA
            </span>

            <h1 className="mt-7 max-w-4xl text-5xl font-black leading-[1.02] tracking-[-0.055em] sm:text-6xl lg:text-7xl">
              Audit aksesibilitas yang{" "}
              <span className="text-blue-600 dark:text-blue-400">
                jelas dan dapat ditindaklanjuti.
              </span>
            </h1>

            <p className="mt-7 max-w-2xl text-lg leading-8 text-slate-600 dark:text-slate-300">
              Periksa kontras, alt text, label form, struktur heading, semantic
              HTML, ARIA, dan masalah aksesibilitas lainnya melalui scanner
              berbasis Go, Chromium, dan axe-core.
            </p>

            <div className="mt-9 flex flex-col gap-3 sm:flex-row">
              <Link
                href="/register"
                className="inline-flex h-13 items-center justify-center gap-2 rounded-2xl bg-blue-600 px-6 text-sm font-bold text-white shadow-xl shadow-blue-600/20 transition hover:-translate-y-0.5 hover:bg-blue-700"
              >
                Buat akun gratis
                <ArrowRight className="size-4" aria-hidden="true" />
              </Link>

              <a
                href="#workflow"
                className="inline-flex h-13 items-center justify-center rounded-2xl border border-slate-200 bg-white px-6 text-sm font-bold text-slate-800 shadow-sm transition hover:-translate-y-0.5 hover:bg-slate-50 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
              >
                Lihat cara kerja
              </a>
            </div>

            <div className="mt-9 flex flex-wrap gap-x-6 gap-y-3 text-sm font-semibold text-slate-600 dark:text-slate-300">
              {[
                "Single-page scan",
                "Histori audit",
                "Manual review",
                "Export laporan",
              ].map((item) => (
                <span key={item} className="inline-flex items-center gap-2">
                  <CheckCircle2
                    className="size-4 text-emerald-600 dark:text-emerald-400"
                    aria-hidden="true"
                  />
                  {item}
                </span>
              ))}
            </div>
          </div>

          <div className="relative">
            <div className="absolute -inset-8 rounded-full bg-blue-600/15 blur-3xl" />

            <div className="relative rounded-[2rem] border border-slate-200 bg-slate-950 p-5 text-white shadow-2xl shadow-blue-950/20 sm:p-7 dark:border-white/10">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-xs font-bold uppercase tracking-[0.18em] text-blue-300">
                    Audit terbaru
                  </p>
                  <p className="mt-1 font-bold">https://example.com</p>
                </div>

                <span className="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs font-bold text-emerald-300">
                  Selesai
                </span>
              </div>

              <div className="mt-8 grid grid-cols-[auto_1fr] items-center gap-7">
                <div
                  className="grid size-32 place-items-center rounded-full"
                  style={{
                    background:
                      "conic-gradient(#2563eb 86%, rgba(255,255,255,0.08) 86%)",
                  }}
                >
                  <div className="grid size-24 place-items-center rounded-full bg-slate-950">
                    <div className="text-center">
                      <strong className="block text-4xl font-black">86</strong>
                      <span className="text-xs text-slate-400">dari 100</span>
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  {[
                    ["Kritis", 0, "bg-red-500"],
                    ["Serius", 2, "bg-orange-500"],
                    ["Sedang", 4, "bg-amber-400"],
                    ["Ringan", 7, "bg-blue-500"],
                  ].map(([label, total, color]) => (
                    <div
                      key={String(label)}
                      className="flex items-center justify-between rounded-xl border border-white/10 bg-white/5 px-4 py-3"
                    >
                      <span className="inline-flex items-center gap-2 text-sm text-slate-300">
                        <span
                          className={`size-2 rounded-full ${String(color)}`}
                        />
                        {label}
                      </span>
                      <strong>{total}</strong>
                    </div>
                  ))}
                </div>
              </div>

              <div className="mt-7 rounded-2xl border border-white/10 bg-white/5 p-4">
                <p className="text-xs font-bold uppercase tracking-[0.14em] text-slate-400">
                  Temuan utama
                </p>
                <p className="mt-2 font-bold">Form elements must have labels</p>
                <code className="mt-3 block overflow-hidden text-ellipsis whitespace-nowrap rounded-lg bg-black/30 px-3 py-2 text-xs text-blue-200">
                  {'<input type="email" name="email">'}
                </code>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        id="features"
        className="border-y border-slate-200 bg-slate-50 py-20 dark:border-white/10 dark:bg-white/[0.02] sm:py-28"
      >
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="max-w-2xl">
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-blue-600 dark:text-blue-400">
              Fitur MVP
            </p>
            <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
              Bukan sekadar angka skor.
            </h2>
            <p className="mt-4 text-base leading-7 text-slate-600 dark:text-slate-400">
              Setiap temuan menyediakan konteks teknis agar developer dapat
              memahami masalah dan memperbaikinya.
            </p>
          </div>

          <div className="mt-12 grid gap-5 md:grid-cols-2 xl:grid-cols-3">
            {features.map((feature) => {
              const Icon = feature.icon;

              return (
                <article
                  key={feature.title}
                  className="rounded-[1.75rem] border border-slate-200 bg-white p-6 shadow-sm transition hover:-translate-y-1 hover:border-blue-300 hover:shadow-xl hover:shadow-blue-600/5 dark:border-white/10 dark:bg-white/[0.03]"
                >
                  <span className="grid size-12 place-items-center rounded-2xl bg-blue-50 text-blue-600 dark:bg-blue-400/10 dark:text-blue-300">
                    <Icon className="size-5" aria-hidden="true" />
                  </span>

                  <h3 className="mt-6 text-lg font-black">{feature.title}</h3>

                  <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-400">
                    {feature.description}
                  </p>
                </article>
              );
            })}
          </div>
        </div>
      </section>

      <section id="workflow" className="py-20 sm:py-28">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="grid gap-14 lg:grid-cols-[0.85fr_1.15fr] lg:items-start">
            <div className="lg:sticky lg:top-28">
              <span className="grid size-14 place-items-center rounded-2xl bg-blue-600 text-white shadow-xl shadow-blue-600/20">
                <Layers3 className="size-6" aria-hidden="true" />
              </span>

              <p className="mt-7 text-sm font-bold uppercase tracking-[0.18em] text-blue-600 dark:text-blue-400">
                Alur sistem
              </p>

              <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
                API cepat, worker terpisah.
              </h2>

              <p className="mt-5 text-base leading-7 text-slate-600 dark:text-slate-400">
                Chromium tidak dijalankan di dalam HTTP handler. API hanya
                membuat data scan dan mengirim pekerjaan ke Redis agar proses
                tetap stabil, dapat dibatalkan, dan dapat diulang.
              </p>
            </div>

            <ol className="space-y-4">
              {steps.map((step, index) => (
                <li
                  key={step}
                  className="flex gap-5 rounded-3xl border border-slate-200 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.03] sm:p-6"
                >
                  <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-slate-950 text-sm font-black text-white dark:bg-blue-600">
                    {index + 1}
                  </span>

                  <div>
                    <h3 className="font-black">{step}</h3>
                    <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-400">
                      {index === 0 &&
                        "Pengguna memilih project lalu memasukkan satu URL website."}
                      {index === 1 &&
                        "Scheme, hostname, DNS, alamat private, loopback, dan metadata cloud diperiksa."}
                      {index === 2 &&
                        "Scan berstatus queued dan task accessibility:scan dikirim ke queue scanner."}
                      {index === 3 &&
                        "Worker membuka profile Chromium sementara, menyuntikkan axe-core, dan menjalankan audit."}
                      {index === 4 &&
                        "Violation, node, selector, dampak, dan metadata halaman disimpan ke PostgreSQL."}
                      {index === 5 &&
                        "Frontend berhenti melakukan polling lalu menampilkan hasil otomatis dan checklist manual."}
                    </p>
                  </div>
                </li>
              ))}
            </ol>
          </div>
        </div>
      </section>

      <section id="security" className="bg-slate-950 py-20 text-white sm:py-28">
        <div className="mx-auto grid max-w-7xl gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:items-center lg:px-8">
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-blue-300">
              Scanner security
            </p>

            <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
              URL scanner tidak boleh menjadi pintu menuju jaringan internal.
            </h2>

            <p className="mt-5 text-base leading-7 text-slate-300">
              Setiap URL diperiksa sebelum masuk antrean dan diperiksa kembali
              saat Chromium meminta resource atau mengikuti redirect.
            </p>

            <Link
              href="/register"
              className="mt-8 inline-flex h-12 items-center gap-2 rounded-xl bg-blue-600 px-5 text-sm font-bold shadow-lg shadow-blue-600/20"
            >
              Mulai pemeriksaan
              <ArrowRight className="size-4" aria-hidden="true" />
            </Link>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            {[
              "Blokir localhost dan loopback",
              "Blokir private dan link-local IP",
              "Validasi seluruh hasil DNS",
              "Validasi ulang setiap redirect",
              "Batasi request dan ukuran data",
              "Blokir protokol non-HTTP",
              "Blokir proses download",
              "Gunakan profile sementara",
            ].map((item) => (
              <div
                key={item}
                className="flex items-start gap-3 rounded-2xl border border-white/10 bg-white/5 p-4"
              >
                <ShieldCheck
                  className="mt-0.5 size-5 shrink-0 text-blue-300"
                  aria-hidden="true"
                />
                <span className="text-sm font-semibold leading-6 text-slate-200">
                  {item}
                </span>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="py-20 sm:py-28">
        <div className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8">
          <div className="rounded-[2.25rem] bg-blue-600 px-6 py-12 text-center text-white shadow-2xl shadow-blue-600/20 sm:px-12">
            <h2 className="text-3xl font-black tracking-tight sm:text-4xl">
              Temukan masalah sebelum pengguna menemukannya.
            </h2>

            <p className="mx-auto mt-4 max-w-2xl text-base leading-7 text-blue-100">
              Mulai dari satu halaman. Simpan histori. Tinjau masalah teknis.
              Lengkapi pemeriksaan manual. Ekspor laporan yang dapat
              ditindaklanjuti.
            </p>

            <Link
              href="/register"
              className="mt-8 inline-flex h-12 items-center gap-2 rounded-xl bg-white px-6 text-sm font-black text-blue-700 shadow-lg"
            >
              Buat workspace
              <ArrowRight className="size-4" aria-hidden="true" />
            </Link>
          </div>
        </div>
      </section>

      <footer className="border-t border-slate-200 py-10 dark:border-white/10">
        <div className="mx-auto flex max-w-7xl flex-col gap-5 px-4 sm:px-6 md:flex-row md:items-center md:justify-between lg:px-8">
          <Brand />

          <p className="max-w-xl text-sm leading-6 text-slate-500 dark:text-slate-400">
            Hasil otomatis hanya mencakup aturan yang dapat diperiksa mesin.
            Pemeriksaan manual tetap diperlukan untuk menentukan kesesuaian
            WCAG.
          </p>
        </div>
      </footer>
    </main>
  );
}
