package laporan

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"

	"sakuragakure/internal/db"
)

// Service membuat laporan kas yang bisa diekspor CSV/Excel/PDF.
type Service struct {
	Q *db.Queries
}

func New(conn *sql.DB) *Service { return &Service{Q: db.New(conn)} }

// Baris adalah satu mutasi untuk laporan.
type Baris struct {
	Tanggal, Pos, Kategori, Keterangan string
	Masuk, Keluar                      int64
}

// Data mengambil semua mutasi untuk laporan kas.
func (s *Service) Data(ctx context.Context) ([]Baris, error) {
	rows, err := s.Q.MutasiLaporan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Baris, 0, len(rows))
	for _, r := range rows {
		b := Baris{Tanggal: r.Tanggal, Pos: r.Pos, Kategori: r.Kategori, Keterangan: r.KeteranganPublik}
		if r.Arah == "masuk" {
			b.Masuk = r.Nominal
		} else {
			b.Keluar = r.Nominal
		}
		out = append(out, b)
	}
	return out, nil
}

func total(data []Baris) (masuk, keluar int64) {
	for _, b := range data {
		masuk += b.Masuk
		keluar += b.Keluar
	}
	return
}

// CSV mengekspor ke CSV (bisa diimpor Google Sheets).
func CSV(data []Baris) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Tanggal", "Pos", "Kategori", "Keterangan", "Masuk", "Keluar"})
	for _, b := range data {
		_ = w.Write([]string{b.Tanggal, b.Pos, b.Kategori, b.Keterangan, strconv.FormatInt(b.Masuk, 10), strconv.FormatInt(b.Keluar, 10)})
	}
	masuk, keluar := total(data)
	_ = w.Write([]string{"", "", "", "TOTAL", strconv.FormatInt(masuk, 10), strconv.FormatInt(keluar, 10)})
	w.Flush()
	return buf.Bytes(), w.Error()
}

// XLSX mengekspor ke Excel dengan formula (total SUM + saldo berjalan).
func XLSX(data []Baris) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Laporan Kas"

	header := []string{"Tanggal", "Pos", "Kategori", "Keterangan", "Masuk", "Keluar", "Saldo"}
	for i, h := range header {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	_ = f.SetCellStyle(sheet, "A1", "G1", headerStyle)

	// saldo berjalan pakai formula
	prev := 0 // baris saldo sebelumnya
	for i, b := range data {
		row := i + 2
		_ = f.SetCellValue(sheet, cell(1, row), b.Tanggal)
		_ = f.SetCellValue(sheet, cell(2, row), b.Pos)
		_ = f.SetCellValue(sheet, cell(3, row), b.Kategori)
		_ = f.SetCellValue(sheet, cell(4, row), b.Keterangan)
		_ = f.SetCellValue(sheet, cell(5, row), b.Masuk)
		_ = f.SetCellValue(sheet, cell(6, row), b.Keluar)
		if prev == 0 {
			_ = f.SetCellFormula(sheet, cell(7, row), fmt.Sprintf("=E%d-F%d", row, row))
		} else {
			_ = f.SetCellFormula(sheet, cell(7, row), fmt.Sprintf("=G%d+E%d-F%d", prev, row, row))
		}
		prev = row
	}

	// baris total
	totalRow := len(data) + 2
	_ = f.SetCellValue(sheet, cell(4, totalRow), "TOTAL")
	_ = f.SetCellFormula(sheet, cell(5, totalRow), fmt.Sprintf("=SUM(E2:E%d)", totalRow-1))
	_ = f.SetCellFormula(sheet, cell(6, totalRow), fmt.Sprintf("=SUM(F2:F%d)", totalRow-1))
	_ = f.SetCellFormula(sheet, cell(7, totalRow), fmt.Sprintf("=SUM(E%d:F%d)", totalRow, totalRow))

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cell(col, row int) string {
	c, _ := excelize.CoordinatesToCellName(col, row)
	return c
}

// PDF mengekspor laporan ke PDF sederhana.
func PDF(data []Baris) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Laporan Kas RT 06 / RW 28 Sakura")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 9)

	colWidths := []float64{28, 30, 30, 80, 30, 30}
	headers := []string{"Tanggal", "Pos", "Kategori", "Keterangan", "Masuk", "Keluar"}
	pdf.SetFont("Arial", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 9)
	for _, b := range data {
		vals := []string{b.Tanggal, b.Pos, b.Kategori, b.Keterangan, rp(b.Masuk), rp(b.Keluar)}
		for i, v := range vals {
			align := "L"
			if i >= 4 {
				align = "R"
			}
			pdf.CellFormat(colWidths[i], 6, truncate(v, 40), "1", 0, align, false, 0, "")
		}
		pdf.Ln(-1)
	}
	masuk, keluar := total(data)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(colWidths[0]+colWidths[1]+colWidths[2]+colWidths[3], 7, "TOTAL", "1", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[4], 7, rp(masuk), "1", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[5], 7, rp(keluar), "1", 0, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func rp(n int64) string {
	if n == 0 {
		return ""
	}
	s := strconv.FormatInt(n, 10)
	var out bytes.Buffer
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte('.')
		}
		out.WriteRune(r)
	}
	return out.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
