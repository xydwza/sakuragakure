# Rencana per fase

Format: daftar file, migrasi, dan test yang akan dibuat per fase. Ditulis sebelum menulis kode.

## Fase 1a — skeleton repo, go.mod, Makefile, Containerfile, /healthz

Tujuan: repo bisa dibangun, `make test` jalan, binary menyala dan menjawab `/healthz`.

File:
- `go.mod` — module `sakuragakure`, `go 1.23.0`
- `cmd/sakuragakure/main.go` — subcommand `serve` (stdlib `flag`), router chi, handler `/healthz`
- `Makefile` — target `build`, `run`, `test` (go vet + go test), `generate` (templ, diisi fase 1e)
- `deploy/Containerfile` — salin dari SPEC §12, nama `sakuragakure`
- `.gitignore` — binary, `data/`, `deploy/.env`, artefak

Migrasi: belum ada (fase 1b).

Test: belum ada logika yang perlu diuji. `/healthz` trivial (ponytail: tanpa test).

Kriteria selesai: `make build` menghasilkan binary; `curl :8080/healthz` → 200; `make test` lolos.
