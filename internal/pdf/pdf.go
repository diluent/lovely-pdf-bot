package pdf

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/jung-kurt/gofpdf"
)

const (
	pageW = 210.0 // A4 mm
	pageH = 297.0
)

// Generate creates an A4 PDF with one image per page, scaled to fit while
// preserving aspect ratio and centered on the page.
func Generate(images [][]byte) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	for i, data := range images {
		cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("image %d: %w", i+1, err)
		}

		iw, ih := float64(cfg.Width), float64(cfg.Height)
		scale := min(pageW/iw, pageH/ih)
		dw, dh := iw*scale, ih*scale
		x, y := (pageW-dw)/2, (pageH-dh)/2

		pdf.AddPage()

		name := fmt.Sprintf("img%d", i)
		opts := gofpdf.ImageOptions{ImageType: sniffType(data)}
		pdf.RegisterImageOptionsReader(name, opts, bytes.NewReader(data))
		pdf.ImageOptions(name, x, y, dw, dh, false, opts, 0, "")
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sniffType(data []byte) string {
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' {
		return "PNG"
	}
	return "JPEG"
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
