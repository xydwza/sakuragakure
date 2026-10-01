-- +goose Up
CREATE TABLE undangan (
  id INTEGER PRIMARY KEY,
  judul TEXT NOT NULL,
  isi TEXT NOT NULL,
  terbit_at TEXT NOT NULL,
  dibuat_oleh INTEGER NOT NULL REFERENCES user(id)
);
CREATE TABLE rsvp (
  undangan_id INTEGER NOT NULL REFERENCES undangan(id),
  rumah_id INTEGER NOT NULL REFERENCES rumah(id),
  jawaban TEXT NOT NULL CHECK (jawaban IN ('hadir','tidak')),
  dijawab_at TEXT NOT NULL,
  PRIMARY KEY (undangan_id, rumah_id)
);
