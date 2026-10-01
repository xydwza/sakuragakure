# Log keputusan

Format: tanggal, keputusan, alasan, alternatif yang ditolak.

- 2026-10-01: SQLite (modernc) dipilih, bukan Postgres. Alasan: ~150 user, satu VPS, backup satu file, binary tanpa CGO. Alternatif: Postgres (berlebihan untuk skala ini).
- 2026-10-01: Login tanpa password via OTP WhatsApp + kode cadangan dari koordinator. Alasan: warga sering lupa password. Alternatif: password (ditolak), email (warga jarang pakai).
- 2026-10-01: Rumah dan penghuni dipisah. Alasan: pengontrak berganti, riwayat iuran tidak boleh rusak.
- 2026-10-01: Iuran tunai lewat koordinator tetap jadi alur utama, transfer opsional. Alasan: mengikuti AD/ART pasal 3.2 dan kebiasaan warga.
- 2026-10-01: Tambah peran `admin` (teknis, tanpa akses keuangan), `rw` dan `perangkat_desa` (baca agregat). Surat pengantar diverifikasi lewat QR ke /v/{kode} tanpa login. Alasan: pemisahan tugas dan kebutuhan verifikasi dari desa. Alternatif: superadmin serba bisa (ditolak, satu akun bisa mengubah kas tanpa jejak).
- 2026-10-01: Expose publik lewat **Cloudflare Tunnel** (cloudflared), bukan Caddy+IP publik seperti SPEC §12. Alasan: server `xydwza` tidak punya IP publik (di belakang NAT, akses SSH pun lewat Cloudflare Access), jadi domain `sakuragakure.my.id` diarahkan ke tunnel Cloudflare. Alternatif: Caddy langsung (ditolak, tidak ada IP publik); FRP/reverse tunnel lain (ditolak, Cloudflare sudah jadi pola standar di server ini).
- 2026-10-01: `cloudflared` untuk Sakura jalan sebagai **kontainer Podman rootless** (`cloudflared-sakuragakure`, network=host, restart=always), terpisah dari 5 tunnel root yang sudah ada. Alasan: isolasi penuh, tanpa sudo, tidak menyentuh konfigurasi root. Tunnel id `9e8ec7ed-d201-4ea7-8e6e-22732410ae72`, ingress `sakuragakure.my.id` → `http://127.0.0.1:8081` (portal dipublikasikan ke loopback saja).
- 2026-10-01: `pos_dana.kas_gang_1..3` diberi `publik=0` (tidak tampil di kas publik). Alasan: mockup hanya menampilkan kas RT, dana sosial, dan rukem. Kas gang adalah alokasi internal.
- 2026-10-01: CHECK pengeluaran-wajib-nota di SPEC §6 memakai `COALESCE(ref_tipe,'') IN (...)` alih-alih `ref_tipe IN (...)`. Alasan: di SQLite `NULL IN (...)` bernilai NULL dan CHECK hanya gagal pada FALSE, sehingga `arah='keluar'` tanpa nota dengan `ref_tipe` NULL akan lolos. COALESCE memastikan NULL diperlakukan sebagai "bukan alokasi/koreksi/impor". Ini koreksi agar aturan AGENTS #4 benar-benar terpenuhi.
