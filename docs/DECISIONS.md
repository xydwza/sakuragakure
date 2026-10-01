# Log keputusan

Format: tanggal, keputusan, alasan, alternatif yang ditolak.

- 2026-10-01: SQLite (modernc) dipilih, bukan Postgres. Alasan: ~150 user, satu VPS, backup satu file, binary tanpa CGO. Alternatif: Postgres (berlebihan untuk skala ini).
- 2026-10-01: Login tanpa password via OTP WhatsApp + kode cadangan dari koordinator. Alasan: warga sering lupa password. Alternatif: password (ditolak), email (warga jarang pakai).
- 2026-10-01: Rumah dan penghuni dipisah. Alasan: pengontrak berganti, riwayat iuran tidak boleh rusak.
- 2026-10-01: Iuran tunai lewat koordinator tetap jadi alur utama, transfer opsional. Alasan: mengikuti AD/ART pasal 3.2 dan kebiasaan warga.
- 2026-10-01: Tambah peran `admin` (teknis, tanpa akses keuangan), `rw` dan `perangkat_desa` (baca agregat). Surat pengantar diverifikasi lewat QR ke /v/{kode} tanpa login. Alasan: pemisahan tugas dan kebutuhan verifikasi dari desa. Alternatif: superadmin serba bisa (ditolak, satu akun bisa mengubah kas tanpa jejak).
