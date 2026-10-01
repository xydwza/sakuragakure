-- +goose Up
ALTER TABLE user ADD COLUMN alamat TEXT;
ALTER TABLE user ADD COLUMN detail TEXT;
ALTER TABLE user ADD COLUMN foto_media_id INTEGER REFERENCES media(id);
