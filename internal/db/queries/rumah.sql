-- name: ListRumah :many
SELECT r.alamat, r.blok, r.nomor, r.gang, r.status, r.bebas_iuran_alasan,
       COALESCE(ph.nama_kk, '') AS nama_kk, ph.jumlah_anggota, r.catatan_validasi
FROM rumah r
LEFT JOIN penghuni ph ON ph.rumah_id = r.id AND ph.selesai IS NULL
ORDER BY r.gang, r.blok, r.nomor;

-- name: ListSetoranSemua :many
SELECT s.gang, s.periode, s.total, s.status, s.dibuat_at, COALESCE(u.nama, '') AS koordinator
FROM setoran s
LEFT JOIN user u ON u.id = s.koordinator_id
ORDER BY s.dibuat_at DESC;

-- name: RekapKependudukan :many
SELECT r.gang,
       COUNT(*) AS total_rumah,
       CAST(SUM(CASE WHEN r.status = 'tetap' THEN 1 ELSE 0 END) AS INTEGER) AS tetap,
       CAST(SUM(CASE WHEN r.status = 'kontrak' THEN 1 ELSE 0 END) AS INTEGER) AS kontrak,
       CAST(SUM(CASE WHEN r.status = 'kosong' THEN 1 ELSE 0 END) AS INTEGER) AS kosong,
       COUNT(ph.id) AS jumlah_kk,
       CAST(COALESCE(SUM(ph.jumlah_anggota), 0) AS INTEGER) AS jumlah_jiwa
FROM rumah r
LEFT JOIN penghuni ph ON ph.rumah_id = r.id AND ph.selesai IS NULL
GROUP BY r.gang
ORDER BY r.gang;
