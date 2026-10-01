package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/justinas/nosurf"

	"sakuragakure/internal/auth"
	"sakuragakure/internal/backup"
	"sakuragakure/internal/db"
	"sakuragakure/internal/iuran"
	"sakuragakure/internal/rumah"
	"sakuragakure/internal/web"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		serve(os.Args[2:])
	case "migrate":
		migrate(os.Args[2:])
	case "import":
		impor(os.Args[2:])
	case "createadmin":
		createadmin(os.Args[2:])
	case "tagihan":
		tagihan(os.Args[2:])
	case "backup":
		backupCmd(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "perintah tidak dikenal: %s\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "pemakaian: sakuragakure <serve|migrate|import|tagihan|backup|createadmin>")
}

func serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", envOr("ADDR", ":8080"), "alamat listen, contoh :8080")
	dbPath := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	fs.Parse(args)

	if dir := filepath.Dir(*dbPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("gagal buat direktori data: %v", err)
		}
	}
	conn, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}

	a := auth.New(conn, auth.SecretKey())

	r := chi.NewRouter()
	r.Use(a.Middleware)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Handle("/static/*", web.StaticHandler())
	r.Get("/sw.js", web.ServiceWorkerHandler())
	h := web.NewHandlers(conn, a, envOr("MEDIA_DIR", "data/media"))
	r.Get("/", h.Beranda)
	r.Get("/kas", h.Kas)
	r.Get("/aturan", h.Aturan)
	r.Get("/pengurus", h.Pengurus)
	r.Get("/kegiatan", h.Kegiatan)
	r.Get("/kegiatan/{slug}", h.AlbumDetail)
	r.Get("/m/{id}/{kind}", h.MediaServe)

	// warga
	r.Group(func(r chi.Router) {
		r.Use(a.RequireLogin)
		r.Get("/rumahku", h.Rumahku)
	})

	// koordinator (tulis alur iuran)
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("koordinator", "pembantu_koordinator"))
		r.Get("/gang/{gang}/tarik", h.Tarik)
		r.Post("/gang/{gang}/tarik", h.TarikToggle)
		r.Get("/gang/{gang}/setor", h.SetorPage)
		r.Post("/gang/{gang}/setor", h.SetorSubmit)
	})
	// baca laporan & rangkuman (semua pengurus)
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("ketua", "wakil", "sekretaris", "bendahara", "koordinator", "pembantu_koordinator"))
		r.Get("/kelola", h.KelolaPage)
		r.Get("/kelola/laporan", h.LaporanPage)
		r.Get("/kelola/laporan/kas.csv", h.LaporanCSV)
		r.Get("/kelola/laporan/kas.xlsx", h.LaporanXLSX)
		r.Get("/kelola/laporan/kas.pdf", h.LaporanPDF)
		r.Get("/kelola/setoran", h.KelolaSetoran)
		r.Get("/kelola/mutasi", h.MutasiPage)
	})
	// tulis keuangan (bendahara/ketua/wakil)
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("bendahara", "ketua", "wakil"))
		r.Post("/kelola/setoran/terima", h.TerimaSetoran)
		r.Post("/kelola/setoran/tolak", h.TolakSetoran)
		r.Get("/kelola/mutasi/baru", h.MutasiBaruPage)
		r.Post("/kelola/mutasi/baru", h.MutasiBaruSubmit)
	})
	// tulis konten (sekretaris/ketua)
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("sekretaris", "ketua"))
		r.Get("/kelola/konten/posting", h.PostingPage)
		r.Post("/kelola/konten/posting", h.PostingSubmit)
	})
	// admin (teknis)
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("admin"))
		r.Get("/admin/user", h.AdminUserList)
		r.Get("/admin/user/{id}", h.AdminUserEdit)
		r.Post("/admin/user/{id}", h.AdminUserSave)
	})

	r.Get("/masuk", h.LoginPage)
	r.Post("/masuk", a.Login)
	r.Post("/keluar", a.Logout)

	log.Printf("sakuragakure listen di %s", *addr)
	if err := http.ListenAndServe(*addr, nosurf.New(r)); err != nil {
		log.Fatal(err)
	}
}

func createadmin(args []string) {
	fs := flag.NewFlagSet("createadmin", flag.ExitOnError)
	nama := fs.String("nama", "", "nama admin")
	username := fs.String("username", "", "username login (bila kosong = nama kecil)")
	wa := fs.String("wa", "", "nomor WhatsApp E.164 (+62...)")
	password := fs.String("password", "", "kata sandi (bila kosong = username)")
	peran := fs.String("peran", "admin", "peran awal")
	gang := fs.Int("gang", 0, "gang (untuk peran koordinator/pembantu)")
	dbPath := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	fs.Parse(args)

	if *nama == "" || *wa == "" {
		log.Fatal("--nama dan --wa wajib diisi")
	}
	user := *username
	if user == "" {
		user = strings.ToLower(strings.Fields(*nama)[0])
	}

	conn, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}

	id, err := auth.BuatUser(conn, *nama, user, *wa, *peran, *password, *gang)
	if err != nil {
		log.Fatalf("gagal buat user: %v", err)
	}
	log.Printf("user %s dibuat (id %d, username %s, peran %s)", *nama, id, user, *peran)
}

func tagihan(args []string) {
	fs := flag.NewFlagSet("tagihan", flag.ExitOnError)
	periode := fs.String("periode", time.Now().Format("2006-01"), "periode YYYY-MM")
	dbPath := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	fs.Parse(args)

	conn, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}
	n, err := iuran.New(conn).GenerateTagihan(context.Background(), *periode)
	if err != nil {
		log.Fatalf("gagal buat tagihan: %v", err)
	}
	log.Printf("tagihan %s: %d dibuat", *periode, n)
}

func backupCmd(args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	dbPath := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	dir := fs.String("dir", envOr("BACKUP_DIR", "data/backup"), "direktori cadangan")
	simpan := fs.Int("simpan", 14, "jumlah hari cadangan disimpan")
	fs.Parse(args)

	conn, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()
	target, err := backup.Jalankan(conn, *dir, *simpan)
	if err != nil {
		log.Fatalf("backup gagal: %v", err)
	}
	log.Printf("backup tersimpan di %s", target)
}

func migrate(args []string) {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	path := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	fs.Parse(args)

	conn, err := db.Open(*path)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()

	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}
	log.Printf("migrasi selesai di %s", *path)
}

func impor(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	path := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	csvPath := fs.String("rumah", "seed/warga_seed.csv", "path CSV rumah")
	dryRun := fs.Bool("dry-run", false, "cetak laporan tanpa menulis ke database")
	fs.Parse(args)

	f, err := os.Open(*csvPath)
	if err != nil {
		log.Fatalf("gagal buka csv: %v", err)
	}
	defer f.Close()

	baris, err := rumah.BacaCSV(f)
	if err != nil {
		log.Fatalf("gagal baca csv: %v", err)
	}

	if *dryRun {
		fmt.Println(rumah.Laporkan(baris))
		return
	}

	conn, err := db.Open(*path)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}

	hasil, err := rumah.Impor(conn, baris)
	if err != nil {
		log.Fatalf("impor gagal: %v", err)
	}
	log.Printf("impor selesai: %d rumah, %d penghuni, %d dilewati (sudah ada)", hasil.Rumah, hasil.Penghuni, hasil.Dilewati)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
