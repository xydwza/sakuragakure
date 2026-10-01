package mock

import (
	_ "embed"
	"net/http"
)

//go:embed mockup.html
var mockupHTML []byte

// Handler menyajikan mockup (mock.sakuragakure.my.id).
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(mockupHTML)
}
