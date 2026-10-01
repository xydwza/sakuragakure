-- name: ListUser :many
SELECT u.id, u.nama, u.no_wa, u.alamat, u.detail, u.aktif, u.foto_media_id,
       (SELECT group_concat(up.peran, ',') FROM user_peran up WHERE up.user_id = u.id) AS peran
FROM user u
ORDER BY u.id;

-- name: UserByID :one
SELECT id, nama, no_wa, alamat, detail, aktif, foto_media_id FROM user WHERE id = ?;

-- name: UpdateUser :exec
UPDATE user SET nama = ?, no_wa = ?, alamat = ?, detail = ? WHERE id = ?;

-- name: UpdateUserFoto :exec
UPDATE user SET foto_media_id = ? WHERE id = ?;
