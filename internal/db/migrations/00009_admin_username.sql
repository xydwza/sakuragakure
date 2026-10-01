-- +goose Up
-- Akun admin yang sudah ada (dibuat via createadmin) diberi username + sandi default.
UPDATE user SET username = 'admin', password_hash = '$2a$10$sTNF5Wtg3aCIvl/RXE1gA.eVjtddZyg4UD1uQzZrKk1tlR54hRToK' WHERE no_wa = '+6289800000000';
