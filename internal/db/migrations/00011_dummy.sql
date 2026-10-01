-- +goose Up
-- Data dummy untuk visual (administrator mengganti dengan data riil).

-- 1) Tagihan kas Oktober 2026 + pembayaran (kelopak beranda tampil terisi)
INSERT INTO tagihan (rumah_id, penghuni_id, jenis, periode, nominal)
SELECT r.id, ph.id, 'kas', '2026-10', 25000
FROM rumah r LEFT JOIN penghuni ph ON ph.rumah_id = r.id AND ph.selesai IS NULL
WHERE r.status != 'kosong' AND r.bebas_iuran_alasan IS NULL
ON CONFLICT(rumah_id, jenis, periode) DO NOTHING;

INSERT INTO pembayaran (tagihan_id, metode, status, dicatat_oleh, dicatat_at)
SELECT t.id, 'tunai',
       CASE WHEN r.id % 10 < 7 THEN 'diterima' WHEN r.id % 10 = 7 THEN 'dipegang' ELSE 'disetor' END,
       (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-05T00:00:00+07:00'
FROM tagihan t JOIN rumah r ON r.id = t.rumah_id
WHERE t.jenis = 'kas' AND t.periode = '2026-10' AND r.id % 10 < 9
ON CONFLICT(tagihan_id) DO NOTHING;

-- 2) Agenda mendatang
INSERT INTO agenda (judul, keterangan, mulai, selesai, tingkat, gang) VALUES
 ('Ronda malam Minggu', 'Tiap gang, jadwal di grup masing-masing', '2026-10-03T21:00:00+07:00', NULL, 'info', NULL),
 ('Batas penarikan iuran Oktober', 'Koordinator gang menarik tanggal 1 sampai 10', '2026-10-10T00:00:00+07:00', NULL, 'penting', NULL),
 ('Kerja bakti gang 1', 'Bersih selokan dan pekarangan', '2026-10-11T07:00:00+07:00', NULL, 'info', 1),
 ('Batas setoran ke bendahara', 'Koordinator gang setor paling lambat tanggal 15', '2026-10-15T00:00:00+07:00', NULL, 'penting', NULL),
 ('Fogging lingkungan', 'Tutup tempat air', '2026-10-18T08:00:00+07:00', NULL, 'info', NULL),
 ('Rapat rutin warga triwulan', 'Pos RT, kehadiran kepala keluarga diharapkan', '2026-10-25T20:00:00+07:00', NULL, 'penting', NULL);

-- 3) Undangan + RSVP
INSERT INTO undangan (judul, isi, terbit_at, dibuat_oleh)
VALUES ('Rapat Rutin Warga Triwulan IV', 'Mengundang Bapak/Ibu kepala keluarga RT 006 untuk hadir dalam rapat rutin triwulan. Agenda: laporan kas Juli sampai September, rencana kegiatan akhir tahun, dan pembahasan iuran rukun kematian.',
        '2026-09-30T21:10:00+07:00', (SELECT id FROM user WHERE no_wa = 'migrasi'));

INSERT INTO rsvp (undangan_id, rumah_id, jawaban, dijawab_at)
SELECT 1, r.id, CASE WHEN r.id % 4 = 0 THEN 'tidak' ELSE 'hadir' END, '2026-10-01T10:00:00+07:00'
FROM rumah r
WHERE r.status != 'kosong' AND r.bebas_iuran_alasan IS NULL AND r.id % 2 = 0;

-- 4) Rukem: tagihan + sebagian sudah bayar
INSERT INTO tagihan (rumah_id, penghuni_id, jenis, periode, nominal)
SELECT r.id, ph.id, 'rukem', '2026', 50000
FROM rumah r LEFT JOIN penghuni ph ON ph.rumah_id = r.id AND ph.selesai IS NULL
WHERE r.status != 'kosong' AND r.bebas_iuran_alasan IS NULL
ON CONFLICT(rumah_id, jenis, periode) DO NOTHING;

INSERT INTO pembayaran (tagihan_id, metode, status, dicatat_oleh, dicatat_at)
SELECT t.id, 'tunai', 'diterima', (SELECT id FROM user WHERE no_wa = 'migrasi'), '2026-10-05T00:00:00+07:00'
FROM tagihan t JOIN rumah r ON r.id = t.rumah_id
WHERE t.jenis = 'rukem' AND r.id % 3 = 0
ON CONFLICT(tagihan_id) DO NOTHING;

-- 5) Galeri: media + album
INSERT INTO media (path, thumb_path, mime, ukuran, akses, dibuat_at) VALUES
 ('/seed/demo/17-agustus.jpg', NULL, 'image/jpeg', 12195, 'publik', '2026-10-01T00:00:00+07:00'),
 ('/seed/demo/kerja-bakti.jpg', NULL, 'image/jpeg', 11644, 'publik', '2026-10-01T00:00:00+07:00'),
 ('/seed/demo/rapat-warga.jpg', NULL, 'image/jpeg', 10867, 'publik', '2026-10-01T00:00:00+07:00'),
 ('/seed/demo/ronda.jpg', NULL, 'image/jpeg', 11763, 'publik', '2026-10-01T00:00:00+07:00');

INSERT INTO album (judul, slug, tanggal, cerita, sampul_media_id, dibuat_oleh) VALUES
 ('17 Agustus 2026', '17-agustus-2026', '2026-08-22', 'Karnaval, voli putri, dan acara puncak Dirgahayu RI.', (SELECT id FROM media WHERE path = '/seed/demo/17-agustus.jpg'), (SELECT id FROM user WHERE no_wa = 'migrasi')),
 ('Kerja bakti gang', 'kerja-bakti-gang', '2026-08-08', 'Bersih selokan dan pekarangan bersama warga.', (SELECT id FROM media WHERE path = '/seed/demo/kerja-bakti.jpg'), (SELECT id FROM user WHERE no_wa = 'migrasi')),
 ('Rapat warga triwulan', 'rapat-warga-triwulan', '2026-07-26', 'Laporan kas dan rencana kegiatan.', (SELECT id FROM media WHERE path = '/seed/demo/rapat-warga.jpg'), (SELECT id FROM user WHERE no_wa = 'migrasi'));

INSERT INTO album_foto (album_id, media_id, keterangan, urutan)
SELECT a.id, m.id, '', 0 FROM album a, media m
WHERE a.slug = '17-agustus-2026' AND m.path = '/seed/demo/17-agustus.jpg';
INSERT INTO album_foto (album_id, media_id, keterangan, urutan)
SELECT a.id, m.id, '', 0 FROM album a, media m
WHERE a.slug = 'kerja-bakti-gang' AND m.path = '/seed/demo/kerja-bakti.jpg';
INSERT INTO album_foto (album_id, media_id, keterangan, urutan)
SELECT a.id, m.id, '', 0 FROM album a, media m
WHERE a.slug = 'rapat-warga-triwulan' AND m.path = '/seed/demo/rapat-warga.jpg';
