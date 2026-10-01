# Rencana per fase

Format: daftar file, migrasi, dan test yang akan dibuat per fase. Ditulis sebelum menulis kode.

## Fase 1a — skeleton repo, go.mod, Makefile, Containerfile, /healthz

Selesai.

## Fase 1b — migrasi skema lengkap + trigger + seed + test trigger

Tujuan: skema SQLite lengkap (SPEC §6), kunci periode via trigger, seed data awal, dan perintah `migrate`.

File:
- `internal/db/migrations/00001_schema.sql` — semua tabel + indeks + trigger (salin persis SPEC §6)
- `internal/db/migrations/00002_seed.sql` — seed `pos_dana`, `setting`, `aturan_pasal`, `inventaris`, `jadwal_rutin`
- `internal/db/db.go` — buka koneksi (modernc.org/sqlite, WAL, busy_timeout), embed migrasi, `Migrate()`
- `cmd/sakuragakure/main.go` — tambah subcommand `migrate`
- `internal/db/migrate_test.go` — test trigger di SQLite in-memory

Migrasi: `00001_schema.sql`, `00002_seed.sql` (goose, embed, hanya maju).

Test: kunci periode insert/update/delete, periode_tutup tak bisa dibuka, pengeluaran tanpa nota gagal (CHECK).

Kriteria selesai: `sakuragakure migrate` membuat DB lengkap; test trigger lolos; `make test` hijau.

Keputusan yang dicatat di DECISIONS.md:
- `pos_dana.kas_gang_1..3` diberi `publik=0` (tidak tampil di kas publik, sesuai mockup yang hanya menampilkan kas RT, dana sosial, rukem).
- Setting tambahan dari §2.1 & §3.1: `nama_rt`, `rukem_mode`, `dansos_per`.

Selesai. Catatan: go mod di-pin ke goose v3.24.0 + modernc.org/sqlite v1.39.0 agar tetap kompatibel Go 1.23 (Containerfile pakai golang:1.23-alpine); versi terbaru menuntut Go 1.26.

## Fase 1c — perintah import --dry-run dan import CSV

Tujuan: membaca `seed/warga_seed.csv` menjadi `rumah` + `penghuni` aktif, dengan laporan dry-run.

File:
- `internal/rumah/import.go` — baca CSV, validasi/laporan (duplikat KK, catatan validasi, ringkasan per gang), `Impor()` idempotent
- `cmd/sakuragakure/main.go` — subcommand `import --rumah <csv> --db <path> [--dry-run]`
- `internal/rumah/import_test.go` — test parse, laporan, impor (tahun_lahir dari umur, perkiraan, gender NULL, "Penghuni G10/14"), idempotensi

Aturan konversi (SPEC §13): tahun_lahir = 2026 - umur (perkiraan=1); umur kosong → NULL (perkiraan=0); gender kosong → NULL; tetap→pemilik, kontrak→kontrak; kosong tanpa penghuni; nama kosong → "Penghuni <alamat>"; mulai = 2026-04-01.

Kriteria selesai: `import --dry-run` mencetak ringkasan per gang/status, baris catatan validasi, duplikasi KK; `import` menulis 120 rumah + penghuni; `make test` hijau.

Selesai.

## Fase 1d — auth OTP + kode cadangan + sesi + middleware peran + createadmin

Tujuan: login tanpa password (OTP WA 6 digit + kode cadangan), sesi 180 hari, CSRF, dan middleware 10 peran.

File:
- `internal/auth/otp.go` — generate (6 digit, hash SHA-256), verify (expiry 5 mnt, maks 5 percobaan), throttle per nomor 3/15 mnt
- `internal/auth/kode.go` — kode cadangan sekali pakai (24 jam), hash + verify
- `internal/auth/session.go` — scs + store SQLite custom (modernc, tanpa CGO; store bawaan scs pakai mattn/CGO)
- `internal/auth/middleware.go` — `RequireLogin`, `RequirePeran` (10 peran)
- `internal/auth/handlers.go` — POST /masuk/wa, POST /masuk/kode, GET /keluar
- `internal/notif/notif.go` — `Enqueue` ke `notif_outbox` (template otp); log kode bila WA gateway kosong (dev)
- `cmd/sakuragakure/main.go` — subcommand `createadmin`; `serve` pasang sesi + CSRF + route auth
- test: `otp_test.go`, `kode_test.go`, `middleware_test.go`

Dependensi baru: `github.com/alexedwards/scs/v2`, `github.com/justinas/nosurf` (keduanya pure Go).

Kriteria selesai: OTP & kode cadangan terverifikasi (benar/salah/kedaluwarsa/percobaan habis); middleware tamu→redirect /masuk, peran salah→403; `createadmin` membuat user admin; `make test` hijau.

Catatan: antrian `notif_outbox` sudah ditulis di fase ini, worker pengirim WA-nya di fase 1j (gateway mati → kode tetap dibuat, login pakai kode cadangan).

Selesai.

## Fase 1e — layout, design token, font self-host, komponen dasar

Tujuan: pondasi visual dari mockup — token warna/radius/font, shell (top bar + bottom nav + rail), komponen dasar, htmx + font di-vendor.

File:
- `internal/web/static/app.css` — token + base + komponen (salin mockup `<style>` §10-211)
- `internal/web/static/vendor/htmx.min.js` — htmx 2 (di-vendor)
- `internal/web/static/fonts/*.woff2` — Plus Jakarta Sans 400/500/600/700/800 + italic 500
- `internal/web/static/embed.go` — embed static + handler (cache, content-type)
- `internal/web/components/*.templ` — layout shell, ikon, komponen (btn, pill, card, pos)
- `internal/web/pages/beranda.templ` — halaman sementara untuk bukti pipeline
- `cmd/sakuragakure/main.go` — `serve` sajikan static + route `/` sementara

Dependensi/tool baru: `github.com/a-h/templ` (tool generate, di-pin di Containerfile).

Kriteria selesai: `make generate` + `make build` jalan; `/` merender shell dengan token sakura; font & htmx dilayani lokal tanpa CDN; `make test` hijau.

Catatan: navigasi per peran + pengalih peran diisi fase 1f (butuh halaman beranda/kas).

Selesai.

## Fase 1f — beranda dan kas publik (kelopak per gang + grafik SVG)

Selesai.

## Fase 1g — tarik iuran + setoran + konfirmasi bendahara

Tujuan: alur keuangan utama (SPEC §7 alur A) — koordinator menandai rumah, menyetor, bendahara menerima.

File:
- `internal/db/queries/iuran.sql` — grid gang, setoran menunggu, dll (sqlc)
- `internal/iuran/iuran.go` — `GenerateTagihan`, `TogglePembayaran`, `Setor`, `TerimaSetoran`, `TolakSetoran` (transaksi DB untuk operasi multi-tabel)
- `internal/iuran/iuran_test.go` — state machine pembayaran + transaksi + saldo
- `internal/web/pages/tarik.templ`, `setor.templ`, `kelola_setoran.templ`
- `internal/web/handlers.go` — route koordinator/bendahara + middleware peran + nav
- `cmd/sakuragakure/main.go` — subcommand `tagihan --periode`

State machine: `dipegang -> disetor -> diterima`; `disetor -> dipegang` hanya lewat tolak; tap membatalkan hanya saat `dipegang`.

Kriteria selesai: koordinator bisa menandai 40 rumah + setor; bendahara terima → saldo + kelopak berubah; state machine diuji; `make test` hijau.

Selesai.

## Fase 1h — mutasi + pengeluaran bernota + modul media

Tujuan: bendahara mencatat pengeluaran (wajib foto nota) dan media dilayani lewat handler ber-hak akses.

File:
- `internal/media/media.go` — upload (decode → re-encode JPEG q82 buang EXIF, full 1920 + thumb 480), `Serve` dengan cek akses
- `internal/media/media_test.go` — upload membuat full + thumb, resize
- `internal/kas/pengeluaran.go` — `CatatPengeluaran` (mutasi keluar bernota)
- `internal/db/queries/kas.sql` — `MutasiSemua` (pengurus, dengan nota_media_id); `media.sql` — `MediaByID`
- `internal/web/pages/mutasi.templ`, `mutasi_baru.templ`
- `internal/web/handlers.go` — `/kelola/mutasi`, `/kelola/mutasi/baru`, `/m/{id}/{full|thumb}`
- `cmd/sakuragakure/main.go` — route + `MEDIA_DIR`

Dependensi baru: `github.com/disintegration/imaging`.

Kriteria selesai: pengeluaran tanpa nota gagal (CHECK); nota tersimpan + thumbnail; media akses `warga` tidak bisa diakses tamu; `make test` hijau.
