-- name: GridGang :many
SELECT r.id, r.alamat, r.blok, r.nomor, r.status, r.bebas_iuran_alasan,
       COALESCE(ph.nama_kk, '') AS nama_kk,
       COALESCE(p.status, '') AS bayar_status
FROM rumah r
LEFT JOIN penghuni ph ON ph.rumah_id = r.id AND ph.selesai IS NULL
LEFT JOIN tagihan t ON t.rumah_id = r.id AND t.jenis = 'kas' AND t.periode = ?
LEFT JOIN pembayaran p ON p.tagihan_id = t.id
WHERE r.gang = ?
ORDER BY r.blok, r.nomor;

-- name: HeldPerPeriode :many
SELECT t.periode,
       CAST(COUNT(*) AS INTEGER) AS jumlah,
       CAST(SUM(t.nominal) AS INTEGER) AS total
FROM pembayaran p
JOIN tagihan t ON t.id = p.tagihan_id
JOIN rumah r ON r.id = t.rumah_id
WHERE p.status = 'dipegang' AND r.gang = ?
GROUP BY t.periode
ORDER BY t.periode DESC;

-- name: SetoranRiwayat :many
SELECT s.id, s.periode, s.total, s.status, s.dibuat_at, s.catatan
FROM setoran s
WHERE s.gang = ?
ORDER BY s.dibuat_at DESC;

-- name: SetoranMenunggu :many
SELECT s.id, s.gang, s.periode, s.total, s.koordinator_id, u.nama AS koordinator, s.dibuat_at
FROM setoran s
JOIN user u ON u.id = s.koordinator_id
WHERE s.status = 'menunggu'
ORDER BY s.periode, s.gang;
