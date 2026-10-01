-- +goose Up
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
  CHECK (arah = 'masuk' OR nota_media_id IS NOT NULL OR COALESCE(ref_tipe,'') IN ('alokasi','koreksi','impor'))
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
