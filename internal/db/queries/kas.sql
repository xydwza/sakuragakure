-- name: SaldoPosis :many
SELECT pos_id,
       CAST(COALESCE(SUM(CASE WHEN arah = 'masuk' THEN nominal ELSE -nominal END), 0) AS INTEGER) AS saldo
FROM mutasi
GROUP BY pos_id;

-- name: SaldoPos :one
SELECT CAST(COALESCE(SUM(CASE WHEN arah = 'masuk' THEN nominal ELSE -nominal END), 0) AS INTEGER) AS saldo
FROM mutasi
WHERE pos_id = ?;

-- name: MutasiPublik :many
SELECT m.id, m.tanggal, m.arah, m.nominal, m.kategori, m.keterangan_publik, p.nama AS pos_nama
FROM mutasi m
JOIN pos_dana p ON p.id = m.pos_id
ORDER BY m.tanggal DESC, m.id DESC
LIMIT ?;

-- name: MutasiSemua :many
SELECT m.id, m.tanggal, m.arah, m.nominal, m.kategori, m.keterangan, m.keterangan_publik,
       p.nama AS pos_nama, m.nota_media_id
FROM mutasi m
JOIN pos_dana p ON p.id = m.pos_id
ORDER BY m.tanggal DESC, m.id DESC
LIMIT ?;

-- name: MutasiLaporan :many
SELECT m.tanggal, p.nama AS pos, m.arah, m.nominal, m.kategori, m.keterangan_publik
FROM mutasi m
JOIN pos_dana p ON p.id = m.pos_id
ORDER BY m.tanggal, m.id;

-- name: WajibIuranPerGang :many
SELECT gang, COUNT(*) AS total
FROM rumah
WHERE status != 'kosong' AND bebas_iuran_alasan IS NULL
GROUP BY gang
ORDER BY gang;

-- name: KelopakGang :many
SELECT r.gang,
       CAST(COALESCE(SUM(CASE WHEN p.status = 'diterima' THEN 1 ELSE 0 END), 0) AS INTEGER) AS diterima,
       CAST(COALESCE(SUM(CASE WHEN p.status IN ('dipegang','disetor','menunggu_verifikasi') THEN 1 ELSE 0 END), 0) AS INTEGER) AS dipegang
FROM rumah r
LEFT JOIN tagihan t ON t.rumah_id = r.id AND t.jenis = 'kas' AND t.periode = ?
LEFT JOIN pembayaran p ON p.tagihan_id = t.id
WHERE r.status != 'kosong' AND r.bebas_iuran_alasan IS NULL
GROUP BY r.gang
ORDER BY r.gang;

-- name: GrafikKasBulanan :many
SELECT substr(tanggal, 1, 7) AS bulan,
       CAST(COALESCE(SUM(CASE WHEN arah = 'masuk' THEN nominal ELSE 0 END), 0) AS INTEGER) AS masuk,
       CAST(COALESCE(SUM(CASE WHEN arah = 'keluar' THEN nominal ELSE 0 END), 0) AS INTEGER) AS keluar
FROM mutasi
WHERE pos_id = 'kas_rt' AND COALESCE(ref_tipe, '') != 'alokasi'
GROUP BY bulan
ORDER BY bulan;

-- name: DansosTahun :one
SELECT COUNT(*) AS jumlah,
       CAST(COALESCE(SUM(nominal), 0) AS INTEGER) AS total
FROM mutasi
WHERE pos_id = 'dana_sosial' AND arah = 'keluar' AND substr(tanggal, 1, 4) = ?;

-- name: NominalDiKoordinator :one
SELECT CAST(COALESCE(SUM(t.nominal), 0) AS INTEGER) AS total
FROM pembayaran p
JOIN tagihan t ON t.id = p.tagihan_id
WHERE p.status IN ('dipegang','disetor','menunggu_verifikasi');

-- name: PosDanaList :many
SELECT id, nama, publik FROM pos_dana ORDER BY id;
