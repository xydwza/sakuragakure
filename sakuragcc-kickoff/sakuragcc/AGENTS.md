# Instruksi untuk agent: SakuraGCC

Kamu membangun portal warga RT 06 / RW 28 Cluster Sakura, Grand Cikarang City, di repo `/devzone/sakuragcc`. Pemilik proyek: Al. Penggunanya bapak-ibu warga perumahan yang membuka dari link grup WhatsApp di HP.

## Baca dulu, berurutan
1. `docs/SPEC.md`: spesifikasi lengkap. Ini sumber kebenaran.
2. `docs/mockup-nature.html`: **mockup utama, ini target desain.** Buka dan klik semua 8 peran lewat pemilih di kanan atas: Publik, Warga, Koordinator, Bendahara, Sekretaris, Ketua, Admin, RW/Desa. Tiru layout, warna, tekstur kayu, copy, dan perilakunya. `docs/mockup.html` adalah versi lama bertema minimal, hanya pembanding alur. Data iuran, pengeluaran, dan undangan di mockup adalah simulasi.
3. `seed/warga_seed.csv`: data nyata 120 rumah. Perhatikan kolom `catatan_validasi`.
4. `docs/OCR_data_sakuragakure.md`: seluruh data nyata yang diekstrak dari web lama, termasuk AD/ART lengkap, angka kas, inventaris, dan daftar data yang masih bolong.

## Aturan keras
1. Stack tetap: Go + chi + templ + htmx (di-vendor) + SQLite (`modernc.org/sqlite`) + sqlc + goose + Podman. Jangan menambah Node, framework JS, ORM, Redis, Postgres, Tailwind, atau CDN runtime tanpa persetujuan Al.
2. Uang selalu INTEGER rupiah. Saldo selalu dihitung dari tabel `mutasi`, tidak pernah disimpan sebagai kolom.
3. Mutasi tidak pernah di-UPDATE atau DELETE. Koreksi = mutasi baru. Periode tertutup dikunci dengan trigger database, bukan hanya di kode.
4. Pengeluaran tanpa nota harus gagal di level database (CHECK) dan di form.
5. Halaman publik tidak boleh memuat nama warga yang terkait status bayar, penerima dansos, alasan medis, umur, gender, atau nomor WA. Tulis test yang membuktikannya.
6. Setiap tabel pengurus wajib punya tombol ekspor TSV/CSV (lihat halaman "Sinkron data" di mockup). Pengurus bekerja dengan spreadsheet, portal harus menyuapi spreadsheet, bukan melawannya.
7. Nilai dari AD/ART (nominal iuran, santunan, batas tanggal) diambil dari tabel `setting`, tidak di-hardcode.
8. Mobile-first: setiap layar harus bekerja di lebar 360 px tanpa scroll horizontal sebelum kamu merapikan versi desktop.
9. Copy Bahasa Indonesia sederhana, sentence case, tombol menyebut hasilnya. Ikuti contoh copy di mockup.
10. File upload tidak pernah disajikan sebagai static folder. Semua lewat handler yang mengecek hak akses. EXIF dibuang lewat re-encode.
11. Fitur yang bergantung WhatsApp harus tetap jalan saat gateway mati (outbox mengantri, login pakai kode cadangan).
12. Hal di SPEC bagian 3.1 (pertanyaan terbuka) jangan diputuskan sendiri. Implementasikan sebagai setting dengan default yang tertulis, dan catat di `docs/DECISIONS.md`.
13. Operasi keuangan yang menyentuh lebih dari satu tabel (terima setoran, tutup buku, cairkan dansos) wajib dalam satu transaksi DB.
14. Peran `admin` adalah teknis, bukan keuangan. Admin tidak boleh bisa mencatat mutasi, menerima setoran, atau tutup buku. Peran `rw` dan `perangkat_desa` hanya baca data agregat. Lihat SPEC bagian 2.1.

## Cara kerja
1. Sebelum menulis kode untuk satu fase, tulis rencana singkat di `docs/PLAN.md`: daftar file, migrasi, dan test yang akan dibuat. Tunggu konfirmasi Al bila rencana menyimpang dari SPEC.
2. Urutan fase 1:
   a. skeleton repo, `go.mod` (module `sakuragcc`), Makefile, Containerfile, `/healthz`
   b. migrasi skema lengkap + trigger + seed `pos_dana`/`setting`/`aturan_pasal` + test trigger
   c. perintah `import --dry-run` dan `import` untuk CSV
   d. auth OTP + kode cadangan + sesi + middleware peran (10 peran, termasuk admin/rw/perangkat_desa) + `createadmin`
   e. layout, design token, font self-host, komponen dasar dari mockup
   f. beranda dan kas publik (termasuk kelopak per gang dan grafik SVG)
   g. tarik iuran + setoran + konfirmasi bendahara
   h. mutasi + pengeluaran bernota + modul media
   i. rumahku, aturan, pengurus, galeri + posting kegiatan
   j. PWA + backup terjadwal + quadlet + Caddyfile
3. Commit kecil per langkah dengan pesan Bahasa Indonesia yang jelas. Jalankan `make test` sebelum setiap commit.
4. Setiap keputusan teknis yang tidak tertulis di SPEC dicatat di `docs/DECISIONS.md` (tanggal, keputusan, alasan, alternatif).
5. Setelah setiap langkah besar, laporkan: apa yang selesai, apa yang menyimpang dari mockup dan kenapa, kriteria selesai mana yang sudah lolos, dan perintah untuk mencobanya.

## Jangan
- Jangan membuat tabel bulan x rumah untuk publik (itu kesalahan web lama).
- Jangan menyimpan NIK atau nomor KK.
- Jangan mengirim pengingat tunggakan ke grup. Selalu pribadi.
- Jangan menambah fitur di luar fase yang sedang dikerjakan.
- Jangan mengubah warna, font, atau radius dari token di SPEC bagian 9.
- Jangan menebak. Kalau SPEC dan mockup tidak menjawab, tanya Al.
