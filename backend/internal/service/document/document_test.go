package document

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

func fixtureJPEG(t *testing.T, width, height int, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode fixture jpeg: %v", err)
	}
	return buf.Bytes()
}

// fixturePDF builds a PDF with one page per color, in order, so tests can
// assert both total page count and page order after a merge.
func fixturePDF(t *testing.T, pageColors ...color.RGBA) []byte {
	t.Helper()
	conf := model.NewDefaultConfiguration()

	pages := make([]io.ReadSeeker, len(pageColors))
	for i, c := range pageColors {
		img := fixtureJPEG(t, 80, 80, c)
		var page bytes.Buffer
		if err := api.ImportImages(nil, &page, []io.Reader{bytes.NewReader(img)}, nil, conf); err != nil {
			t.Fatalf("build fixture pdf page %d: %v", i, err)
		}
		pages[i] = bytes.NewReader(page.Bytes())
	}

	if len(pages) == 1 {
		var single bytes.Buffer
		if _, err := single.ReadFrom(pages[0]); err != nil {
			t.Fatalf("read single fixture page: %v", err)
		}
		return single.Bytes()
	}

	var out bytes.Buffer
	if err := api.MergeRaw(pages, &out, false, conf); err != nil {
		t.Fatalf("assemble multi-page fixture pdf: %v", err)
	}
	return out.Bytes()
}

func TestMerge_ImagesOnly(t *testing.T) {
	svc := NewService()

	files := []service.InputFile{
		{Name: "a.jpg", Data: fixtureJPEG(t, 100, 100, color.RGBA{255, 0, 0, 255}), ContentType: "image/jpeg"},
		{Name: "b.jpg", Data: fixtureJPEG(t, 100, 100, color.RGBA{0, 255, 0, 255}), ContentType: "image/jpeg"},
		{Name: "c.jpg", Data: fixtureJPEG(t, 100, 100, color.RGBA{0, 0, 255, 255}), ContentType: "image/jpeg"},
	}

	result, err := svc.Merge(files, 500)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if result.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", result.TotalPages)
	}
	if result.SizeKB <= 0 {
		t.Errorf("SizeKB = %.2f, want > 0", result.SizeKB)
	}
	if len(result.Data) == 0 {
		t.Error("Data is empty")
	}
}

func TestMerge_ImagesAndMultiPagePDF_PreservesOrder(t *testing.T) {
	svc := NewService()

	red := color.RGBA{255, 0, 0, 255}
	green := color.RGBA{0, 255, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	yellow := color.RGBA{255, 255, 0, 255}

	// A 2-page PDF (green, blue) sandwiched between two single images (red, yellow).
	// Expected final page order: red, green, blue, yellow.
	files := []service.InputFile{
		{Name: "first.jpg", Data: fixtureJPEG(t, 90, 90, red), ContentType: "image/jpeg"},
		{Name: "middle.pdf", Data: fixturePDF(t, green, blue), ContentType: "application/pdf"},
		{Name: "last.jpg", Data: fixtureJPEG(t, 90, 90, yellow), ContentType: "image/jpeg"},
	}

	result, err := svc.Merge(files, 500)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	if result.TotalPages != 4 {
		t.Fatalf("TotalPages = %d, want 4 (1 + 2 + 1)", result.TotalPages)
	}
	if result.SizeKB <= 0 {
		t.Errorf("SizeKB = %.2f, want > 0", result.SizeKB)
	}

	wantOrder := []color.RGBA{red, green, blue, yellow}
	gotOrder := pageDominantColors(t, result.Data)
	if len(gotOrder) != len(wantOrder) {
		t.Fatalf("extracted %d page images, want %d", len(gotOrder), len(wantOrder))
	}
	for i, want := range wantOrder {
		if !closeColor(gotOrder[i], want) {
			t.Errorf("page %d color = %v, want %v (order across files and within the multi-page PDF must be preserved)", i+1, gotOrder[i], want)
		}
	}
}

// pageDominantColors extracts each page's embedded image and returns its
// top-left pixel color, ordered by page number, to verify merge ordering.
func pageDominantColors(t *testing.T, pdfData []byte) []color.RGBA {
	t.Helper()
	conf := model.NewDefaultConfiguration()

	pageImages, err := api.ExtractImagesRaw(bytes.NewReader(pdfData), nil, conf)
	if err != nil {
		t.Fatalf("ExtractImagesRaw: %v", err)
	}

	byPage := make(map[int]color.RGBA)
	for _, m := range pageImages {
		for _, im := range m {
			data, err := io.ReadAll(im)
			if err != nil {
				t.Fatalf("read extracted image: %v", err)
			}
			decoded, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("decode extracted image: %v", err)
			}
			r, g, b, a := decoded.At(0, 0).RGBA()
			byPage[im.PageNr] = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		}
	}

	out := make([]color.RGBA, 0, len(byPage))
	for i := 1; i <= len(byPage); i++ {
		c, ok := byPage[i]
		if !ok {
			t.Fatalf("missing extracted image for page %d", i)
		}
		out = append(out, c)
	}
	return out
}

// closeColor tolerates the small rounding drift JPEG re-encoding introduces.
func closeColor(a, b color.RGBA) bool {
	const tolerance = 10
	diff := func(x, y uint8) int {
		d := int(x) - int(y)
		if d < 0 {
			d = -d
		}
		return d
	}
	return diff(a.R, b.R) <= tolerance && diff(a.G, b.G) <= tolerance && diff(a.B, b.B) <= tolerance
}

func TestMerge_NoFiles(t *testing.T) {
	svc := NewService()
	if _, err := svc.Merge(nil, 500); err == nil {
		t.Fatal("expected an error for empty input, got nil")
	}
}
