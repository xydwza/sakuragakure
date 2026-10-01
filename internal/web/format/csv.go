package format

import (
	"bytes"
	"encoding/csv"
)

// CSV membangun CSV (header + baris) untuk diimpor spreadsheet.
func CSV(header []string, rows [][]string) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(header)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	return buf.Bytes()
}
