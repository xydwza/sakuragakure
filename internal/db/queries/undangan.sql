-- name: ListUndangan :many
SELECT id, judul, isi, terbit_at, dibuat_oleh FROM undangan ORDER BY id DESC;

-- name: UndanganByID :one
SELECT id, judul, isi, terbit_at, dibuat_oleh FROM undangan WHERE id = ?;

-- name: BuatUndangan :one
INSERT INTO undangan (judul, isi, terbit_at, dibuat_oleh) VALUES (?, ?, ?, ?) RETURNING id;

-- name: RsvpJawaban :one
SELECT jawaban FROM rsvp WHERE undangan_id = ? AND rumah_id = ?;

-- name: SetRsvp :exec
INSERT INTO rsvp (undangan_id, rumah_id, jawaban, dijawab_at) VALUES (?, ?, ?, ?)
ON CONFLICT(undangan_id, rumah_id) DO UPDATE SET jawaban = excluded.jawaban, dijawab_at = excluded.dijawab_at;

-- name: RsvpCount :one
SELECT
  CAST(SUM(CASE WHEN jawaban = 'hadir' THEN 1 ELSE 0 END) AS INTEGER) AS hadir,
  CAST(SUM(CASE WHEN jawaban = 'tidak' THEN 1 ELSE 0 END) AS INTEGER) AS tidak
FROM rsvp WHERE undangan_id = ?;

-- name: WajibIuranRumahCount :one
SELECT COUNT(*) AS total FROM rumah WHERE status != 'kosong' AND bebas_iuran_alasan IS NULL;
