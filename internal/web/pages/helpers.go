package pages

import (
	"net/url"
	"strconv"
	"strings"
)

// PeranLabel menampilkan nama peran untuk struktur pengurus.
func PeranLabel(peran string, gang int) string {
	switch peran {
	case "ketua":
		return "Ketua RT"
	case "wakil":
		return "Wakil ketua RT"
	case "sekretaris":
		return "Sekretaris"
	case "bendahara":
		return "Bendahara"
	case "koordinator":
		return "Koordinator gang " + strconv.Itoa(gang)
	case "pembantu_koordinator":
		return "Pembantu koordinator gang " + strconv.Itoa(gang)
	}
	return peran
}

// Sapaan mengembalikan sapaan untuk chat WA.
func Sapaan(peran string, gang int) string {
	switch peran {
	case "ketua":
		return "Pak RT"
	case "wakil":
		return "Pak Wakil RT"
	case "sekretaris":
		return "Pak Sekretaris"
	case "bendahara":
		return "Pak Bendahara"
	case "koordinator":
		return "Pak Koordinator Gang " + strconv.Itoa(gang)
	case "pembantu_koordinator":
		return "Pak Koordinator Gang " + strconv.Itoa(gang)
	}
	return "Pengurus"
}

// WaChatLink membuat tautan wa.me dengan pesan siap kirim.
func WaChatLink(noWA, sapaan string) string {
	digit := strings.TrimPrefix(strings.TrimSpace(noWA), "+")
	pesan := "Halo " + sapaan + ", saya warga RT 06 / RW 28 Sakura."
	return "https://wa.me/" + digit + "?text=" + url.QueryEscape(pesan)
}

// AvatarInitials mengambil inisial nama (maks 2 kata).
func AvatarInitials(nama string) string {
	p := strings.Fields(nama)
	if len(p) == 0 {
		return "?"
	}
	out := p[0][:1]
	if len(p) > 1 {
		out += p[1][:1]
	}
	return strings.ToUpper(out)
}

// Ayat memecah isi_md menjadi daftar ayat (penomoran "N. " dibuang).
func Ayat(md string) []string {
	lines := strings.Split(strings.TrimSpace(md), "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if i := strings.Index(l, ". "); i > 0 && isDigits(l[:i]) {
			l = l[i+2:]
		}
		out = append(out, l)
	}
	return out
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// MonthClass kelas visual status iuran per bulan.
func MonthClass(status string) string {
	switch status {
	case "diterima":
		return "m ok"
	case "dipegang", "disetor", "menunggu_verifikasi":
		return "m held"
	}
	return "m no"
}

// MonthLabel label singkat status iuran per bulan.
func MonthLabel(status string) string {
	switch status {
	case "diterima":
		return "Lunas"
	case "dipegang", "disetor", "menunggu_verifikasi":
		return "Di koordinator"
	}
	return "Belum"
}
