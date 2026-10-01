package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"sakuragakure/internal/db"
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
	default:
		fmt.Fprintf(os.Stderr, "perintah tidak dikenal: %s\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "pemakaian: sakuragakure <serve|migrate|import|backup|createadmin>")
}

func serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", envOr("ADDR", ":8080"), "alamat listen, contoh :8080")
	fs.Parse(args)

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("sakuragakure listen di %s", *addr)
	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatal(err)
	}
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

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
