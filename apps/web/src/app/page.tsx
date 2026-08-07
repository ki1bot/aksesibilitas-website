import {
  ArrowRight,
  CheckCircle2,
  FileSearch,
  Gauge,
  Keyboard,
  ListChecks,
  LockKeyhole,
  ScanSearch,
  ShieldCheck,
  Sparkles,
  Wrench,
} from "lucide-react";
import Link from "next/link";

import { Brand } from "@/components/brand";

const benefits = [
  {
    icon: ScanSearch,
    title: "Temukan masalah penting",
    description:
      "Periksa warna, gambar, formulir, heading, tombol, dan struktur halaman dalam satu proses.",
  },
  {
    icon: FileSearch,
    title: "Lihat bagian yang bermasalah",
    description:
      "Setiap temuan menunjukkan elemen yang perlu diperiksa agar Anda tidak mencari secara manual.",
  },
  {
    icon: Wrench,
    title: "Pahami cara memperbaikinya",
    description:
      "Dapatkan penjelasan singkat, dampak masalah, dan petunjuk perbaikan yang dapat langsung dikerjakan.",
  },
  {
    icon: Keyboard,
    title: "Lengkapi dengan pemeriksaan manual",
    description:
      "Gunakan daftar pemeriksaan untuk keyboard, fokus, zoom, warna, dan pembaca layar.",
  },
  {
    icon: ListChecks,
    title: "Simpan riwayat pemeriksaan",
    description:
      "Kelompokkan website ke dalam project dan buka kembali hasil pemeriksaan kapan saja.",
  },
  {
    icon: LockKeyhole,
    title: "Periksa URL dengan lebih aman",
    description:
      "Sistem menolak alamat lokal, jaringan pribadi, dan tujuan yang berisiko sebelum pemeriksaan dimulai.",
  },
];

const steps = [
  {
    number: "01",
    title: "Tambahkan website",
    description:
      "Buat project, lalu masukkan alamat halaman publik yang ingin Anda periksa.",
  },
  {
    number: "02",
    title: "Tunggu proses selesai",
    description:
      "AksesCheck membuka halaman, menjalankan pemeriksaan, lalu menyimpan hasilnya.",
  },
  {
    number: "03",
    title: "Perbaiki satu per satu",
    description:
      "Mulai dari masalah dengan dampak paling besar, tambahkan catatan, lalu unduh laporan.",
  },
];

export default function Home() {
  return (
    <main className="min-h-dvh overflow-hidden bg-white text-slate-950 dark:bg-slate-950 dark:text-white">
      <header className="sticky top-0 z-40 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/90">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-3 px-4 sm:h-[4.5rem] sm:px-6 lg:px-8">
          <div className="min-w-0 shrink">
            <Brand />
          </div>

          <nav
            className="hidden items-center gap-7 md:flex"
            aria-label="Navigasi halaman utama"
          >
            <a
              href="#manfaat"
              className="text-sm font-semibold text-slate-600 transition hover:text-blue-600 dark:text-slate-300 dark:hover:text-blue-400"
            >
              Manfaat
            </a>

            <a
              href="#cara-kerja"
              className="text-sm font-semibold text-slate-600 transition hover:text-blue-600 dark:text-slate-300 dark:hover:text-blue-400"
            >
              Cara kerja
            </a>

            <a
              href="#keamanan"
              className="text-sm font-semibold text-slate-600 transition hover:text-blue-600 dark:text-slate-300 dark:hover:text-blue-400"
            >
              Keamanan
            </a>
          </nav>

          <div className="flex shrink-0 items-center gap-2">
            <Link
              href="/login"
              className="hidden h-10 items-center rounded-xl px-4 text-sm font-bold text-slate-700 transition hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-white/5 sm:inline-flex"
            >
              Masuk
            </Link>

            <Link
              href="/register"
              className="inline-flex h-10 items-center gap-2 rounded-xl bg-blue-600 px-3.5 text-sm font-bold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700 sm:px-4"
            >
              <span className="sm:hidden">Daftar</span>
              <span className="hidden sm:inline">Coba sekarang</span>
              <ArrowRight className="size-4" aria-hidden="true" />
            </Link>
          </div>
        </div>
      </header>

      <section className="relative">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_10%_15%,rgba(13,148,136,0.18),transparent_28%),radial-gradient(circle_at_90%_80%,rgba(217,70,239,0.09),transparent_28%)]" />

        <div className="absolute inset-0 opacity-[0.03] [background-image:linear-gradient(to_right,currentColor_1px,transparent_1px),linear-gradient(to_bottom,currentColor_1px,transparent_1px)] [background-size:48px_48px]" />

        <div className="relative mx-auto grid max-w-7xl gap-12 px-4 py-14 sm:px-6 sm:py-20 lg:grid-cols-[minmax(0,1.04fr)_minmax(360px,0.96fr)] lg:items-center lg:gap-14 lg:px-8 lg:py-28">
          <div className="min-w-0">
            <span className="inline-flex max-w-full items-center gap-2 rounded-full border border-blue-200 bg-blue-50 px-3 py-1.5 text-xs font-bold text-blue-700 dark:border-blue-400/20 dark:bg-blue-400/10 dark:text-blue-300">
              <ShieldCheck className="size-4 shrink-0" aria-hidden="true" />
              <span>Mengacu pada WCAG 2.2 Level A dan AA</span>
            </span>

            <h1 className="mt-6 max-w-4xl text-[clamp(2.5rem,8vw,4.75rem)] font-black leading-[1.03] tracking-[-0.05em]">
              Cari masalah aksesibilitas{" "}
              <span className="text-blue-600 dark:text-blue-400">
                sebelum pengguna menemukannya.
              </span>
            </h1>

            <p className="mt-6 max-w-2xl text-base leading-7 text-slate-600 sm:text-lg sm:leading-8 dark:text-slate-300">
              AksesCheck membantu Anda memeriksa halaman website, memahami
              masalahnya, dan menentukan bagian mana yang harus diperbaiki lebih
              dulu.
            </p>

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <Link
                href="/register"
                className="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 text-sm font-bold text-white shadow-xl shadow-blue-600/20 transition hover:-translate-y-0.5 hover:bg-blue-700 sm:px-6"
              >
                Buat akun gratis
                <ArrowRight className="size-4" aria-hidden="true" />
              </Link>

              <a
                href="#cara-kerja"
                className="inline-flex min-h-12 items-center justify-center rounded-xl border border-slate-200 bg-white px-5 py-3 text-sm font-bold text-slate-800 shadow-sm transition hover:-translate-y-0.5 hover:bg-slate-50 sm:px-6 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:bg-white/10"
              >
                Lihat cara kerjanya
              </a>
            </div>

            <div className="mt-8 grid gap-3 text-sm font-semibold text-slate-600 sm:grid-cols-2 dark:text-slate-300">
              {[
                "Tidak perlu memasang ekstensi",
                "Hasil tersimpan di setiap project",
                "Ada daftar pemeriksaan manual",
                "Laporan dapat diunduh",
              ].map((item) => (
                <span
                  key={item}
                  className="inline-flex min-w-0 items-start gap-2"
                >
                  <CheckCircle2
                    className="mt-0.5 size-4 shrink-0 text-emerald-600 dark:text-emerald-400"
                    aria-hidden="true"
                  />

                  <span>{item}</span>
                </span>
              ))}
            </div>
          </div>

          <div className="relative mx-auto w-full max-w-xl lg:max-w-none">
            <div className="absolute -inset-6 rounded-full bg-blue-600/15 blur-3xl sm:-inset-8" />

            <div className="relative overflow-hidden rounded-[1.75rem] border border-slate-200 bg-slate-950 p-4 text-white shadow-2xl shadow-blue-950/20 sm:p-6 lg:p-7 dark:border-white/10">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-blue-300 sm:text-xs">
                    Hasil pemeriksaan
                  </p>

                  <p className="mt-1 truncate text-sm font-bold sm:text-base">
                    https://contohwebsite.id
                  </p>
                </div>

                <span className="shrink-0 rounded-full border border-emerald-400/20 bg-emerald-400/10 px-2.5 py-1 text-[11px] font-bold text-emerald-300 sm:px-3 sm:text-xs">
                  Selesai
                </span>
              </div>

              <div className="mt-7 grid gap-6 sm:grid-cols-[auto_minmax(0,1fr)] sm:items-center">
                <div
                  className="mx-auto grid size-28 place-items-center rounded-full sm:mx-0 sm:size-32"
                  style={{
                    background:
                      "conic-gradient(var(--score-color) 86%, rgba(255,253,250,0.08) 86%)",
                  }}
                >
                  <div className="grid size-20 place-items-center rounded-full bg-slate-950 sm:size-24">
                    <div className="text-center">
                      <strong className="block text-3xl font-black sm:text-4xl">
                        86
                      </strong>

                      <span className="text-[11px] text-slate-400 sm:text-xs">
                        dari 100
                      </span>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-2.5 sm:gap-3">
                  {[
                    ["Kritis", 0, "bg-red-500"],
                    ["Serius", 2, "bg-orange-500"],
                    ["Sedang", 4, "bg-amber-400"],
                    ["Ringan", 7, "bg-blue-500"],
                  ].map(([label, total, color]) => (
                    <div
                      key={String(label)}
                      className="min-w-0 rounded-xl border border-white/10 bg-white/5 px-3 py-3 sm:px-4"
                    >
                      <span className="inline-flex max-w-full items-center gap-2 text-[11px] text-slate-400 sm:text-xs">
                        <span
                          className={`size-2 shrink-0 rounded-full ${String(color)}`}
                        />

                        <span className="truncate">{label}</span>
                      </span>

                      <strong className="mt-2 block text-lg sm:text-xl">
                        {total}
                      </strong>
                    </div>
                  ))}
                </div>
              </div>

              <div className="mt-5 rounded-2xl border border-white/10 bg-white/5 p-4 sm:mt-6">
                <div className="flex items-center gap-2 text-[11px] font-bold uppercase tracking-[0.14em] text-slate-400 sm:text-xs">
                  <Sparkles
                    className="size-4 shrink-0 text-blue-300"
                    aria-hidden="true"
                  />
                  Temuan yang perlu diperbaiki
                </div>

                <p className="mt-3 text-sm font-bold sm:text-base">
                  Kolom email belum memiliki label yang jelas
                </p>

                <p className="mt-2 text-sm leading-6 text-slate-400">
                  Pengguna pembaca layar dapat kesulitan memahami informasi yang
                  harus diisi.
                </p>

                <code className="mt-4 block overflow-hidden text-ellipsis whitespace-nowrap rounded-lg bg-black/30 px-3 py-2 text-xs text-blue-200">
                  {'<input type="email" name="email">'}
                </code>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section
        id="manfaat"
        className="border-y border-slate-200 bg-slate-50 py-16 sm:py-20 lg:py-24 dark:border-white/10 dark:bg-white/[0.02]"
      >
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="max-w-2xl">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-600 sm:text-sm dark:text-blue-400">
              Yang Anda dapatkan
            </p>

            <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
              Hasil yang mudah dibaca, bukan sekadar angka.
            </h2>

            <p className="mt-4 text-base leading-7 text-slate-600 dark:text-slate-400">
              Fokus utamanya adalah membantu Anda memahami masalah dan mengambil
              tindakan yang tepat.
            </p>
          </div>

          <div className="mt-10 grid gap-4 sm:mt-12 md:grid-cols-2 xl:grid-cols-3">
            {benefits.map((benefit) => {
              const Icon = benefit.icon;

              return (
                <article
                  key={benefit.title}
                  className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm transition hover:-translate-y-1 hover:border-blue-300 hover:shadow-xl hover:shadow-blue-600/5 sm:rounded-3xl sm:p-6 dark:border-white/10 dark:bg-white/[0.03]"
                >
                  <span className="grid size-11 place-items-center rounded-xl bg-blue-50 text-blue-600 sm:size-12 sm:rounded-2xl dark:bg-blue-400/10 dark:text-blue-300">
                    <Icon className="size-5" aria-hidden="true" />
                  </span>

                  <h3 className="mt-5 text-lg font-black sm:mt-6">
                    {benefit.title}
                  </h3>

                  <p className="mt-2.5 text-sm leading-6 text-slate-600 sm:mt-3 dark:text-slate-400">
                    {benefit.description}
                  </p>
                </article>
              );
            })}
          </div>
        </div>
      </section>

      <section id="cara-kerja" className="py-16 sm:py-20 lg:py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="grid gap-10 lg:grid-cols-[minmax(0,0.78fr)_minmax(0,1.22fr)] lg:items-start lg:gap-14">
            <div className="lg:sticky lg:top-24">
              <span className="grid size-12 place-items-center rounded-xl bg-blue-600 text-white shadow-xl shadow-blue-600/20 sm:size-14 sm:rounded-2xl">
                <Gauge className="size-6" aria-hidden="true" />
              </span>

              <p className="mt-6 text-xs font-bold uppercase tracking-[0.16em] text-blue-600 sm:mt-7 sm:text-sm dark:text-blue-400">
                Cara kerja
              </p>

              <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
                Tiga langkah dari URL sampai daftar perbaikan.
              </h2>

              <p className="mt-4 text-base leading-7 text-slate-600 sm:mt-5 dark:text-slate-400">
                Anda tidak perlu memahami cara kerja scanner untuk mulai
                menggunakannya. Cukup tambahkan halaman dan baca hasilnya.
              </p>
            </div>

            <ol className="space-y-3 sm:space-y-4">
              {steps.map((step) => (
                <li
                  key={step.number}
                  className="flex gap-4 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:gap-5 sm:rounded-3xl sm:p-6 dark:border-white/10 dark:bg-white/[0.03]"
                >
                  <span className="grid size-11 shrink-0 place-items-center rounded-xl bg-slate-950 text-xs font-black text-white sm:size-12 sm:rounded-2xl sm:text-sm dark:bg-blue-600">
                    {step.number}
                  </span>

                  <div className="min-w-0">
                    <h3 className="text-base font-black sm:text-lg">
                      {step.title}
                    </h3>

                    <p className="mt-1.5 text-sm leading-6 text-slate-600 sm:mt-2 dark:text-slate-400">
                      {step.description}
                    </p>
                  </div>
                </li>
              ))}
            </ol>
          </div>
        </div>
      </section>

      <section
        id="keamanan"
        className="bg-slate-950 py-16 text-white sm:py-20 lg:py-24"
      >
        <div className="mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,0.95fr)] lg:items-center lg:gap-14 lg:px-8">
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-blue-300 sm:text-sm">
              Keamanan tetap diperhatikan
            </p>

            <h2 className="mt-3 max-w-2xl text-3xl font-black tracking-tight sm:text-4xl">
              Sistem tidak langsung membuka setiap URL yang diberikan.
            </h2>

            <p className="mt-4 max-w-2xl text-base leading-7 text-slate-300 sm:mt-5">
              Sebelum pemeriksaan dimulai, alamat website divalidasi untuk
              mengurangi risiko akses ke jaringan lokal, layanan internal, dan
              tujuan berbahaya.
            </p>
          </div>

          <div className="grid gap-3 sm:grid-cols-2 sm:gap-4">
            {[
              [
                "Alamat lokal ditolak",
                "Localhost dan loopback tidak dapat dipindai.",
              ],
              [
                "Jaringan pribadi ditolak",
                "Alamat IP private dan link-local tidak diteruskan.",
              ],
              [
                "Sesi akun diamankan",
                "Cookie sesi dan token keamanan dipisahkan.",
              ],
              [
                "Password dapat dipulihkan",
                "Tautan reset hanya berlaku sekali dan memiliki batas waktu.",
              ],
            ].map(([title, description]) => (
              <article
                key={title}
                className="rounded-2xl border border-white/10 bg-white/5 p-4 sm:p-5"
              >
                <ShieldCheck
                  className="size-5 text-blue-300"
                  aria-hidden="true"
                />

                <h3 className="mt-4 font-black">{title}</h3>

                <p className="mt-2 text-sm leading-6 text-slate-400">
                  {description}
                </p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="py-16 sm:py-20 lg:py-24">
        <div className="mx-auto max-w-5xl px-4 sm:px-6">
          <div className="relative overflow-hidden rounded-[1.75rem] bg-blue-600 px-5 py-10 text-center text-white shadow-2xl shadow-blue-600/20 sm:rounded-[2.25rem] sm:px-12 sm:py-14 lg:py-16">
            <div className="absolute -left-20 -top-20 size-64 rounded-full bg-white/10 blur-2xl" />
            <div className="absolute -bottom-24 -right-16 size-72 rounded-full bg-violet-500/20 blur-3xl" />

            <div className="relative">
              <h2 className="text-3xl font-black tracking-tight sm:text-4xl">
                Mulai dari satu halaman website.
              </h2>

              <p className="mx-auto mt-4 max-w-2xl text-sm leading-6 text-white/80 sm:text-base sm:leading-7">
                Buat akun, tambahkan website, lalu gunakan hasil pemeriksaan
                sebagai daftar pekerjaan yang jelas.
              </p>

              <Link
                href="/register"
                className="mt-7 inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-white px-5 py-3 text-sm font-black text-blue-700 shadow-lg transition hover:-translate-y-0.5 hover:bg-blue-50 sm:mt-8 sm:px-6"
              >
                Buat akun gratis
                <ArrowRight className="size-4" aria-hidden="true" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      <footer className="border-t border-slate-200 py-7 dark:border-white/10">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 text-sm text-slate-500 sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8">
          <div className="shrink-0">
            <Brand />
          </div>

          <p className="max-w-2xl leading-6 sm:text-right">
            <p className="max-w-full text-xs font-medium leading-6 sm:text-sm sm:leading-7">
              © {new Date().getFullYear()} All rights reserved.
            </p>
            AksesCheck membantu proses pemeriksaan, bukan menggantikan audit
            manual secara menyeluruh.
          </p>
        </div>
      </footer>
    </main>
  );
}
