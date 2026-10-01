package auth

import "testing"

func TestOTPGenerateVerify(t *testing.T) {
	conn := newTestDB(t)
	kode, err := GenerateOTP(conn, "+628123")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(kode) != 6 {
		t.Fatalf("panjang kode harus 6, dapat %q", kode)
	}
	if err := VerifyOTP(conn, "+628123", kode); err != nil {
		t.Fatalf("verify kode benar: %v", err)
	}
	if err := VerifyOTP(conn, "+628123", "000000"); err == nil {
		t.Fatalf("kode salah harusnya gagal")
	}
}

func TestOTPPercobaanHabis(t *testing.T) {
	conn := newTestDB(t)
	kode, _ := GenerateOTP(conn, "+628123")
	for i := 0; i < otpMaksPercobaan; i++ {
		_ = VerifyOTP(conn, "+628123", "salah")
	}
	if err := VerifyOTP(conn, "+628123", kode); err == nil {
		t.Fatalf("percobaan habis harusnya gagal meski kode benar")
	}
}

func TestOTPKedaluwarsa(t *testing.T) {
	conn := newTestDB(t)
	kode, _ := GenerateOTP(conn, "+628123")
	if _, err := conn.Exec(`UPDATE otp SET kedaluwarsa = '2020-01-01T00:00:00+07:00'`); err != nil {
		t.Fatal(err)
	}
	if err := VerifyOTP(conn, "+628123", kode); err == nil {
		t.Fatalf("kode kedaluwarsa harusnya gagal")
	}
}

func TestOTPThrottle(t *testing.T) {
	conn := newTestDB(t)
	for i := 0; i < 3; i++ {
		if _, err := GenerateOTP(conn, "+628123"); err != nil {
			t.Fatalf("generate ke-%d: %v", i, err)
		}
	}
	if _, err := GenerateOTP(conn, "+628123"); err != ErrThrottle {
		t.Fatalf("ingin ErrThrottle, dapat %v", err)
	}
}
