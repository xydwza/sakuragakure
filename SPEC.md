# SakuraGCC: Portal Warga RT 06 / RW 28

Spesifikasi lengkap untuk membangun portal warga RT 006 / RW 028, Cluster Sakura, Perumahan Grand Cikarang City, Desa Karangraharja, Kecamatan Cikarang Utara, Kabupaten Bekasi.

Dokumen ini adalah sumber kebenaran. Kalau ada konflik antara dokumen ini dan asumsi agent, dokumen ini yang menang. Kalau ada yang tidak tercakup, tanya dulu, jangan menebak.

Referensi visual: dua berkas, buka keduanya di browser.
- `docs/mockup-nature.html` — **mockup utama, tema nature (kayu, daun, kertas). Inilah target desain final.** Memuat 8 peran, struktur pengurus bergaya papan kayu, pengelolaan dana multi-pos, undangan RT dengan RSVP, rukem, sinkron spreadsheet, dan manajemen pengguna.
- `docs/mockup.html` — mockup pertama, tema minimal. Dipakai hanya sebagai pembanding alur, bukan acuan visual.

Semua layar, warna, copy, dan alur di mockup nature adalah target desain.

---

## 0. Konteks dan masalah

Web lama (`rtjepri.netlify.app`) dikeluhkan warga karena:

1. "Data Kas" hanya matriks iuran 14 kolom x 120 baris. Tidak ada saldo, tidak ada pengeluaran, tidak ada nota.
2. Privasi bocor: pengunjung tanpa login bisa melihat nama, alamat, status bayar per rumah, dan penerima dana sosial beserta alasan medisnya.
3. Tidak terbaca di HP, padahal warga membuka dari link grup WhatsApp.
4. Data kotor: status campur huruf besar-kecil, "Kosong" dipakai sebagai nama, font unicode aneh, satu orang punya dua rumah tercatat ganda.
5. AD/ART berupa PDF slide yang di-embed. Tidak bisa dicari.
6. Agenda lama tidak hilang, galeri hanya carousel.

Skala: 120 unit rumah, 3 gang, blok G10, G11, G12, G19, G20, G21. Sekitar 87 tetap, 18 kontrak, 15 kosong. Pengguna aktif maksimal ~150 orang. Beban sangat kecil, jadi kesederhanaan operasional lebih penting dari skalabilitas.

## 1. Prinsip produk (tidak boleh dilanggar)

1. **Transparansi kas adalah fitur utama.** Setiap pengeluaran wajib punya foto nota. Bulan yang sudah ditutup buku tidak bisa diubah. Koreksi selalu berupa mutasi baru (jurnal balik), tidak pernah edit atau hapus.
2. **Privasi bertingkat.** Publik hanya melihat agregat. Warga hanya melihat rumahnya sendiri. Koordinator melihat gangnya. Pengurus inti melihat semua.
3. **Mobile-first.** Didesain untuk layar 360 sampai 414 px dulu. Desktop adalah bonus.
4. **Mengikuti alur nyata, bukan memaksa alur baru.** Iuran ditarik tunai oleh koordinator gang tanggal 1 sampai 10, disetor ke bendahara paling lambat tanggal 15. Sistem mencatat alur ini, bukan menggantinya dengan pembayaran online wajib.
5. **Satu aksi, satu ketukan.** Koordinator mencatat iuran dengan mengetuk kotak rumah. Bendahara mengonfirmasi setoran dengan satu tombol.
6. **Aturan AD/ART menjadi logika sistem**, dan nilainya disimpan di tabel `setting` agar bisa diubah tanpa deploy.
7. **Minim dependensi, satu binary, offline-friendly.** Tidak ada build Node. Tidak ada CDN di runtime.

## 2. Peran dan hak akses

Satu user bisa punya beberapa peran (misalnya koordinator gang 3 sekaligus warga G12/03).

| Kemampuan | Publik | Warga | Koordinator gang | Sekretaris | Bendahara | Ketua / Wakil |
|---|---|---|---|---|---|---|
| Beranda, agenda, galeri, AD/ART, struktur | ya | ya | ya | ya | ya | ya |
| Saldo per pos, grafik, daftar mutasi (tanpa nota) | ya | ya | ya | ya | ya | ya |
| Progress iuran per gang (agregat, tanpa alamat) | ya | ya | ya | ya | ya | ya |
| Lihat foto nota | | ya | ya | ya | ya | ya |
| Status iuran rumah sendiri, ajukan layanan | | ya | ya | ya | ya | ya |
| Checklist iuran per rumah di gangnya, setor | | | gang sendiri | | | |
| Lihat nama + status bayar per rumah | | rumah sendiri | gang sendiri | | semua | semua |
| Konfirmasi setoran, catat pengeluaran, tutup buku | | | | | ya | lihat |
| Detail dana sosial (nama, alasan, bukti) | | pengajuan sendiri | | ya | ya | ya |
| Setujui dana sosial dan surat | | | | | | ya |
| Data rumah dan penghuni (umur, gender, anggota) | | | gang sendiri (nama saja) | ya | ya | ya |
| Posting kegiatan, agenda, pengumuman | | | | ya | | ya |
| Kelola user dan peran | | | | | | ketua |
| Audit log | | | | | ya | ya |
| Kelola user, setting, impor, status sistem | | | | | | ketua (user saja); admin semua |

Struktur pengurus di halaman publik dibangkitkan dari tabel peran, bukan gambar.

### 2.1 Peran di luar pengurus RT

| Peran | Untuk siapa | Boleh | Tidak boleh |
|---|---|---|---|
| `admin` | pengelola teknis aplikasi (developer) | kelola user dan peran (termasuk menetapkan ketua), setting, impor data, lihat audit log, status backup, status gateway WA, outbox | membuat atau mengubah mutasi, konfirmasi setoran, tutup buku, menyetujui dansos/surat. Admin bukan pengurus keuangan |
| `rw` | ketua RW 028 dan pengurus RW | baca semua laporan kas yang sudah ditutup + PDF-nya, rekap kependudukan agregat (jumlah KK, jiwa, tetap/kontrak/kosong per gang), daftar surat pengantar yang terbit, agenda | data per rumah, nama penunggak, detail dansos, mengubah apa pun |
| `perangkat_desa` | petugas Desa Karangraharja | rekap kependudukan agregat, laporan kas tertutup, verifikasi surat pengantar | sama seperti `rw` |

Aturan:
- Akun `rw` dan `perangkat_desa` dibuat oleh ketua atau admin, login dengan cara yang sama (OTP WA), dan bisa diberi tanggal kedaluwarsa (`user.aktif_sampai`).
- Setiap surat pengantar PDF memuat kode verifikasi dan QR menuju `/v/{kode}`. Halaman ini publik dan hanya menampilkan: nomor surat, keperluan, nama pemohon dengan sensor (contoh "Ag*** Ra*****"), alamat blok, tanggal terbit, status sah/dibatalkan. Perangkat desa tidak perlu login untuk memverifikasi.
- Pemisahan tugas: tidak ada satu akun pun yang bisa mencatat pengeluaran sekaligus menutup buku tanpa jejak. Admin tidak pernah menyentuh uang. Bendahara tidak bisa mengubah peran user.
- Multi-RT (seluruh RW 028) belum dikerjakan. Tetapi semua tabel inti sudah aman ditambah kolom `rt_id` nanti; jangan hardcode "RT 06" di logika, ambil dari `setting.nama_rt`.

Pengurus saat ini (AD/ART 11 April 2026): Ketua Jepri Kiat Susanto, Wakil Wiyanto, Sekretaris Ferry, Bendahara Danu. Koordinator gang 1 Ipan Sovan. Gang 2 Ajat dengan pembantu Wahyu dan Haryanto. Gang 3 Bambang Subekti dengan pembantu Sagito dan Muhlis.

## 3. Aturan bisnis dari AD/ART

Semua nominal dan batas tanggal disimpan di tabel `setting`.

| Kunci setting | Nilai awal | Sumber |
|---|---|---|
| `iuran_kas_bulanan` | 25000 | pasal 3.2 ayat 2 |
| `iuran_tarik_mulai` / `iuran_tarik_sampai` | 1 / 10 | pasal 3.2 ayat 2 |
| `setor_batas_tanggal` | 15 | pasal 3.2 ayat 2 |
| `tunggakan_blokir_bulan` | 3 | pasal 3.2 ayat 9 |
| `dansos_rawat_inap` | 200000 | pasal 3.4 a (tidak untuk perawatan di rumah) |
| `dansos_melahirkan` | 200000 | pasal 3.4 b |
| `dansos_duka_ortu` | 250000 (dari kas gang) | pasal 3.4 c |
| `dansos_maks_per_tahun` | 1 | pasal 3.4 ayat 3 |
| `rukem_iuran` | 50000 | pasal 3.5 |
| `rukem_santunan` | 2500000 | pasal 3.5 |
| `hajatan_min_hari` | 3 | pasal 3.3 |
| `tamu_lapor_jam` | 24 | pasal 3.7 ayat 2 |
| `tinggal_rumah_lapor_jam` | 72 | pasal 3.7 ayat 3 |
| `warga_baru_lapor_jam` | 72 | pasal 3.2 ayat 4 |

Logika:

- **Wajib iuran** = rumah berstatus `tetap` atau `kontrak` dan tidak punya flag bebas iuran. Rumah `kosong` tidak ditagih.
- **Bebas iuran** disimpan per rumah beserta alasan (contoh dari data lama: G10/02 koordinator gang 1, G10/12 rumah kedua milik KK yang sama, G21/04 wakil ketua). Daftar lengkap dikonfirmasi ke pengurus.
- **Tunggakan** = jumlah bulan berturut-turut tanpa pembayaran, dihitung mundur dari bulan terakhir yang sudah lewat tanggal 10. Bulan berjalan tidak dihitung sebelum tanggal 11.
- **Blokir layanan**: tunggakan >= `tunggakan_blokir_bulan` membuat pengajuan surat dan dana sosial ditolak otomatis dengan pesan sopan yang menyebut pasal 3.2 ayat 9 dan nominal yang perlu dilunasi.
- **Dana sosial**: cek otomatis (a) tidak diblokir tunggakan, (b) belum pernah menerima dalam tahun kalender berjalan, (c) jenis valid. Nominal diisi otomatis dari setting dan tidak bisa diubah warga. Ketua atau wakil yang menyetujui. Setelah disetujui, bendahara mencairkan, dan pencairan menjadi mutasi keluar dari pos `dana_sosial` (atau `kas_gang_N` untuk duka orang tua) dengan bukti.
- **Hajatan**: tanggal acara minimal H+3 dari hari pengajuan. Bisa sekalian meminjam inventaris (panggung, sound system Politron + 2 mic wireless, terpal 8 m, gerobak). Peminjaman bentrok tanggal ditolak.
- **Tamu**: laporan menyimpan nama, hubungan, tanggal datang dan pulang, foto KTP opsional (akses pengurus saja, dihapus otomatis 30 hari setelah tanggal pulang).

### 3.1 Pertanyaan terbuka (wajib dikonfirmasi pengurus, implementasikan sebagai setting atau flag)

1. **Rukem Rp50.000**: sekali saat menjadi warga, per tahun, atau ditarik setiap ada kematian? Data lama semuanya nol. Default implementasi: tipe tagihan `rukem` dengan mode `sekali | per_kejadian | bulanan` di setting, default `per_kejadian`, dan fitur rukem disembunyikan sampai dikonfirmasi.
2. **Dana sosial per orang atau per KK?** Data lama ada penyaluran Rp400.000 untuk "anak dan istri dirawat di RS", sedangkan aturan Rp200.000. Default: per orang yang dirawat, maksimal satu kali per orang per tahun. Konfirmasi.
3. **Kas gang** untuk duka orang tua: sumbernya dari mana? Default: pos `kas_gang_1..3` diisi manual lewat alokasi dari kas RT.
4. **Daftar lengkap bebas iuran.**
5. **Rumah dengan data konflik** (lihat kolom `catatan_validasi` di `seed/warga_seed.csv`), misalnya G11/14 berstatus tetap tapi nama "Kosong", G21/06 berstatus kosong tapi ada umur dan anggota.

## 4. Stack

Dipilih untuk satu orang yang memelihara, di satu VPS kecil, bertahun-tahun.

| Lapisan | Pilihan | Alasan |
|---|---|---|
| Bahasa | Go 1.23+ | Satu binary, cross-compile ke ARM kalau nanti pindah ke SBC |
| Router | `github.com/go-chi/chi/v5` | Ringan, standar net/http |
| Template | `github.com/a-h/templ` | Type-safe, komponen, tanpa runtime JS |
| Interaksi | htmx 2 (file di-vendor ke `web/static/vendor/`, bukan CDN) + sedikit vanilla JS | Tidak perlu SPA |
| Database | SQLite via `modernc.org/sqlite` (pure Go, tanpa CGO), mode WAL, `busy_timeout=5000` | 150 user tidak butuh Postgres. Backup = satu file |
| Query | `sqlc` untuk generate kode query bertipe | Tanpa ORM |
| Migrasi | `github.com/pressly/goose/v3` dengan SQL ter-embed | Jalan otomatis saat start |
| Sesi | `github.com/alexedwards/scs/v2` + store SQLite | Cookie aman, sesi panjang |
| CSRF | `github.com/justinas/nosurf` | |
| Gambar | `github.com/disintegration/imaging` + re-encode (otomatis membuang EXIF) | Resize, thumbnail, buang lokasi GPS |
| PDF | `github.com/go-pdf/fpdf` | Laporan bulanan dan surat pengantar |
| CSS | Satu file `app.css` tulisan tangan dengan token dari mockup | Tanpa Tailwind, tanpa build |
| Grafik | SVG dirender server-side di templ | Tanpa library chart |
| WhatsApp | `aldinokemal/go-whatsapp-web-multidevice` sebagai container terpisah, dipanggil lewat REST dari outbox | Self-hosted, nomor khusus RT |
| Reverse proxy | Caddy | HTTPS otomatis |
| Container | Podman rootless + Quadlet (systemd) | |
| Backup | `VACUUM INTO` terjadwal dari aplikasi + restic ke lokasi kedua | |

Tidak boleh ditambahkan tanpa izin: Node/npm build step, React/Vue, ORM, Redis, Postgres, CDN runtime, layanan SaaS berbayar.

## 5. Struktur repo

```
/devzone/sakuragcc/
├── AGENTS.md                 # instruksi untuk agent (baca dulu)
├── docs/
│   ├── SPEC.md               # dokumen ini
│   ├── mockup.html           # referensi visual dan alur
│   ├── PLAN.md               # rencana per fase (ditulis agent)
│   └── DECISIONS.md          # log keputusan (agent menambah entri)
├── seed/warga_seed.csv       # data rumah dan KK yang sudah dibersihkan
├── cmd/sakuragcc/main.go     # subcommand: serve, migrate, import, backup, createadmin
├── internal/
│   ├── config/               # env + setting dari DB
│   ├── db/                   # koneksi, migrasi (embed), sqlc output
│   │   ├── migrations/*.sql
│   │   └── queries/*.sql
│   ├── auth/                 # OTP WA, kode sekali pakai, sesi, middleware peran
│   ├── rumah/                # rumah, penghuni, pindah/ganti penghuni
│   ├── iuran/                # tagihan, pembayaran, setoran, tunggakan
│   ├── kas/                  # pos dana, mutasi, tutup buku, laporan
│   ├── layanan/              # dana sosial, surat, tamu, hajatan, aduan
│   ├── konten/               # album, foto, agenda, pengumuman, aturan
│   ├── inventaris/
│   ├── notif/                # outbox WA, template pesan, worker
│   ├── media/                # upload, resize, strip EXIF, akses terlindungi
│   ├── pdf/
│   ├── audit/
│   └── web/                  # handler, routing, templ components, middleware
│       ├── components/*.templ
│       ├── pages/*.templ
│       └── static/{app.css, app.js, fonts/, vendor/htmx.min.js, icons/, manifest.webmanifest, sw.js}
├── deploy/
│   ├── Containerfile
│   ├── sakuragcc.container   # Quadlet
│   ├── wa-gateway.container  # Quadlet
│   ├── Caddyfile
│   ├── backup.sh / backup.service / backup.timer
│   └── env.example
├── Makefile
└── go.mod                    # module sakuragcc
```

## 6. Skema database

Konvensi: uang = INTEGER rupiah (tidak pernah float). Tanggal = TEXT ISO `YYYY-MM-DD`, waktu = TEXT RFC3339 zona Asia/Jakarta. Periode = TEXT `YYYY-MM`. Nama tabel dan kolom memakai istilah AD/ART (rumah, penghuni, iuran) agar mudah dicocokkan dengan pengurus.

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE setting (key TEXT PRIMARY KEY, value TEXT NOT NULL, updated_at TEXT NOT NULL);

CREATE TABLE rumah (
  id INTEGER PRIMARY KEY,
  alamat TEXT NOT NULL UNIQUE,              -- format kanonik 'G10/01', 'G10/12A'
  blok TEXT NOT NULL, nomor TEXT NOT NULL,
  gang INTEGER NOT NULL CHECK (gang IN (1,2,3)),
  status TEXT NOT NULL CHECK (status IN ('tetap','kontrak','kosong')),
  bebas_iuran_alasan TEXT,                  -- NULL = wajib iuran
  catatan_validasi TEXT
);

CREATE TABLE penghuni (                     -- satu baris = satu KK pada satu periode tinggal
  id INTEGER PRIMARY KEY,
  rumah_id INTEGER NOT NULL REFERENCES rumah(id),
  nama_kk TEXT NOT NULL,
  gender TEXT CHECK (gender IN ('L','P')),
  tahun_lahir INTEGER,                      -- simpan tahun lahir, bukan umur
  tahun_lahir_perkiraan INTEGER NOT NULL DEFAULT 0,
  jumlah_anggota INTEGER NOT NULL DEFAULT 0,
  status_huni TEXT NOT NULL CHECK (status_huni IN ('pemilik','kontrak')),
  mulai TEXT NOT NULL, selesai TEXT,        -- selesai NULL = masih tinggal
  no_wa TEXT                                -- E.164, untuk login dan notifikasi
);
CREATE UNIQUE INDEX one_active_penghuni ON penghuni(rumah_id) WHERE selesai IS NULL;

CREATE TABLE user (
  id INTEGER PRIMARY KEY,
  nama TEXT NOT NULL,
  no_wa TEXT NOT NULL UNIQUE,
  penghuni_id INTEGER REFERENCES penghuni(id),
  aktif INTEGER NOT NULL DEFAULT 1,
  aktif_sampai TEXT,                        -- untuk akun rw / perangkat_desa sementara
  created_at TEXT NOT NULL
);
CREATE TABLE user_peran (
  user_id INTEGER NOT NULL REFERENCES user(id),
  peran TEXT NOT NULL CHECK (peran IN ('warga','koordinator','pembantu_koordinator','sekretaris','bendahara','wakil','ketua','admin','rw','perangkat_desa')),
  gang INTEGER NOT NULL DEFAULT 0,          -- 0 = tidak terikat gang; wajib 1..3 untuk koordinator/pembantu
  urutan INTEGER NOT NULL DEFAULT 0,        -- urutan tampil di struktur
  PRIMARY KEY (user_id, peran, gang)
);

CREATE TABLE pos_dana (
  id TEXT PRIMARY KEY,                      -- 'kas_rt','dana_sosial','rukem','kas_gang_1','kas_gang_2','kas_gang_3'
  nama TEXT NOT NULL, publik INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE media (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL, thumb_path TEXT,
  mime TEXT NOT NULL, ukuran INTEGER NOT NULL,
  akses TEXT NOT NULL CHECK (akses IN ('publik','warga','pengurus','pemilik')),
  pemilik_user_id INTEGER REFERENCES user(id),
  hapus_setelah TEXT,                       -- auto-delete (KTP tamu)
  dibuat_at TEXT NOT NULL
);

CREATE TABLE tagihan (
  id INTEGER PRIMARY KEY,
  rumah_id INTEGER NOT NULL REFERENCES rumah(id),
  penghuni_id INTEGER REFERENCES penghuni(id),
  jenis TEXT NOT NULL CHECK (jenis IN ('kas','rukem')),
  periode TEXT NOT NULL,
  nominal INTEGER NOT NULL,
  UNIQUE (rumah_id, jenis, periode)
);

CREATE TABLE setoran (
  id INTEGER PRIMARY KEY,
  gang INTEGER NOT NULL,
  periode TEXT NOT NULL,
  koordinator_id INTEGER NOT NULL REFERENCES user(id),
  total INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('menunggu','diterima','ditolak')),
  dibuat_at TEXT NOT NULL,
  dikonfirmasi_oleh INTEGER REFERENCES user(id), dikonfirmasi_at TEXT, catatan TEXT
);

CREATE TABLE pembayaran (
  id INTEGER PRIMARY KEY,
  tagihan_id INTEGER NOT NULL UNIQUE REFERENCES tagihan(id),
  metode TEXT NOT NULL CHECK (metode IN ('tunai','transfer','impor')),
  status TEXT NOT NULL CHECK (status IN ('dipegang','menunggu_verifikasi','disetor','diterima','ditolak')),
  dicatat_oleh INTEGER NOT NULL REFERENCES user(id),
  dicatat_at TEXT NOT NULL,
  bukti_media_id INTEGER REFERENCES media(id),
  setoran_id INTEGER REFERENCES setoran(id)
);

CREATE TABLE mutasi (
  id INTEGER PRIMARY KEY,
  pos_id TEXT NOT NULL REFERENCES pos_dana(id),
  tanggal TEXT NOT NULL,
  arah TEXT NOT NULL CHECK (arah IN ('masuk','keluar')),
  nominal INTEGER NOT NULL CHECK (nominal > 0),
  kategori TEXT NOT NULL,
  keterangan TEXT NOT NULL,                 -- versi internal (boleh berisi nama)
  keterangan_publik TEXT NOT NULL,          -- versi publik, tanpa data pribadi
  nota_media_id INTEGER REFERENCES media(id),
  ref_tipe TEXT,                            -- 'setoran','transfer_bukti','dansos','alokasi','koreksi','impor','saldo_awal'
  ref_id INTEGER,
  koreksi_dari INTEGER REFERENCES mutasi(id),
  dibuat_oleh INTEGER NOT NULL REFERENCES user(id),
  dibuat_at TEXT NOT NULL,
  CHECK (arah = 'masuk' OR nota_media_id IS NOT NULL OR ref_tipe IN ('alokasi','koreksi','impor'))
);

CREATE TABLE periode_tutup (
  periode TEXT PRIMARY KEY,
  ditutup_oleh INTEGER NOT NULL REFERENCES user(id),
  ditutup_at TEXT NOT NULL,
  snapshot_saldo TEXT NOT NULL,             -- JSON {pos_id: saldo}
  pdf_media_id INTEGER REFERENCES media(id)
);

-- kunci periode di level database
CREATE TRIGGER mutasi_lock_ins BEFORE INSERT ON mutasi
WHEN EXISTS (SELECT 1 FROM periode_tutup WHERE periode = substr(NEW.tanggal,1,7))
BEGIN SELECT RAISE(ABORT, 'periode sudah ditutup'); END;
CREATE TRIGGER mutasi_lock_upd BEFORE UPDATE ON mutasi
BEGIN SELECT RAISE(ABORT, 'mutasi tidak boleh diubah, buat koreksi'); END;
CREATE TRIGGER mutasi_lock_del BEFORE DELETE ON mutasi
BEGIN SELECT RAISE(ABORT, 'mutasi tidak boleh dihapus, buat koreksi'); END;
CREATE TRIGGER periode_tutup_lock BEFORE DELETE ON periode_tutup
BEGIN SELECT RAISE(ABORT, 'periode tertutup tidak bisa dibuka'); END;

CREATE TABLE pengajuan (
  id INTEGER PRIMARY KEY,
  jenis TEXT NOT NULL CHECK (jenis IN ('dansos','surat','tamu','hajatan','aduan','bukti_bayar')),
  rumah_id INTEGER NOT NULL REFERENCES rumah(id),
  diajukan_oleh INTEGER NOT NULL REFERENCES user(id),
  data TEXT NOT NULL,                       -- JSON sesuai jenis
  nominal INTEGER,                          -- dansos
  status TEXT NOT NULL CHECK (status IN ('diajukan','diproses','disetujui','ditolak','selesai')),
  ditangani_oleh INTEGER REFERENCES user(id),
  catatan_pengurus TEXT,
  dibuat_at TEXT NOT NULL, diubah_at TEXT NOT NULL
);

CREATE TABLE surat (
  id INTEGER PRIMARY KEY,
  pengajuan_id INTEGER NOT NULL UNIQUE REFERENCES pengajuan(id),
  nomor TEXT NOT NULL UNIQUE,               -- 012/RT006/RW028/X/2026
  kode_verifikasi TEXT NOT NULL UNIQUE,     -- acak 10 karakter, untuk /v/{kode}
  dibatalkan_at TEXT,
  pdf_media_id INTEGER REFERENCES media(id)
);

CREATE TABLE album (id INTEGER PRIMARY KEY, judul TEXT NOT NULL, slug TEXT NOT NULL UNIQUE, tanggal TEXT NOT NULL, cerita TEXT, sampul_media_id INTEGER REFERENCES media(id), dibuat_oleh INTEGER NOT NULL REFERENCES user(id));
CREATE TABLE album_foto (album_id INTEGER NOT NULL REFERENCES album(id), media_id INTEGER NOT NULL REFERENCES media(id), keterangan TEXT, urutan INTEGER NOT NULL DEFAULT 0, PRIMARY KEY (album_id, media_id));

CREATE TABLE jadwal_rutin (id INTEGER PRIMARY KEY, judul TEXT NOT NULL, aturan TEXT NOT NULL, gang INTEGER, aktif INTEGER NOT NULL DEFAULT 1);
  -- aturan contoh: 'WEEKLY;SAT;21:00' (ronda), 'EVERY=3M;GANG' (kerja bakti per gang), 'EVERY=4M' (fogging), 'MONTHLY;DAY=10' (batas tarik)
CREATE TABLE agenda (id INTEGER PRIMARY KEY, judul TEXT NOT NULL, keterangan TEXT, mulai TEXT NOT NULL, selesai TEXT, tingkat TEXT NOT NULL CHECK (tingkat IN ('info','penting','darurat')), gang INTEGER, rutin_id INTEGER REFERENCES jadwal_rutin(id));

CREATE TABLE pengumuman (id INTEGER PRIMARY KEY, judul TEXT NOT NULL, isi TEXT NOT NULL, tingkat TEXT NOT NULL, terbit_at TEXT NOT NULL, kedaluwarsa TEXT, kirim_wa INTEGER NOT NULL DEFAULT 0);

CREATE TABLE aturan_pasal (id TEXT PRIMARY KEY, bab TEXT NOT NULL, judul TEXT NOT NULL, isi_md TEXT NOT NULL, urutan INTEGER NOT NULL);

CREATE TABLE inventaris (id INTEGER PRIMARY KEY, nama TEXT NOT NULL, kategori TEXT, jumlah TEXT, lokasi TEXT, kondisi TEXT NOT NULL, catatan TEXT, bisa_dipinjam INTEGER NOT NULL DEFAULT 1);
CREATE TABLE peminjaman (id INTEGER PRIMARY KEY, inventaris_id INTEGER NOT NULL REFERENCES inventaris(id), pengajuan_id INTEGER REFERENCES pengajuan(id), mulai TEXT NOT NULL, selesai TEXT NOT NULL, status TEXT NOT NULL);

CREATE TABLE notif_outbox (id INTEGER PRIMARY KEY, no_wa TEXT NOT NULL, template TEXT NOT NULL, payload TEXT NOT NULL, status TEXT NOT NULL CHECK (status IN ('antri','terkirim','gagal')), percobaan INTEGER NOT NULL DEFAULT 0, kirim_setelah TEXT NOT NULL, error TEXT);
CREATE TABLE otp (no_wa TEXT NOT NULL, kode_hash TEXT NOT NULL, kedaluwarsa TEXT NOT NULL, percobaan INTEGER NOT NULL DEFAULT 0);
CREATE TABLE kode_cadangan (kode_hash TEXT PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES user(id), dibuat_oleh INTEGER NOT NULL REFERENCES user(id), kedaluwarsa TEXT NOT NULL, dipakai_at TEXT);
CREATE TABLE audit_log (id INTEGER PRIMARY KEY, user_id INTEGER, aksi TEXT NOT NULL, objek TEXT NOT NULL, objek_id TEXT, detail TEXT, ip TEXT, at TEXT NOT NULL);
```

Saldo pos selalu dihitung dari `mutasi` (SUM). Jangan menyimpan kolom saldo yang bisa tidak sinkron. Snapshot saldo hanya ada di `periode_tutup`.

Seed awal: `pos_dana` (6 baris), `setting` (bagian 3), `aturan_pasal` (isi AD/ART, teksnya ada di `docs/mockup.html` konstanta `ATURAN`), `inventaris` (panggung RT, sound system Politron + 2 mic wireless, terpal biru oranye 8 m, gerobak), `jadwal_rutin` (ronda malam Minggu per gang, kerja bakti per gang 3 bulanan, fogging 4 bulanan, rapat rutin 3 bulanan, batas tarik tanggal 10, batas setor tanggal 15).

## 7. Alur utama

**A. Iuran tunai (alur utama, ~90% transaksi)**
1. Tanggal 1 setiap bulan, job membuat `tagihan` jenis `kas` untuk semua rumah wajib iuran, terhubung ke penghuni aktif.
2. Koordinator membuka "Tarik iuran", melihat grid rumah gangnya per blok, mengetuk rumah yang membayar. Ketukan membuat `pembayaran` status `dipegang` (htmx, respons < 300 ms, swap satu kotak saja). Ketukan ulang membatalkan selama status masih `dipegang`.
3. Warga menerima WA otomatis: "Iuran Oktober 2026 Rp25.000 untuk G12/07 sudah dicatat oleh Pak Bambang."
4. Koordinator menekan "Setor ke bendahara" per periode. Sistem membuat `setoran` status `menunggu`, semua pembayaran terkait menjadi `disetor`. Bendahara mendapat WA.
5. Bendahara menekan "Terima" (atau "Tolak" dengan catatan bila jumlah uang tidak cocok, pembayaran kembali ke `dipegang`). Saat diterima: pembayaran menjadi `diterima` dan satu `mutasi` masuk ke `kas_rt` dibuat dengan `ref_tipe='setoran'`. Semua dalam satu transaksi DB.
6. Tanggal 11, warga yang belum bayar mendapat pengingat pribadi. Tanggal 16, koordinator yang belum setor mendapat pengingat.

**B. Transfer oleh warga**: warga mengunggah bukti, pembayaran `menunggu_verifikasi`. Bendahara verifikasi, langsung `diterima` dan mutasi masuk dibuat dengan `ref_tipe='transfer_bukti'` (tidak lewat koordinator).

**C. Pengeluaran**: bendahara memilih pos, mengisi keperluan, nominal, kategori, foto nota (kamera langsung, wajib), dan keterangan publik. Validasi: tanggal tidak di periode tertutup, saldo pos cukup (peringatan, bukan blokir).

**D. Tutup buku**: tombol tersedia untuk periode yang sudah lewat dan tidak ada setoran `menunggu`. Menghasilkan snapshot saldo, PDF laporan (saldo awal, pemasukan per gang, pengeluaran dengan nomor nota, saldo akhir, kolom tanda tangan ketua dan bendahara), lalu mengirim tautan PDF ke grup lewat WA. Setelah ditutup, koreksi = mutasi baru di periode berjalan dengan `koreksi_dari`.

**E. Dana sosial**: warga mengajukan, sistem menjalankan cek aturan (bagian 3) dan menampilkan hasilnya sebagai daftar centang sebelum tombol kirim aktif (lihat mockup, peran Warga, Ajukan, Dana sosial). Ketua menyetujui, bendahara mencairkan dengan bukti serah terima. Publik hanya melihat "N penyaluran, total Rp X".

**F. Surat pengantar**: warga memilih keperluan dan anggota keluarga. Ketua menyetujui, sistem membuat nomor surat berurutan per tahun dan PDF dengan kop RT, lalu mengirimkannya ke WA warga.

**G. Ganti penghuni**: pengurus menutup penghuni lama (isi `selesai`), membuat penghuni baru. Tagihan lama tetap terhubung ke penghuni lama. Rumah otomatis berstatus `kosong` bila tidak ada penghuni aktif.

**H. Posting kegiatan**: sekretaris mengisi judul, tanggal, album (baru atau existing), cerita, banyak foto sekaligus. Browser memperkecil foto ke sisi terpanjang 1920 px sebelum upload (canvas), server memperkecil ulang, membuat thumbnail 480 px, re-encode JPEG kualitas 82 (EXIF termasuk GPS hilang). Opsi kirim pengumuman WA ke grup.

**I. Login**: tanpa password. Warga memasukkan nomor WA yang sudah didaftarkan pengurus, menerima kode 6 digit (berlaku 5 menit, maksimal 5 percobaan). Sesi berlaku 180 hari. Cadangan bila gateway WA mati: koordinator atau pengurus bisa membuat kode sekali pakai (berlaku 24 jam) dari layar mereka untuk diberikan langsung. Tidak ada pendaftaran mandiri.

## 8. Daftar halaman

Publik: `/` beranda, `/kas`, `/kegiatan`, `/kegiatan/{slug}`, `/aturan` (+ anchor `#pasal-3.4`), `/pengurus`, `/agenda`, `/masuk`.

Warga: `/rumahku`, `/rumahku/bayar`, `/ajukan`, `/ajukan/{jenis}`, `/pengajuan/{id}`.

Koordinator: `/gang/{n}/tarik?periode=YYYY-MM`, `/gang/{n}/setor`.

Pengurus: `/kelola` (ringkasan), `/kelola/mutasi`, `/kelola/mutasi/baru`, `/kelola/setoran`, `/kelola/tunggakan`, `/kelola/tutup-buku`, `/kelola/rumah`, `/kelola/rumah/{alamat}`, `/kelola/pengajuan`, `/kelola/konten/posting`, `/kelola/agenda`, `/kelola/pengumuman`, `/kelola/inventaris`, `/kelola/user`, `/kelola/setting`, `/kelola/audit`, `/kelola/impor`.

RW dan perangkat desa: `/laporan` (rekap kependudukan agregat + daftar laporan kas tertutup + PDF), `/laporan/surat` (daftar surat terbit). Verifikasi publik: `/v/{kode}`.

Admin: `/admin` (status sistem: versi, ukuran DB, backup terakhir, status gateway WA, antrian outbox), `/admin/user`, `/admin/setting`, `/admin/impor`, `/admin/audit`.

Lain: `/m/{id}/{full|thumb}` media dengan pengecekan `akses` di handler (jangan pernah menyajikan folder upload sebagai static), `/healthz`.

Navigasi: bottom nav maksimal 5 item di mobile, left rail di layar >= 900 px. Item berbeda per peran, persis seperti mockup. User dengan banyak peran mendapat pengalih peran di header.

## 9. Design system — tema nature

Ambil token langsung dari `docs/mockup-nature.html` (blok `<style>` di `:root`). Jangan mengarang gaya baru.

Tema: alam dan kayu. Kertas krem, kayu jati, daun hijau, aksen emas. Dasarnya adalah papan pengumuman RT yang dibuat bagus, bukan dashboard korporat.

Permukaan:
- `.wood` = papan kayu, dipakai untuk top bar, rail, bottom nav, bingkai papan struktur, kartu saldo utama, dan bar setoran. Serat kayu dibuat dari `repeating-linear-gradient`, bukan gambar.
- `.wood.lite` = kayu lebih terang, khusus bingkai papan struktur.
- Kartu biasa memakai `--surface` (kertas) dengan garis `--line`.
- Latar halaman memakai dua `radial-gradient` lembut (hijau di kanan atas, kayu di kiri bawah) dengan `background-attachment: fixed`.

Catatan kontras yang sudah diperbaiki dan jangan diulang: teks krem di atas kayu terang gagal WCAG. Gradien `.wood` harus mulai dari `--wood-2`, bukan `--wood-1`. Elemen apa pun di dalam `.wood` yang berisi teks gelap wajib menyetel `color` sendiri (lihat `.plaque-in`).

Tipografi: `Fraunces` (serif, variable) untuk judul, angka besar, dan nama pada bagan. `Plus Jakarta Sans` untuk UI dan isi. Keduanya di-self-host sebagai woff2 di `static/fonts/`. Angka memakai `font-variant-numeric: tabular-nums`.

Warna (terang / gelap):

| Token | Terang | Gelap | Pemakaian |
|---|---|---|---|
| `--paper` | #F2ECDC | #131A13 | latar halaman, teks di atas tombol gelap |
| `--surface` | #FCF8EE | #1B241A | kartu |
| `--sunk` | #E9E0C9 | #232D20 | latar sekunder, kartu pos dana |
| `--ink` | #22301F | #EDE7D6 | teks utama |
| `--ink-2` / `--ink-3` | #56664F / #8A9480 | #AEBAA2 / #7C8A74 | teks sekunder |
| `--line` | #DCCFAE | #34412F | garis |
| `--leaf` / `--leaf-2` / `--leaf-soft` | #4C7A36 / #6D9B4E / #E2EDD3 | #8CC068 / #A6D083 / #1F2F1B | lunas, pemasukan, sukses, daun progres |
| `--moss` | #2C4A2A | #C7E0B0 | tombol primer |
| `--wood-1` / `--wood-2` / `--wood-3` | #A9794A / #7E5631 / #5E3E21 | #8A6240 / #5E4229 / #42301C | papan kayu |
| `--wood-ink` | #FDF3DE | #F6E7CC | teks di atas kayu |
| `--gold` | #C79A2E | #E0B44E | aksen papan nama (teks papan memakai #EFCE70) |
| `--clay` / `--clay-soft` | #A8562C / #F6E4D8 | #D98B5F / #33231A | pengeluaran, status "di koordinator" |
| `--sky` / `--sky-soft` | #3F6F82 / #DDEAEE | #7FB3C4 / #1B2A2F | info, bebas iuran |
| `--amber` / `--amber-soft` | #8A6413 / #F7ECD2 | #E2B65C / #302615 | peringatan, jatuh tempo |
| `--berry` / `--berry-soft` | #8E3B46 / #F7E3E4 | #E08A90 / #331E20 | tunggakan, error |

Body 15 px, minimum 12 px untuk meta, judul halaman 28 px weight 900 (Fraunces).

Bentuk: radius 8 / 14 / 22 px sesuai hierarki, tombol pill, tap target minimal 44 px.

Elemen khas (harus ada):
- **Papan kayu struktur pengurus** (`.plaque`): bingkai kayu, panel kertas di dalamnya, papan nama kayu dengan teks emas, bagan pohon dengan batang dan cabang dari gradien kayu. Kartu wakil, sekretaris, dan bendahara berbentuk daun (`border-radius: 6px 44px 6px 44px` dengan garis tulang daun diagonal). Avatar berupa lingkaran berbingkai kayu ganda. Di layar sempit cabang disembunyikan dan bagan menjadi daftar lipat (`<details>`), seperti panel kanan pada referensi.
- **Daun progres per gang** di beranda: satu daun = satu rumah wajib iuran, hijau penuh bila diterima bendahara, tanah liat bila masih di koordinator. Selalu diurutkan terisi lebih dulu sehingga tidak bisa dipetakan ke alamat.
- **Grid kotak rumah** di layar koordinator, dikelompokkan per blok, nomor rumah besar, nama KK kecil, lima status visual: belum, di koordinator, disetor, diterima, kosong (arsir), bebas (garis putus).
- **Bar lengket** di bawah grid koordinator berisi total uang di tangan dan tombol setor.

Larangan desain: tabel bulan x rumah untuk publik, ikon dekoratif tanpa fungsi, carousel otomatis, tekstur kayu berupa file gambar, mode gelap sebagai default (ikuti preferensi sistem). Huruf kapital semua hanya boleh untuk label mikro (`.role`, `.tag`, judul kotak gang) dan papan nama, tidak untuk kalimat.

Mode gelap adalah hutan malam, bukan abu-abu: kayu menjadi lebih pekat, daun menjadi hijau terang, kertas menjadi hijau kehitaman.

Copy: Bahasa Indonesia sederhana, kalimat aktif, sentence case. Tombol menyebut hasilnya ("Setor ke bendahara", "Simpan pengeluaran", "Terbitkan kegiatan"), dan toast memakai kata yang sama ("Setoran dikirim"). Pesan error menjelaskan apa yang salah dan cara memperbaikinya, tanpa minta maaf.

## 10. Privasi dan keamanan

- Publik tidak pernah menerima nama warga yang terhubung ke status bayar, nama penerima dansos, alasan medis, umur, gender, atau nomor WA. Uji ini secara otomatis (bagian 15).
- Jangan menyimpan NIK atau nomor KK di fase 1 sampai 3. Kalau nanti dibutuhkan untuk surat, enkripsi kolom dengan kunci dari env.
- Foto KTP tamu: akses `pengurus`, hapus otomatis 30 hari setelah tanggal pulang.
- Semua foto: EXIF dihapus lewat re-encode.
- Nota: akses `warga` (login). Bukti dansos: akses `pengurus` + `pemilik`.
- Cookie `Secure`, `HttpOnly`, `SameSite=Lax`. CSRF di semua POST. Rate limit OTP: 3 per 15 menit per nomor, 10 per jam per IP.
- Semua aksi tulis masuk `audit_log`.
- Header keamanan: CSP `default-src 'self'; img-src 'self' data: blob:; style-src 'self'; script-src 'self'; connect-src 'self'`.

## 11. Kemudahan pengguna

- Link dibagikan di grup WA, halaman pertama tanpa login harus sudah berguna (saldo, kelopak gang, agenda, foto). Sediakan Open Graph image dan judul agar pratinjau link di WA rapi.
- Login sekali, bertahan 180 hari. PWA bisa dipasang ke layar utama (manifest + ikon sakura + service worker yang men-cache shell dan halaman kas terakhir).
- Koordinator: seluruh penarikan satu gang selesai < 5 menit, tanpa mengetik.
- Bendahara: catat pengeluaran < 60 detik dari kamera HP.
- Notifikasi WA singkat dan personal. Template disimpan di kode (`internal/notif/templates.go`), contoh:
  - `iuran_dicatat`: "Iuran {bulan} Rp{nominal} untuk {alamat} sudah dicatat oleh {koordinator}. Terima kasih."
  - `pengingat_iuran`: "Iuran {bulan} untuk {alamat} belum tercatat. Bisa dibayar ke {koordinator} atau transfer lalu unggah bukti di {link}."
  - `setoran_masuk`: "Setoran gang {gang} {bulan} Rp{total} dari {koordinator} menunggu konfirmasi: {link}."
  - `laporan_bulanan`: "Laporan kas RT 06 {bulan} sudah ditutup. Saldo kas Rp{saldo}. Lihat: {link}."
- Pengingat pribadi tidak pernah dikirim ke grup.
- Teks minimal 15 px untuk isi, kontras WCAG AA, fokus keyboard terlihat, `prefers-reduced-motion` dihormati.

## 12. Deploy di VPS

Asumsi: Linux dengan Podman rootless, kode di `/devzone/sakuragcc`, data runtime di `/data/sakuragcc`, domain `{DOMAIN}`, Caddy terpasang di host.

```bash
# 1. sekali saja
sudo mkdir -p /data/sakuragcc/{db,media,backup,wa} && sudo chown -R $USER: /data/sakuragcc
loginctl enable-linger $USER
cp deploy/env.example deploy/.env   # isi SESSION_KEY, BASE_URL, WA_GATEWAY_URL, WA_USER, WA_PASS
chmod 600 deploy/.env

# 2. build
cd /devzone/sakuragcc
make test
podman build -t localhost/sakuragcc:latest -f deploy/Containerfile .

# 3. pasang quadlet
mkdir -p ~/.config/containers/systemd
cp deploy/sakuragcc.container deploy/wa-gateway.container ~/.config/containers/systemd/
systemctl --user daemon-reload
systemctl --user start wa-gateway sakuragcc

# 4. data awal
podman exec sakuragcc /app/sakuragcc import --rumah /seed/warga_seed.csv --dry-run   # tinjau laporan
podman exec sakuragcc /app/sakuragcc import --rumah /seed/warga_seed.csv
podman exec -it sakuragcc /app/sakuragcc createadmin --nama "Admin" --wa +62xxxxxxxxxx --peran admin
# lalu dari /admin/user: buat akun ketua, wakil, sekretaris, bendahara, koordinator. akun warga dibuat massal dari impor + nomor WA

# 5. pairing WhatsApp: buka UI gateway lewat SSH tunnel, scan QR dengan HP nomor khusus RT
ssh -L 3000:127.0.0.1:3000 user@vps     # lalu buka http://localhost:3000

# 6. Caddy: tambahkan blok dari deploy/Caddyfile ke /etc/caddy/Caddyfile
sudo systemctl reload caddy

# 7. backup offsite
cp deploy/backup.service deploy/backup.timer ~/.config/systemd/user/
systemctl --user daemon-reload && systemctl --user enable --now backup.timer
```

`deploy/Containerfile`:
```dockerfile
FROM docker.io/library/golang:1.23-alpine AS build
WORKDIR /src
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sakuragcc ./cmd/sakuragcc

FROM docker.io/library/alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app && mkdir -p /data && chown app /data
ENV TZ=Asia/Jakarta
COPY --from=build /out/sakuragcc /app/sakuragcc
COPY seed /seed
USER app
EXPOSE 8080
ENTRYPOINT ["/app/sakuragcc"]
CMD ["serve"]
```

`deploy/sakuragcc.container`:
```ini
[Unit]
Description=SakuraGCC portal warga
After=wa-gateway.service

[Container]
Image=localhost/sakuragcc:latest
ContainerName=sakuragcc
PublishPort=127.0.0.1:8080:8080
Volume=/data/sakuragcc/db:/data/db:Z
Volume=/data/sakuragcc/media:/data/media:Z
Volume=/data/sakuragcc/backup:/data/backup:Z
EnvironmentFile=/devzone/sakuragcc/deploy/.env
UserNS=keep-id:uid=10001,gid=10001
HealthCmd=wget -qO- http://127.0.0.1:8080/healthz || exit 1

[Service]
Restart=always

[Install]
WantedBy=default.target
```

`deploy/wa-gateway.container`:
```ini
[Container]
Image=docker.io/aldinokemal2104/go-whatsapp-web-multidevice:latest
ContainerName=wa-gateway
PublishPort=127.0.0.1:3000:3000
Volume=/data/sakuragcc/wa:/app/storages:Z
EnvironmentFile=/devzone/sakuragcc/deploy/.env
Exec=rest

[Service]
Restart=always

[Install]
WantedBy=default.target
```
Cek README image gateway untuk nama variabel basic auth yang benar pada versi yang dipakai, lalu set di `.env`.

`deploy/Caddyfile`:
```
{DOMAIN} {
    encode zstd gzip
    header {
        Strict-Transport-Security "max-age=31536000"
        X-Content-Type-Options nosniff
        Referrer-Policy strict-origin-when-cross-origin
    }
    reverse_proxy 127.0.0.1:8080
}
```

Kalau app berjalan di container dan menghubungi gateway di container lain, pakai `WA_GATEWAY_URL=http://host.containers.internal:3000`, atau satukan keduanya dalam satu pod.

Backup: aplikasi menjalankan `VACUUM INTO '/data/backup/app-YYYY-MM-DD.db'` setiap pukul 02.00 dan menyimpan 14 hari terakhir. `backup.sh` menjalankan `restic backup /data/sakuragcc/backup /data/sakuragcc/media` ke repo kedua (disk lain atau rclone remote), lalu `restic forget --keep-daily 14 --keep-monthly 12 --prune`. Uji pemulihan sekali sebelum go-live.

Update aplikasi: `git pull && make test && podman build -t localhost/sakuragcc:latest -f deploy/Containerfile . && systemctl --user restart sakuragcc`. Migrasi berjalan otomatis saat start dan hanya maju (tidak ada migrasi turun di produksi).

Risiko WA: gateway tidak resmi dan nomor bisa diblokir. Pakai nomor khusus RT, jangan nomor pribadi. Batasi kirim (satu pesan per 3 sampai 6 detik, acak), jangan blast massal ke ratusan nomor sekaligus, dan semua fitur inti harus tetap jalan bila gateway mati (outbox mengantri, login pakai kode cadangan).

## 13. Migrasi data

- `seed/warga_seed.csv` sudah dibersihkan: alamat kanonik, status lowercase, nama dari font unicode dinormalisasi, nama di-title-case, kolom `catatan_validasi` berisi masalah yang perlu dicek pengurus (19 baris).
- Perintah `import` wajib punya `--dry-run` yang mencetak: jumlah rumah per gang per status, baris dengan catatan validasi, duplikasi KK lintas rumah.
- Umur di CSV dikonversi menjadi `tahun_lahir = 2026 - umur` dengan `tahun_lahir_perkiraan = 1`.
- Rumah berstatus tetap/kontrak membuat satu `penghuni` aktif dengan `mulai = 2026-04-01`. Rumah kontrak tanpa nama (G10/14) membuat penghuni dengan nama "Penghuni G10/14" dan catatan validasi.
- Riwayat iuran April sampai September 2026 dari web lama diimpor dari CSV terpisah (`seed/iuran_2026.csv`, kolom `alamat,periode,nominal`) yang disiapkan bendahara. Diimpor sebagai pembayaran metode `impor` status `diterima` dengan satu mutasi masuk per gang per bulan, `ref_tipe='impor'`. Periode yang sudah dikonfirmasi bendahara langsung ditutup buku dengan catatan "saldo migrasi".
- Saldo awal kas sebelum April 2026 dicatat sebagai satu mutasi masuk `ref_tipe='saldo_awal'` setelah dikonfirmasi bendahara.
- Dana sosial yang sudah tersalur (2 baris di web lama: 11 April 2026 Rp200.000, 19 April 2026 Rp400.000) diimpor sebagai mutasi keluar dari `dana_sosial` dengan keterangan publik generik ("Santunan warga").

## 14. Fase dan kriteria selesai

**Fase 1 (MVP, target 2 sampai 3 minggu)**: login OTP + kode cadangan, peran, rumah/penghuni + impor, beranda publik, kas publik, tarik iuran koordinator, setoran + konfirmasi, mutasi + pengeluaran bernota, rumahku (status iuran), AD/ART, struktur, galeri + posting, PWA dasar, backup.

Selesai bila:
- Koordinator bisa menandai 40 rumah dan menyetor di HP 360 px tanpa scroll horizontal.
- Setelah bendahara menerima setoran, saldo publik dan kelopak gang berubah tanpa langkah lain.
- Mutasi di periode tertutup tidak bisa ditambah, diubah, atau dihapus, termasuk lewat SQL langsung (trigger).
- Test otomatis memastikan respons publik (`/`, `/kas`, `/kegiatan`, `/agenda`) tidak mengandung satu pun nama KK dari tabel penghuni.
- Lighthouse mobile: performance >= 90, accessibility >= 95. Beranda < 150 KB tanpa foto.

**Fase 2**: tagihan otomatis bulanan, pengingat WA, tunggakan + blokir layanan, transfer + verifikasi bukti, tutup buku + PDF, dana sosial dengan cek aturan, surat pengantar PDF bernomor.

**Fase 3**: lapor tamu, izin hajatan + peminjaman inventaris, aduan lingkungan dengan status, agenda rutin otomatis (ronda, kerja bakti, fogging, rapat), pengumuman + WA, polling satu suara per rumah, antrian offline untuk ketukan koordinator di service worker.

## 15. Testing

- Unit test table-driven untuk: hitung tunggakan (termasuk batas tanggal 10/11), kelayakan dansos, state machine pembayaran (`dipegang -> disetor -> diterima`, `disetor -> dipegang` hanya lewat tolak setoran), penomoran surat, nominal integer.
- Test integrasi dengan SQLite in-memory untuk trigger kunci periode dan alur setoran lengkap dalam satu transaksi.
- Test privasi publik (lihat kriteria fase 1).
- Test pemisahan tugas: admin mendapat 403 di semua endpoint `/kelola/mutasi*`, `/kelola/setoran*`, `/kelola/tutup-buku`; `rw` dan `perangkat_desa` mendapat 403 di semua POST dan tidak menerima nama KK di `/laporan`.
- Test `/v/{kode}` hanya menampilkan nama tersensor.
- Test handler dengan `httptest` untuk setiap middleware peran: akses yang tidak berhak harus 403 (atau redirect ke `/masuk` untuk tamu), bukan 200 dengan halaman kosong. Koordinator gang 1 tidak boleh bisa menandai rumah gang 2.
- `make test` menjalankan `go vet`, `staticcheck`, dan `go test ./...`.
