-- +goose Up
-- Data historis dari web lama (OCR rtjepri.netlify.app, 1 Okt 2026).
-- Sumber: data/OCR_data_sakuragakure.md bagian 6, 8, dan 10.
-- Kas RT: iuran per bulan (total). Dana sosial: 2 penyaluran. Agenda: 3 lama.
-- Dicatat sebagai mutasi ref_tipe='impor'/'alokasi' oleh user "Sistem".

INSERT INTO user (nama, no_wa, aktif, created_at) VALUES ('Sistem', 'migrasi', 0, '2026-10-01T00:00:00+07:00');

-- iuran masuk kas_rt (total per bulan dari web lama)
INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, ref_tipe, dibuat_oleh, dibuat_at) VALUES
  ('kas_rt', '2026-04-15', 'masuk', 2500000, 'Iuran', 'Iuran April 2026 (migrasi)', 'Iuran April 2026 (migrasi)', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00'),
  ('kas_rt', '2026-05-15', 'masuk', 2600000, 'Iuran', 'Iuran Mei 2026 (migrasi)', 'Iuran Mei 2026 (migrasi)', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00'),
  ('kas_rt', '2026-06-15', 'masuk', 2150000, 'Iuran', 'Iuran Juni 2026 (migrasi)', 'Iuran Juni 2026 (migrasi)', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00'),
  ('kas_rt', '2026-07-15', 'masuk', 825000,  'Iuran', 'Iuran Juli 2026 (migrasi)', 'Iuran Juli 2026 (migrasi)', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00'),
  ('kas_rt', '2026-08-15', 'masuk', 275000,  'Iuran', 'Iuran Agustus 2026 (migrasi)', 'Iuran Agustus 2026 (migrasi)', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00');

-- alokasi kas_rt -> dana_sosial (dana sosial diambil dari kas RT, pasal 3.4)
INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, ref_tipe, dibuat_oleh, dibuat_at) VALUES
  ('kas_rt', '2026-04-01', 'keluar', 600000, 'Transfer', 'Alokasi ke dana sosial (migrasi)', 'Alokasi ke dana sosial (migrasi)', 'alokasi', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00'),
  ('dana_sosial', '2026-04-01', 'masuk', 600000, 'Transfer', 'Alokasi dari kas RT (migrasi)', 'Alokasi dari kas RT (migrasi)', 'alokasi', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00');

-- penyaluran dana sosial (2 baris dari web lama)
INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, ref_tipe, dibuat_oleh, dibuat_at) VALUES
  ('dana_sosial', '2026-04-11', 'keluar', 200000, 'Dana sosial', 'Santunan warga', 'Santunan warga', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00'),
  ('dana_sosial', '2026-04-19', 'keluar', 400000, 'Dana sosial', 'Santunan warga', 'Santunan warga', 'impor', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-01T00:00:00+07:00');

-- agenda lama dari web lama (semua sudah lewat per 1 Okt 2026)
INSERT INTO agenda (judul, keterangan, mulai, selesai, tingkat) VALUES
  ('Persiapan Acara Kemerdekaan', 'Acara lomba', '2026-08-16T07:00:00+07:00', '2026-08-17T16:00:00+07:00', 'info'),
  ('Pertemuan perdana remaja RT 06', 'Rapat pembentukan organisasi remaja lingkungan RT 06', '2026-06-20T20:00:00+07:00', '2026-06-21T10:00:00+07:00', 'info'),
  ('Kegiatan 17 an', 'Lomba voli', '2026-08-08T16:00:00+07:00', '2026-08-09T11:00:00+07:00', 'info');
