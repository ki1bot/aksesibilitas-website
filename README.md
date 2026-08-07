# AksesCheck ID

AksesCheck ID adalah aplikasi pemeriksaan aksesibilitas website yang menggunakan Next.js, TypeScript, Go, PostgreSQL, Redis, Chromium, chromedp, dan axe-core.

Aplikasi menjalankan pemeriksaan otomatis terhadap satu halaman website berdasarkan aturan WCAG 2.2 Level A dan AA yang dapat diperiksa oleh axe-core. Hasil scan menampilkan tingkat dampak, aturan yang dilanggar, DOM snippet, selector, saran perbaikan, pemeriksaan manual, histori scan, serta laporan JSON dan PDF.

## Batas hasil pemeriksaan

Hasil otomatis tidak membuktikan bahwa sebuah website telah memenuhi seluruh ketentuan WCAG.

Pemeriksaan otomatis hanya dapat menemukan masalah yang dapat dikenali oleh aturan mesin. Pengujian manual menggunakan keyboard, pembaca layar, pembesaran halaman, pemeriksaan urutan fokus, dan evaluasi pengalaman pengguna tetap diperlukan.

Skor otomatis AksesCheck ID merupakan indikator internal berdasarkan jumlah dan tingkat dampak violation. Skor tersebut bukan sertifikasi kepatuhan WCAG.

## Teknologi

### Frontend

- Next.js 16
- React 19
- TypeScript
- Tailwind CSS 4
- TanStack Query
- React Hook Form
- Zod
- Recharts
- Lucide React
- Sonner
- openapi-fetch
- openapi-typescript

### Backend

- Go
- chi
- pgx
- sqlc
- Goose
- PostgreSQL
- Redis
- Asynq

### Scanner

- Chromium atau Google Chrome
- chromedp
- axe-core
- Network request interception
- SSRF validation
- Resource limits
- Redirect limits

## Arsitektur

```text
Browser
   |
   v
Next.js Web
   |
   | HTTP + session cookie + CSRF
   v
Go API
   |
   +--------------------+
   |                    |
   v                    v
PostgreSQL          Redis / Asynq
                         |
                         v
                    Go Worker
                         |
                         v
                 Chromium + axe-core
```

API dan worker dijalankan sebagai proses terpisah.

API menangani autentikasi, project, scan, review, dan laporan. API tidak menjalankan Chromium secara langsung.

Worker mengambil pekerjaan dari antrean Redis, membuka halaman menggunakan Chromium, menjalankan axe-core, lalu menyimpan hasilnya ke PostgreSQL.

## Fitur utama

- Pendaftaran dan login pengguna
- Session cookie HTTP-only
- Perlindungan CSRF
- Pembuatan dan pengelolaan project
- Pemindaian satu halaman website
- Antrean pemindaian menggunakan Redis dan Asynq
- Status scan queued, running, completed, failed, dan cancelled
- Pembatalan dan pengulangan scan
- Pemeriksaan WCAG 2.2 Level A dan AA
- Pengelompokan critical, serious, moderate, dan minor
- DOM snippet dan selector
- Failure summary dan referensi aturan
- Pemeriksaan manual
- Histori scan
- Laporan JSON
- Laporan PDF
- Validasi SSRF
- Rate limiting
- OpenAPI sebagai kontrak API dan TypeScript

## Persyaratan

Pastikan perangkat sudah memiliki:

- Go sesuai versi pada `go.mod`
- Node.js 22 atau lebih baru
- pnpm 11.17.0
- PostgreSQL
- Redis atau server yang kompatibel dengan Redis
- Google Chrome atau Chromium
- sqlc
- Goose

Project ini dapat dijalankan tanpa Docker.

## Instalasi sqlc dan Goose

Jalankan melalui Git Bash:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.0
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Pastikan folder binary Go sudah berada di dalam `PATH`.

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Periksa instalasi:

```bash
sqlc version
goose -version
```

## Mengambil project

```bash
git clone https://github.com/ki1bot/aksesibilitas-website.git
cd aksesibilitas-website
```

## Menginstal dependency

```bash
pnpm install
go mod download
```

## Menyiapkan environment

Salin file environment root:

```bash
cp .env.example .env
```

Salin environment frontend:

```bash
cp apps/web/.env.local.example apps/web/.env.local
```

Contoh konfigurasi `.env`:

```env
APP_ENV="development"
API_ADDR=":8080"
DATABASE_URL="postgres://aksescheck:CHANGE_ME@localhost:5432/aksesibilitaswebsite?sslmode=disable"
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD=""
REDIS_DB="0"
WEB_ORIGIN="http://localhost:3000"
SESSION_COOKIE_NAME="aksesibilitaswebsite_session"
SESSION_TTL="168h"
SCAN_TIMEOUT="60s"
SCAN_QUEUE="scanner"
WORKER_CONCURRENCY="2"
CHROME_PATH="C:/Program Files/Google/Chrome/Application/chrome.exe"
NEXT_PUBLIC_API_URL="http://localhost:8080/api/v1"
NEXT_PUBLIC_CSRF_COOKIE_NAME="aksesibilitaswebsite_session_csrf"
```

Contoh `apps/web/.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_CSRF_COOKIE_NAME=aksesibilitaswebsite_session_csrf
```

Jangan menyimpan password asli di dalam `.env.example` atau `.env.local.example`.

File `.env` dan `.env.local` hanya digunakan pada perangkat lokal dan tidak boleh dimasukkan ke repository.

## Menyiapkan PostgreSQL

Masuk ke PostgreSQL menggunakan akun administrator:

```bash
psql -U postgres
```

Buat user dan database:

```sql
CREATE USER aksescheck WITH PASSWORD 'GANTI_DENGAN_PASSWORD_YANG_KUAT';
CREATE DATABASE aksesibilitaswebsite OWNER aksescheck;
```

Keluar dari PostgreSQL:

```sql
\q
```

Sesuaikan `DATABASE_URL` pada `.env`:

```env
DATABASE_URL="postgres://aksescheck:GANTI_DENGAN_PASSWORD_YANG_KUAT@localhost:5432/aksesibilitaswebsite?sslmode=disable"
```

Karakter khusus pada username atau password harus diubah menjadi format URL encoding.

## Menjalankan migration

Muat `DATABASE_URL` dari `.env` ke Git Bash:

```bash
export DATABASE_URL="$(sed -n 's/^DATABASE_URL=//p' .env | tr -d '\r' | sed 's/^"//;s/"$//')"
```

Pastikan environment tersedia:

```bash
if [ -z "$DATABASE_URL" ]; then
  echo "DATABASE_URL tidak ditemukan pada file .env"
  exit 1
fi
```

Uji koneksi:

```bash
psql "$DATABASE_URL" -c "SELECT current_user, current_database();"
```

Jalankan migration:

```bash
goose -dir db/migrations postgres "$DATABASE_URL" status
goose -dir db/migrations postgres "$DATABASE_URL" up
```

## Menyinkronkan axe-core

```bash
pnpm sync:axe
```

Perintah tersebut menyalin sumber axe-core dari dependency ke:

```text
embedded/axe.min.js
```

Jalankan kembali perintah tersebut setelah versi axe-core berubah.

## Menghasilkan kode

Hasilkan query Go dari sqlc:

```bash
sqlc generate
```

Hasilkan tipe TypeScript dari OpenAPI:

```bash
pnpm generate:api
```

Jalankan keduanya sekaligus:

```bash
pnpm generate
```

File hasil generate tidak boleh diedit secara manual:

```text
internal/database/sqlcgen
apps/web/src/lib/api/schema.d.ts
```

## Menjalankan aplikasi

Pastikan PostgreSQL dan Redis sudah aktif.

### Terminal pertama — API

```bash
pnpm dev:api
```

API tersedia pada:

```text
http://localhost:8080
```

Health endpoint:

```text
http://localhost:8080/api/v1/health
```

Selama API berjalan, terminal akan tetap digunakan oleh proses `go run`.

### Terminal kedua — Worker

```bash
pnpm dev:worker
```

Worker akan mengambil task dari antrean Redis dengan nama yang ditentukan oleh `SCAN_QUEUE`.

### Terminal ketiga — Frontend

```bash
pnpm dev:web
```

Frontend tersedia pada:

```text
http://localhost:3000
```

## Halaman frontend

```text
/
 /register
 /login
 /dashboard
 /projects/{projectId}
 /scans/{scanId}
```

## API utama

### Authentication

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/me
POST /api/v1/auth/logout
```

### Projects

```text
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/{projectId}/
PATCH  /api/v1/projects/{projectId}/
DELETE /api/v1/projects/{projectId}/
```

### Scans

```text
GET    /api/v1/scans
POST   /api/v1/scans
GET    /api/v1/scans/{scanId}/
DELETE /api/v1/scans/{scanId}/
POST   /api/v1/scans/{scanId}/cancel
POST   /api/v1/scans/{scanId}/retry
GET    /api/v1/scans/{scanId}/violations
GET    /api/v1/scans/{scanId}/manual-review
POST   /api/v1/scans/{scanId}/reports
```

### Violations

```text
GET   /api/v1/violations/{violationId}/
PATCH /api/v1/violations/{violationId}/
```

### Manual review

```text
PATCH /api/v1/manual-review/items/{itemId}
```

### Reports

```text
GET /api/v1/reports/{reportId}
GET /api/v1/reports/{reportId}/download
```

## Perlindungan SSRF

Scanner menerapkan pemeriksaan berikut:

- Hanya menerima protokol HTTP dan HTTPS
- Menolak URL dengan kredensial
- Menolak localhost
- Menolak domain `.localhost` dan `.local`
- Menolak alamat loopback
- Menolak alamat private
- Menolak alamat link-local
- Menolak multicast
- Menolak alamat dokumentasi
- Menolak alamat cloud metadata
- Memeriksa seluruh hasil A dan AAAA dari DNS
- Memeriksa kembali setiap request Chromium
- Memeriksa kembali setiap redirect
- Membatasi jumlah request halaman
- Membatasi ukuran data halaman
- Membatasi jumlah navigasi dan redirect
- Menolak proses download
- Menggunakan profile Chromium sementara

Validasi SSRF mengurangi risiko, tetapi worker produksi tetap sebaiknya dijalankan di lingkungan terisolasi dengan akses jaringan yang dibatasi.

## Pengujian

Jalankan formatter Go:

```bash
pnpm format:go
```

Jalankan Go vet:

```bash
pnpm vet:go
```

Jalankan unit test Go:

```bash
pnpm test:go
```

Jalankan unit test dengan race detector:

```bash
pnpm test:go:race
```

Jalankan ESLint:

```bash
pnpm lint:web
```

Jalankan pemeriksaan TypeScript:

```bash
pnpm typecheck:web
```

Jalankan production build frontend:

```bash
pnpm build:web
```

Jalankan seluruh pemeriksaan Go:

```bash
pnpm check:go
```

Jalankan seluruh pemeriksaan frontend:

```bash
pnpm check:web
```

Jalankan seluruh pemeriksaan project:

```bash
pnpm check
```

## Production build frontend

```bash
pnpm build:web
pnpm --filter web start
```

API dan worker perlu dibangun secara terpisah:

```bash
go build -o bin/aksescheck-api.exe ./apps/api/cmd/api
go build -o bin/aksescheck-worker.exe ./apps/scanner/cmd/scanner
```

Jalankan API:

```bash
./bin/aksescheck-api.exe
```

Jalankan worker:

```bash
./bin/aksescheck-worker.exe
```

## Alur scan

1. Pengguna memilih project dan memasukkan URL.
2. API memvalidasi URL dan membatasi request pengguna.
3. API membuat data scan dengan status `queued`.
4. API mengirim task ke Redis melalui Asynq.
5. Worker mengambil task.
6. Worker mengubah status menjadi `running`.
7. Worker membuka Chromium dengan profile sementara.
8. Worker memeriksa seluruh network request.
9. Worker menyuntikkan axe-core.
10. Worker menjalankan aturan WCAG yang didukung.
11. Worker menyimpan halaman, violation, dan node ke PostgreSQL.
12. Worker membuat checklist pemeriksaan manual.
13. Worker menghitung skor otomatis.
14. Worker mengubah status menjadi `completed`.
15. Frontend berhenti melakukan polling dan menampilkan hasil.

## Continuous Integration

Workflow GitHub Actions berada di:

```text
.github/workflows/ci.yml
```

CI menjalankan:

- sqlc generation
- pemeriksaan perubahan file hasil sqlc
- gofmt validation
- Go vet
- Go test dengan race detector
- OpenAPI TypeScript generation
- pemeriksaan perubahan schema TypeScript
- ESLint
- TypeScript validation
- Next.js production build

## Batas MVP

Versi MVP memindai satu halaman per scan.

Crawler multi-page, penjadwalan scan otomatis, organisasi multi-tenant lanjutan, browser matrix, visual regression, dan integrasi CI eksternal belum menjadi bagian dari MVP.
