-- name: ListAturan :many
SELECT id, bab, judul, isi_md, urutan FROM aturan_pasal ORDER BY urutan;

-- name: ListPengurus :many
SELECT u.id, u.nama, u.no_wa, u.alamat, u.detail, u.foto_media_id, up.peran, up.gang
FROM user_peran up
JOIN user u ON u.id = up.user_id
WHERE up.peran IN ('ketua','wakil','sekretaris','bendahara','koordinator','pembantu_koordinator')
ORDER BY up.urutan, u.id;

-- name: RumahUser :one
SELECT r.id, r.alamat, r.blok, r.gang, r.status, ph.nama_kk, ph.jumlah_anggota, ph.status_huni
FROM user u
JOIN penghuni ph ON ph.id = u.penghuni_id
JOIN rumah r ON r.id = ph.rumah_id
WHERE u.id = ?;

-- name: IuranRumah :many
SELECT t.periode, COALESCE(p.status, '') AS bayar_status, t.nominal
FROM tagihan t
LEFT JOIN pembayaran p ON p.tagihan_id = t.id
WHERE t.rumah_id = ? AND t.jenis = 'kas'
ORDER BY t.periode;

-- name: ListAlbum :many
SELECT id, judul, slug, tanggal, cerita, sampul_media_id FROM album ORDER BY tanggal DESC;

-- name: AlbumBySlug :one
SELECT id, judul, slug, tanggal, cerita, sampul_media_id FROM album WHERE slug = ?;

-- name: FotoAlbum :many
SELECT af.media_id, af.keterangan, af.urutan FROM album_foto af WHERE af.album_id = ? ORDER BY af.urutan, af.media_id;

-- name: AgendaMendatang :many
SELECT judul, keterangan, mulai, tingkat FROM agenda WHERE mulai >= ? ORDER BY mulai LIMIT ?;
