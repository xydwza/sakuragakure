-- name: MediaByID :one
SELECT id, path, thumb_path, mime, ukuran, akses, pemilik_user_id, hapus_setelah, dibuat_at
FROM media
WHERE id = ?;
