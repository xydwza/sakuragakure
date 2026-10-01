package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/justinas/nosurf"

	"sakuragakure/internal/auth"
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

	a := auth.New(conn, auth.SessionKey())

	r := chi.NewRouter()
	r.Use(a.Sessions.LoadAndSave)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.Handle("/static/*", web.StaticHandler())
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

	// koordinator
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("koordinator", "pembantu_koordinator"))
		r.Get("/gang/{gang}/tarik", h.Tarik)
		r.Post("/gang/{gang}/tarik", h.TarikToggle)
		r.Get("/gang/{gang}/setor", h.SetorPage)
		r.Post("/gang/{gang}/setor", h.SetorSubmit)
	})
	// bendahara + pengurus
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("bendahara", "ketua", "wakil"))
		r.Get("/kelola/setoran", h.KelolaSetoran)
		r.Post("/kelola/setoran/terima", h.TerimaSetoran)
		r.Post("/kelola/setoran/tolak", h.TolakSetoran)
		r.Get("/kelola/mutasi", h.MutasiPage)
		r.Get("/kelola/mutasi/baru", h.MutasiBaruPage)
		r.Post("/kelola/mutasi/baru", h.MutasiBaruSubmit)
	})
	// sekretaris + ketua (posting)
	r.Group(func(r chi.Router) {
		r.Use(a.RequirePeran("sekretaris", "ketua"))
		r.Get("/kelola/konten/posting", h.PostingPage)
		r.Post("/kelola/konten/posting", h.PostingSubmit)
	})

	r.Get("/masuk", a.LoginPage)
	r.Post("/masuk/wa", a.MintaOTP)
	r.Post("/masuk/kode", a.VerifikasiOTP)
	r.Post("/masuk/kode-cadangan", a.KodeCadangan)
	r.Post("/keluar", a.Keluar)

	log.Printf("sakuragakure listen di %s", *addr)
	if err := http.ListenAndServe(*addr, nosurf.New(r)); err != nil {
		log.Fatal(err)
	}
}

func createadmin(args []string) {
	fs := flag.NewFlagSet("createadmin", flag.ExitOnError)
	nama := fs.String("nama", "", "nama admin")
	wa := fs.String("wa", "", "nomor WhatsApp E.164 (+62...)")
	peran := fs.String("peran", "admin", "peran awal")
	gang := fs.Int("gang", 0, "gang (untuk peran koordinator/pembantu)")
	dbPath := fs.String("db", envOr("DB_PATH", "data/sakuragakure.db"), "path file SQLite")
	fs.Parse(args)

	if *nama == "" || *wa == "" {
		log.Fatal("--nama dan --wa wajib diisi")
	}

	conn, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("gagal buka database: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi gagal: %v", err)
	}

	id, err := auth.BuatUser(conn, *nama, *wa, *peran, *gang)
	if err != nil {
		log.Fatalf("gagal buat user: %v", err)
	}
	log.Printf("user %s dibuat (id %d, peran %s)", *nama, id, *peran)
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
