package components

// NavItem adalah satu entri navigasi (bawah/rail).
type NavItem struct {
	Key, Label, Href string
}

var iconPath = map[string]string{
	"home":   `<path d="M3 11l9-7 9 7v9a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z"/>`,
	"kas":    `<rect x="3" y="6" width="18" height="13" rx="2"/><path d="M3 10h18M16 15h2"/>`,
	"foto":   `<rect x="3" y="5" width="18" height="15" rx="2"/><circle cx="9" cy="11" r="2"/><path d="M21 17l-5-5-9 8"/>`,
	"buku":   `<path d="M4 5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2z"/><path d="M4 19V5"/>`,
	"orang":  `<circle cx="9" cy="8" r="3"/><path d="M3 20c0-3 3-5 6-5s6 2 6 5"/><circle cx="17" cy="9" r="2.4"/><path d="M15.5 14.6c3 0 5.5 1.6 5.5 4.4"/>`,
	"rumah":  `<path d="M4 20V10l8-6 8 6v10z"/><path d="M10 20v-5h4v5"/>`,
	"form":   `<rect x="5" y="3" width="14" height="18" rx="2"/><path d="M9 8h6M9 12h6M9 16h3"/>`,
	"cek":    `<rect x="4" y="4" width="16" height="16" rx="3"/><path d="M8 12l3 3 5-6"/>`,
	"kirim":  `<path d="M4 12l16-8-6 16-2.5-6.5z"/>`,
	"grafik": `<path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>`,
	"daftar": `<path d="M8 6h13M8 12h13M8 18h13M3 6h1M3 12h1M3 18h1"/>`,
	"alert":  `<path d="M12 3l10 18H2z"/><path d="M12 10v5M12 18v.5"/>`,
	"upload": `<path d="M12 16V4M6 10l6-6 6 6M4 20h16"/>`,
}

// Icon mengembalikan SVG untuk sebuah nama ikon.
func Icon(name string) string {
	p, ok := iconPath[name]
	if !ok {
		p = `<circle cx="12" cy="12" r="8"/>`
	}
	return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">` + p + `</svg>`
}
