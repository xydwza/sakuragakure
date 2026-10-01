#!/bin/sh
# Offsite: salin cadangan + media ke repo kedua (disk lain / rclone remote).
# Uji pemulihan sekali sebelum go-live.
set -e
restic -r "$RESTIC_REPO" backup /data/sakuragakure/backup /data/sakuragakure/media
restic -r "$RESTIC_REPO" forget --keep-daily 14 --keep-monthly 12 --prune
