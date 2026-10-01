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
