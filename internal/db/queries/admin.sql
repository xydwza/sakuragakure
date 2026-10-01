-- name: ListAudit :many
SELECT a.id, a.user_id, COALESCE(u.nama, '') AS user_nama, a.aksi, a.objek, a.objek_id, a.detail, a.ip, a.at
FROM audit_log a
LEFT JOIN user u ON u.id = a.user_id
ORDER BY a.id DESC
LIMIT 200;

-- name: UserPeranList :many
SELECT peran, gang FROM user_peran WHERE user_id = ? ORDER BY urutan;

-- name: AddUserPeran :exec
INSERT OR IGNORE INTO user_peran (user_id, peran, gang, urutan) VALUES (?, ?, ?, ?);

-- name: RemoveUserPeran :exec
DELETE FROM user_peran WHERE user_id = ? AND peran = ? AND gang = ?;

-- name: ToggleUserAktif :exec
UPDATE user SET aktif = ? WHERE id = ?;

-- name: ListMedia :many
SELECT id, path, thumb_path, mime, ukuran, akses, pemilik_user_id, dibuat_at FROM media ORDER BY id DESC LIMIT 200;

-- name: DeleteMedia :exec
DELETE FROM media WHERE id = ?;

-- name: DeleteAlbum :exec
DELETE FROM album WHERE id = ?;

-- name: DeleteAlbumFoto :exec
DELETE FROM album_foto WHERE album_id = ? AND media_id = ?;

-- name: ListFotoAlbum :many
SELECT media_id FROM album_foto WHERE album_id = ?;
