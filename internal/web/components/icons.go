package components

// NavItem adalah satu entri navigasi (bawah/rail).
type NavItem struct {
	Key, Label, Href, Icon string
}

// iconAlias memetakan key halaman (nav) ke nama ikon.
var iconAlias = map[string]string{
	"beranda":  "home",
	"kegiatan": "foto",
	"aturan":   "buku",
	"pengurus": "orang",
	"rumahku":  "rumah",
	"tarik":    "cek",
	"setor":    "kirim",
	"setoran":  "kirim",
	"mutasi":   "daftar",
	"kelola":   "grafik",
	"laporan":  "grafik",
	"posting":  "upload",
	"admin":    "form",
}

// IconName mengembalikan nama ikon (fallback ke Key bila kosong).
func (n NavItem) IconName() string {
	if n.Icon != "" {
		return n.Icon
	}
	if a, ok := iconAlias[n.Key]; ok {
		return a
	}
	return n.Key
}

// iconPath berisi path ikon filled (Material-style), bukan outline lingkaran.
var iconPath = map[string]string{
	"home":   `<path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"/>`,
	"kas":    `<path d="M21 18v1c0 1.1-.9 2-2 2H5a2 2 0 0 1-2-2V5c0-1.1.9-2 2-2h14a2 2 0 0 1 2 2v1h-9a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h9zM12 16V8h10v8H12zm4-2.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3z"/>`,
	"foto":   `<path d="M21 19V5c0-1.1-.9-2-2-2H5a2 2 0 0 0-2 2v14c0 1.1.9 2 2 2h14a2 2 0 0 0 2-2zM8.5 13.5l2.5 3 3.5-4.5 4.5 6H5l3.5-4.5z"/>`,
	"buku":   `<path d="M21 5c-1.1-.35-2.3-.5-3.5-.5-1.95 0-4.05.4-5.5 1.5C10.55 4.9 8.45 4.5 6.5 4.5c-1.2 0-2.4.15-3.5.5-.75.25-1.4.55-2 1v14.65c0 .25.25.5.5.5.1 0 .15-.05.25-.05C3.1 20.45 5.05 20 6.5 20c1.95 0 4.05.4 5.5 1.5 1.35-.85 3.8-1.5 5.5-1.5 1.65 0 3.35.3 4.75 1.05.1.05.15.05.25.05.25 0 .5-.25.5-.5V6c-.6-.45-1.25-.75-2-1zm0 13.5c-1.1-.35-2.3-.5-3.5-.5-1.7 0-4.15.65-5.5 1.5V8c1.35-.85 3.8-1.5 5.5-1.5 1.2 0 2.4.15 3.5.5v11.5z"/>`,
	"orang":  `<path d="M16 11a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm-8 0a3 3 0 1 0 0-6 3 3 0 0 0 0 6zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05A4.22 4.22 0 0 1 17 16.5V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>`,
	"rumah":  `<path d="M12 3l9 8h-3v8a1 1 0 0 1-1 1h-3v-6h-4v6H7a1 1 0 0 1-1-1v-8H3z"/>`,
	"form":   `<path d="M7 2a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H7zm1 4h8v2H8V6zm0 4h8v2H8v-2zm0 4h5v2H8v-2z"/>`,
	"cek":    `<path d="M18 7l-1.41-1.41L10.25 12 7.48 9.23 6.07 10.64l4.18 4.18L18 7zm4.24-1.41L11.66 16.17 7.48 12l-1.41 1.41 5.59 5.59L23.65 6.99l-1.41-1.4zM.41 13.41L6 19l1.41-1.41L1.83 12 .41 13.41z"/>`,
	"kirim":  `<path d="M2 21L23 12 2 3v7l15 2-15 2v7z"/>`,
	"grafik": `<path d="M5 9.2h3V19H5V9.2zM10.6 5h2.8v14h-2.8V5zm5.6 8H19v6h-2.8v-6z"/>`,
	"daftar": `<path d="M3 13h2v-2H3v2zm0 4h2v-2H3v2zm0-8h2V7H3v2zm4 4h14v-2H7v2zm0 4h14v-2H7v2zM7 7v2h14V7H7z"/>`,
	"alert":  `<path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>`,
	"upload": `<path d="M9 16h6v-6h4l-7-7-7 7h4v6zm-5 2h16v2H4v-2z"/>`,
}

// Icon mengembalikan SVG filled untuk sebuah nama ikon.
func Icon(name string) string {
	p, ok := iconPath[name]
	if !ok {
		p = `<circle cx="12" cy="12" r="8"/>`
	}
	return `<svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor" aria-hidden="true">` + p + `</svg>`
}
