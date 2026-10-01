-- +goose Up
ALTER TABLE user ADD COLUMN password_hash TEXT;
