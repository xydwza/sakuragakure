-- Seed awal: pos_dana, setting, aturan_pasal, inventaris, jadwal_rutin
-- Sumber: SPEC §3 dan §6; teks AD/ART dari docs/mockup.html (konstanta ATURAN).

-- +goose Up
INSERT INTO pos_dana (id, nama, publik) VALUES
  ('kas_rt', 'Kas RT', 1),
  ('dana_sosial', 'Dana sosial', 1),
  ('rukem', 'Rukun kematian', 1),
  ('kas_gang_1', 'Kas gang 1', 0),
  ('kas_gang_2', 'Kas gang 2', 0),
  ('kas_gang_3', 'Kas gang 3', 0);

INSERT INTO setting (key, value, updated_at) VALUES
  ('iuran_kas_bulanan', '25000', '2026-10-01T00:00:00+07:00'),
  ('iuran_tarik_mulai', '1', '2026-10-01T00:00:00+07:00'),
  ('iuran_tarik_sampai', '10', '2026-10-01T00:00:00+07:00'),
  ('setor_batas_tanggal', '15', '2026-10-01T00:00:00+07:00'),
  ('tunggakan_blokir_bulan', '3', '2026-10-01T00:00:00+07:00'),
  ('dansos_rawat_inap', '200000', '2026-10-01T00:00:00+07:00'),
  ('dansos_melahirkan', '200000', '2026-10-01T00:00:00+07:00'),
  ('dansos_duka_ortu', '250000', '2026-10-01T00:00:00+07:00'),
  ('dansos_maks_per_tahun', '1', '2026-10-01T00:00:00+07:00'),
  ('rukem_iuran', '50000', '2026-10-01T00:00:00+07:00'),
  ('rukem_santunan', '2500000', '2026-10-01T00:00:00+07:00'),
  ('hajatan_min_hari', '3', '2026-10-01T00:00:00+07:00'),
  ('tamu_lapor_jam', '24', '2026-10-01T00:00:00+07:00'),
  ('tinggal_rumah_lapor_jam', '72', '2026-10-01T00:00:00+07:00'),
  ('warga_baru_lapor_jam', '72', '2026-10-01T00:00:00+07:00'),
  ('nama_rt', 'RT 06', '2026-10-01T00:00:00+07:00'),
  ('rukem_mode', 'per_kejadian', '2026-10-01T00:00:00+07:00'),
  ('dansos_per', 'per_orang', '2026-10-01T00:00:00+07:00');

INSERT INTO aturan_pasal (id, bab, judul, isi_md, urutan) VALUES
  ('1', 'Bab I', 'Pendahuluan', 'Warga RT 006 / RW 028 adalah warga yang menetap di wilayah Perumahan GCC, Cluster Sakura, meliputi Gang 01, Gang 02, dan Gang 03. Warga diharapkan memiliki kepedulian untuk saling menghargai dan menyayangi sebagai satu keluarga besar dengan tujuan hidup tentram, aman, berbudi luhur, cerdas, dan kreatif.', 1),
  ('2', 'Bab II', 'Kedudukan dan status warga', '1. RT 006 berada di wilayah RW 028, Perumahan Grand Cikarang City, Cluster Sakura, Desa Karangraharja, Kecamatan Cikarang Utara.
2. Warga RT 006 adalah warga yang menetap dan tinggal di wilayah RT 006, baik milik sendiri maupun menyewa.', 2),
  ('3.1', 'Bab III', 'Hak warga', '1. Mengeluarkan pendapat lisan maupun tulisan kepada pengurus.
2. Mengikuti setiap kegiatan di lingkungan RT 006 / RW 028.
3. Mengetahui laporan keuangan dan kas RT 006 / RW 028.
4. Mendapatkan pelayanan administrasi dan kewilayahan.', 3),
  ('3.2', 'Bab III', 'Kewajiban warga', '1. Berpartisipasi aktif menjaga keamanan, kebersihan, ketertiban, dan kerukunan.
2. Warga menetap wajib memiliki KTP, warga kontrak wajib memiliki KTP atau identitas lain. Setiap KK wajib membayar iuran kas RT Rp25.000 per bulan, diambil koordinator gang tanggal 1 sampai 10 dan disetorkan ke bendahara paling lambat tanggal 15.
3. KK wajib melaporkan perubahan status: datang, pindah, kelahiran, perkawinan, kematian.
4. Warga baru wajib lapor ke ketua RT maksimal 3x24 jam dengan fotokopi KTP, KK, dan surat nikah.
5. Pertemuan rutin dilaksanakan setiap 3 bulan sekali. Yang berhalangan dianggap menyetujui hasil rapat.
6. Iuran yang telah diputuskan rapat wajib dibayar tanpa kecuali.
7. Pengurusan administrasi oleh kepala keluarga atau anggota keluarga dewasa dengan identitas diri.
8. Wajib mematuhi hasil rapat, Peraturan Dasar, dan Peraturan Rumah Tangga.
9. Warga yang tidak tertib iuran selama 3 bulan berturut-turut tanpa alasan yang diterima tidak mendapatkan pelayanan administrasi dan dana sosial.', 4),
  ('3.3', 'Bab III', 'Perayaan, hajatan, dan acara keluarga', '1. Hajatan dengan hiburan atau penutupan jalan wajib lapor dan minta izin RT/RW paling lambat 3 hari sebelumnya.
2. Bila terjadi keributan, ketua RT atau pengurus berhak menghentikan acara.', 5),
  ('3.4', 'Bab III', 'Dana sosial warga', '1. Dana sosial diambil dari kas RT.
2. Rawat inap di rumah sakit atau klinik: santunan Rp200.000 (tidak berlaku untuk perawatan di rumah).
3. Ibu melahirkan: santunan Rp200.000.
4. Anggota keluarga meninggal yang tidak serumah (orang tua atau mertua): santunan Rp250.000 dari kas gang.
5. Setiap warga hanya mendapat kompensasi satu kali dalam satu tahun.', 6),
  ('3.5', 'Bab III', 'Rukun kematian', '1. Sumber dana iuran warga Rp50.000, berlaku untuk penghuni kontrak maupun tetap.
2. Setiap warga berhak mendapat santunan kematian Rp2.500.000.
3. Penghuni baru wajib membayar iuran rukun kematian.', 7),
  ('3.6', 'Bab III', 'Kerja bakti dan kebersihan', '1. Warga wajib membersihkan pekarangan dan selokan di area rumah masing-masing.
2. Kerja bakti dilaksanakan 3 bulan sekali di setiap gang.
3. Fogging dilaksanakan setiap 4 bulan sekali.', 8),
  ('3.7', 'Bab III', 'Ketertiban, keamanan, dan ronda', '1. Dilarang menggunakan rumah dan fasilitas umum untuk narkoba, judi, dan tindak kriminal lainnya.
2. Tamu atau kerabat yang menginap wajib dilaporkan ke ketua RT atau koordinator gang sebelum 1x24 jam.
3. Wajib lapor bila meninggalkan rumah 3x24 jam atau lebih.
4. Orang di luar keluarga inti yang ikut tinggal (orang tua, mertua, ART, babysitter) wajib dilaporkan paling lambat 3x24 jam.
5. Wajib ikut ronda setiap malam Minggu di gang masing-masing.
6. Dilarang berbuat anarkis membawa nama pribadi, golongan, agama, atau suku.
7. Setiap warga wajib menyalakan lampu teras pada malam hari.
8. Memarkir kendaraan di tempat aman dan terjangkau pandangan pemilik.', 9),
  ('3.8', 'Bab III', 'Lain-lain', '1. Hewan peliharaan tidak boleh mengganggu warga lain. Bila ada korban, menjadi tanggung jawab pemilik.
2. Wajib memasang bendera merah putih pada hari besar kenegaraan.
3. Wajib memiliki bak sampah. Sampah dibungkus plastik sebelum dibuang.
4. Pengurus menerima kritik dan saran untuk perbaikan lingkungan.', 10),
  ('4', 'Bab IV', 'Pindahan', '1. Warga yang pindah keluar wajib lapor ke ketua RT dan membuat surat pindah.
2. Warga yang pindah keluar bukan lagi warga RT 006.
3. Warga ber-KTP RT 006 yang tidak lagi tinggal di RT 006 bukan warga RT 006.
4. Dalam hal tertentu ketua RT berhak mengambil kebijakan dengan musyawarah pengurus.', 11);

INSERT INTO inventaris (nama, kategori, jumlah, lokasi, kondisi, catatan, bisa_dipinjam) VALUES
  ('Panggung RT', 'Panggung', '1 set', 'Gudang RT', 'baik', '', 1),
  ('Sound system Politron + 2 mic wireless', 'Audio', '1 set', 'Gudang RT', 'baik', '', 1),
  ('Terpal biru oranye 8 m', 'Perlengkapan', '1', 'Gudang RT', 'baik', '', 1),
  ('Gerobak', 'Perlengkapan', '1', 'Gudang RT', 'baik', '', 1);

INSERT INTO jadwal_rutin (judul, aturan, gang, aktif) VALUES
  ('Ronda malam Minggu', 'WEEKLY;SAT;21:00', NULL, 1),
  ('Kerja bakti', 'EVERY=3M;GANG', NULL, 1),
  ('Fogging lingkungan', 'EVERY=4M', NULL, 1),
  ('Rapat rutin warga triwulan', 'EVERY=3M', NULL, 1),
  ('Batas tarik iuran', 'MONTHLY;DAY=10', NULL, 1),
  ('Batas setor iuran', 'MONTHLY;DAY=15', NULL, 1);
