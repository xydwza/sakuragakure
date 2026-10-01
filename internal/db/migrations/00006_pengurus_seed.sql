-- +goose Up
-- Pengurus RT (struktur dari AD/ART 11 April 2026). Data WA/alamat/detail dummy,
-- administrator menyesuaikan. Kata sandi awal seragam: pengurus123

INSERT INTO user (nama, no_wa, alamat, detail, password_hash, aktif, created_at) VALUES
  ('Jepri Kiat Susanto', '+628970000001', 'G11/19', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Wiyanto',            '+628970000002', 'G21/04', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Ferry',              '+628970000003', 'G12/02', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Danu Kurnianto',     '+628970000004', 'G20/23', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Ipan Sopwan Aliyudin', '+628970000005', 'G10/02', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Aji Purwaji Jumadi', '+628970000006', 'G10/21', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Wahyu Widodo',       '+628970000007', 'G21/17', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Moch. Haryanto',     '+628970000008', 'G21/19', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Bambang Subekti',    '+628970000009', 'G12/03', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('Sagito Eko Susilo',  '+628970000010', 'G11/15', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00'),
  ('M. Muhlis Chudin',   '+628970000011', 'G20/15', '', '$2a$10$dyVdxgmQUw.PyOJqkau.wu6WC0z2wJgYl23X4Nx0PcWhSGgB6v4Ju', 1, '2026-10-01T00:00:00+07:00');

INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'ketua', 0, 0 FROM user WHERE no_wa = '+628970000001';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'wakil', 0, 1 FROM user WHERE no_wa = '+628970000002';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'sekretaris', 0, 2 FROM user WHERE no_wa = '+628970000003';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'bendahara', 0, 3 FROM user WHERE no_wa = '+628970000004';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'koordinator', 1, 4 FROM user WHERE no_wa = '+628970000005';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'koordinator', 2, 5 FROM user WHERE no_wa = '+628970000006';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'pembantu_koordinator', 2, 6 FROM user WHERE no_wa = '+628970000007';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'pembantu_koordinator', 2, 7 FROM user WHERE no_wa = '+628970000008';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'koordinator', 3, 8 FROM user WHERE no_wa = '+628970000009';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'pembantu_koordinator', 3, 9 FROM user WHERE no_wa = '+628970000010';
INSERT INTO user_peran (user_id, peran, gang, urutan)
SELECT id, 'pembantu_koordinator', 3, 10 FROM user WHERE no_wa = '+628970000011';
