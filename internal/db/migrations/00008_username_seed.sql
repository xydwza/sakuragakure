-- +goose Up
-- Username + kata sandi default (password = username). Lihat cred.md.
UPDATE user SET username = 'jepri',    password_hash = '$2a$10$r7PYNAp3VH8gH/fWUiy.Xu0483hlmTq5QJR6GUoxYSXpc3Jd35weG' WHERE no_wa = '+628970000001';
UPDATE user SET username = 'wiyanto',  password_hash = '$2a$10$9CgPxXrkJO0UqwkN963Nvefpt0iNXyXGHnMPCt5jhzfF34tUlkOcS' WHERE no_wa = '+628970000002';
UPDATE user SET username = 'ferry',    password_hash = '$2a$10$.MOCYbmD1r66wF.0KkiNcOZswh2ogCB3B89KXNSSpnZz1gwAoL9He' WHERE no_wa = '+628970000003';
UPDATE user SET username = 'danu',     password_hash = '$2a$10$viXSsXwBKc57gsbIMovhkeyXd1IfgyI.mG1nMvRB0ESO6BXFoSj7O' WHERE no_wa = '+628970000004';
UPDATE user SET username = 'ipan',     password_hash = '$2a$10$ScLX8srAfrpXsOT72nMYSOPkVohf0LsQ1F8UPNUFjlItq9TLomLkG' WHERE no_wa = '+628970000005';
UPDATE user SET username = 'ajat',     password_hash = '$2a$10$FY4pHFAZ8Gv3u0GA6Friv.Tz.0EiFBJ0nT3Qxmc.gYoRev1A8V9dC' WHERE no_wa = '+628970000006';
UPDATE user SET username = 'wahyu',    password_hash = '$2a$10$D2o8wJkIY9ZA047VHZNx..cQQHuO.C.ueHEpNRIJXsWS3zoAnFZSe' WHERE no_wa = '+628970000007';
UPDATE user SET username = 'haryanto', password_hash = '$2a$10$Itx7L3tN5P8KXrVoVCG07e9vm7QiMNIrcaMXT/GaYFIcvahIlYmAO' WHERE no_wa = '+628970000008';
UPDATE user SET username = 'bambang',  password_hash = '$2a$10$l2/QIx4hNL/Xw4PJ3yjNC.mM68CT2a9ESSxkOygVYfrtYAIRxxDpq' WHERE no_wa = '+628970000009';
UPDATE user SET username = 'sagito',   password_hash = '$2a$10$C8Lu0t/iBQDe0znrcSND0.nD5ZOEKkQCD5evR5oWz9uqNmfFtrsCe' WHERE no_wa = '+628970000010';
UPDATE user SET username = 'muhlis',   password_hash = '$2a$10$wjXJCdagr0nCJAmedHZiN.MVXd/qCi3gT2us99Ty1VA91119Le51i' WHERE no_wa = '+628970000011';
