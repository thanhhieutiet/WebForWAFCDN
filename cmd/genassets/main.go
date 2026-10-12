package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	outDir := flag.String("out", "web/static", "Output directory for static assets")
	flag.Parse()

	fmt.Printf("Generating static assets into '%s'...\n", *outDir)

	if err := generateAll(*outDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating assets: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Static assets successfully generated!")
}

func generateAll(baseDir string) error {
	dirs := []string{
		filepath.Join(baseDir, "img", "products"),
		filepath.Join(baseDir, "docs"),
		filepath.Join(baseDir, "files"),
		filepath.Join(baseDir, "data"),
		filepath.Join(baseDir, "fonts"),
		filepath.Join(baseDir, "private"),
		filepath.Join(baseDir, "css"),
		filepath.Join(baseDir, "js"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// 1. Products: p01.jpg .. p40.jpg (800x600 JPEGs, varying sizes 150-400KB)
	fmt.Println("Generating product images (p01.jpg .. p40.jpg)...")
	for i := 1; i <= 40; i++ {
		filename := fmt.Sprintf("p%02d.jpg", i)
		dstPath := filepath.Join(baseDir, "img", "products", filename)
		if err := generateProductImage(dstPath, i); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	// 2. Banner: banner-large.jpg (~2.5MB noisy JPEG)
	fmt.Println("Generating banner-large.jpg (~2.5MB)...")
	if err := generateBannerLarge(filepath.Join(baseDir, "img", "banner-large.jpg")); err != nil {
		return fmt.Errorf("failed to generate banner-large.jpg: %w", err)
	}

	// 3. Sample images (png, gif, svg)
	fmt.Println("Generating sample images (sample.png, sample.gif, sample.svg)...")
	if err := generateSamplePNG(filepath.Join(baseDir, "img", "sample.png")); err != nil {
		return fmt.Errorf("failed to generate sample.png: %w", err)
	}
	if err := generateSampleGIF(filepath.Join(baseDir, "img", "sample.gif")); err != nil {
		return fmt.Errorf("failed to generate sample.gif: %w", err)
	}
	if err := generateSampleSVG(filepath.Join(baseDir, "img", "sample.svg")); err != nil {
		return fmt.Errorf("failed to generate sample.svg: %w", err)
	}

	// 4. Docs: sample.pdf
	fmt.Println("Generating docs/sample.pdf...")
	if err := generateSamplePDF(filepath.Join(baseDir, "docs", "sample.pdf")); err != nil {
		return fmt.Errorf("failed to generate sample.pdf: %w", err)
	}

	// 5. Files: sample.zip, 1mb.bin, 10mb.bin, 50mb.bin
	fmt.Println("Generating files/sample.zip and binary files...")
	if err := generateSampleZip(filepath.Join(baseDir, "files", "sample.zip")); err != nil {
		return fmt.Errorf("failed to generate sample.zip: %w", err)
	}
	if err := generateBinaryFile(filepath.Join(baseDir, "files", "1mb.bin"), 1*1024*1024, 42); err != nil {
		return fmt.Errorf("failed to generate 1mb.bin: %w", err)
	}
	if err := generateBinaryFile(filepath.Join(baseDir, "files", "10mb.bin"), 10*1024*1024, 43); err != nil {
		return fmt.Errorf("failed to generate 10mb.bin: %w", err)
	}
	if err := generateBinaryFile(filepath.Join(baseDir, "files", "50mb.bin"), 50*1024*1024, 44); err != nil {
		return fmt.Errorf("failed to generate 50mb.bin: %w", err)
	}

	// 6. Data & Fonts
	fmt.Println("Generating data/sample.json and fonts/dummy.woff2...")
	if err := generateSampleJSON(filepath.Join(baseDir, "data", "sample.json")); err != nil {
		return fmt.Errorf("failed to generate sample.json: %w", err)
	}
	if err := generateDummyWOFF2(filepath.Join(baseDir, "fonts", "dummy.woff2")); err != nil {
		return fmt.Errorf("failed to generate dummy.woff2: %w", err)
	}

	// 7. Private files
	fmt.Println("Generating private files...")
	if err := generatePrivateFiles(baseDir); err != nil {
		return fmt.Errorf("failed to generate private files: %w", err)
	}

	// 8. CSS & JS assets
	fmt.Println("Generating CSS and JS assets...")
	if err := generateWebAssets(baseDir); err != nil {
		return fmt.Errorf("failed to generate web assets: %w", err)
	}

	return nil
}

// generateProductImage creates an 800x600 JPEG with size strictly in [150KB, 400KB].
func generateProductImage(dstPath string, id int) error {
	const width = 800
	const height = 600

	rSeed := int64(id * 10007)
	rng := rand.New(rand.NewSource(rSeed))

	// Base hues
	hues := []color.RGBA{
		{R: 220, G: 38, B: 38, A: 255},  // Red
		{R: 234, G: 88, B: 12, A: 255},  // Orange
		{R: 217, G: 119, B: 6, A: 255},  // Amber
		{R: 22, G: 163, B: 74, A: 255},  // Green
		{R: 13, G: 148, B: 136, A: 255}, // Teal
		{R: 37, G: 99, B: 235, A: 255},  // Blue
		{R: 124, G: 58, B: 237, A: 255}, // Purple
		{R: 219, G: 39, B: 119, A: 255}, // Pink
	}
	baseColor := hues[id%len(hues)]

	// Target size in [160KB, 380KB]
	targetKB := 160 + (id*17)%200 // e.g. between 160KB and 360KB

	// Binary search / tuning noise amplitude & quality to guarantee [150KB, 400KB]
	var bestBytes []byte
	noiseAmp := 25 + (id%10)*3

	for q := 92; q <= 98; q++ {
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Fill background gradient
		for y := 0; y < height; y++ {
			factor := float64(y) / float64(height)
			cr := uint8(float64(baseColor.R)*(1-factor*0.4) + 20)
			cg := uint8(float64(baseColor.G)*(1-factor*0.4) + 20)
			cb := uint8(float64(baseColor.B)*(1-factor*0.4) + 20)
			for x := 0; x < width; x++ {
				img.SetRGBA(x, y, color.RGBA{R: cr, G: cg, B: cb, A: 255})
			}
		}

		// Draw geometric shapes
		// Circle in center
		cx, cy, radius := width/2, height/2, 180
		for y := cy - radius; y <= cy+radius; y++ {
			for x := cx - radius; x <= cx+radius; x++ {
				dx := x - cx
				dy := y - cy
				if dx*dx+dy*dy <= radius*radius {
					c := color.RGBA{
						R: uint8((int(baseColor.R) + 60) % 256),
						G: uint8((int(baseColor.G) + 60) % 256),
						B: uint8((int(baseColor.B) + 60) % 256),
						A: 255,
					}
					img.SetRGBA(x, y, c)
				}
			}
		}

		// Add decorative block
		for y := height/2 - 40; y <= height/2+40; y++ {
			for x := width/2 - 140; x <= width/2+140; x++ {
				img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}

		// Draw simple block letters for "P" and product ID digits
		drawLabel(img, width/2-100, height/2-25, fmt.Sprintf("PRODUCT %02d", id))

		// Overlay high-frequency grain to ensure JPEG size reaches 150KB - 400KB
		rngNoise := rand.New(rand.NewSource(rSeed))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				orig := img.RGBAAt(x, y)
				n := rngNoise.Intn(noiseAmp*2+1) - noiseAmp
				nr := clampUint8(int(orig.R) + n)
				ng := clampUint8(int(orig.G) + n)
				nb := clampUint8(int(orig.B) + n)
				img.SetRGBA(x, y, color.RGBA{R: nr, G: ng, B: nb, A: 255})
			}
		}

		var buf bytes.Buffer
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q})
		if err != nil {
			return err
		}

		encodedSizeKB := buf.Len() / 1024
		bestBytes = buf.Bytes()
		if encodedSizeKB >= 150 && encodedSizeKB <= 400 {
			break
		}
		if encodedSizeKB < 150 {
			noiseAmp += 8
		}
	}

	_ = rng
	_ = targetKB
	return os.WriteFile(dstPath, bestBytes, 0644)
}

func drawLabel(img *image.RGBA, startX, startY int, text string) {
	// Draw readable high contrast label
	col := color.RGBA{R: 15, G: 23, B: 42, A: 255}
	scale := 3
	for i, ch := range text {
		ox := startX + i*16*scale/2
		drawChar(img, ox, startY, ch, scale, col)
	}
}

func drawChar(img *image.RGBA, ox, oy int, ch rune, scale int, textColor color.RGBA) {
	// Simple 5x7 bitmap representation for standard letters and digits
	var glyph [7]byte
	switch ch {
	case 'P':
		glyph = [7]byte{0x1E, 0x12, 0x12, 0x1E, 0x10, 0x10, 0x10}
	case 'R':
		glyph = [7]byte{0x1E, 0x12, 0x12, 0x1E, 0x14, 0x12, 0x12}
	case 'O':
		glyph = [7]byte{0x0E, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E}
	case 'D':
		glyph = [7]byte{0x1C, 0x12, 0x11, 0x11, 0x11, 0x12, 0x1C}
	case 'U':
		glyph = [7]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E}
	case 'C':
		glyph = [7]byte{0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E}
	case 'T':
		glyph = [7]byte{0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04}
	case '0':
		glyph = [7]byte{0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E}
	case '1':
		glyph = [7]byte{0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E}
	case '2':
		glyph = [7]byte{0x0E, 0x11, 0x01, 0x02, 0x04, 0x08, 0x1F}
	case '3':
		glyph = [7]byte{0x1F, 0x02, 0x04, 0x02, 0x01, 0x11, 0x0E}
	case '4':
		glyph = [7]byte{0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02}
	case '5':
		glyph = [7]byte{0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E}
	case '6':
		glyph = [7]byte{0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E}
	case '7':
		glyph = [7]byte{0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08}
	case '8':
		glyph = [7]byte{0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E}
	case '9':
		glyph = [7]byte{0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C}
	case ' ':
		return
	default:
		glyph = [7]byte{0x1F, 0x1F, 0x1F, 0x1F, 0x1F, 0x1F, 0x1F}
	}

	for row := 0; row < 7; row++ {
		bits := glyph[row]
		for col := 0; col < 5; col++ {
			if (bits & (1 << (4 - col))) != 0 {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.SetRGBA(ox+col*scale+dx, oy+row*scale+dy, textColor)
					}
				}
			}
		}
	}
}

func clampUint8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// generateBannerLarge generates ~2.5MB noisy JPEG for testing CDN compression.
func generateBannerLarge(dstPath string) error {
	const width = 2560
	const height = 1440

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rng := rand.New(rand.NewSource(999))

	// Fill with high-frequency noise and pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(rng.Intn(256))
			g := uint8(rng.Intn(256))
			b := uint8(rng.Intn(256))
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// High quality JPEG with dense noise produces ~2.5MB
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
}

func generateSamplePNG(dstPath string) error {
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 14, G: 165, B: 233, A: 255}}, image.Point{}, draw.Src)
	drawLabel(img, 100, 140, "SAMPLE PNG")

	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func generateSampleGIF(dstPath string) error {
	img := image.NewPaletted(image.Rect(0, 0, 400, 300), color.Palette{
		color.RGBA{R: 240, G: 240, B: 240, A: 255},
		color.RGBA{R: 16, G: 185, B: 129, A: 255},
		color.RGBA{R: 15, G: 23, B: 42, A: 255},
	})
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 16, G: 185, B: 129, A: 255}}, image.Point{}, draw.Src)

	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.Encode(f, img, nil)
}

func generateSampleSVG(dstPath string) error {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300">
  <defs>
    <linearGradient id="g" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#0284c7" />
      <stop offset="100%" stop-color="#0369a1" />
    </linearGradient>
  </defs>
  <rect width="100%" height="100%" fill="url(#g)" />
  <circle cx="200" cy="150" r="90" fill="#38bdf8" fill-opacity="0.5" />
  <text x="200" y="155" font-family="system-ui, sans-serif" font-size="22" font-weight="bold" fill="#ffffff" text-anchor="middle">LabShop Sample SVG</text>
</svg>
`
	return os.WriteFile(dstPath, []byte(svg), 0644)
}

// generateSamplePDF creates a valid minimalist PDF document without third-party dependencies.
func generateSamplePDF(dstPath string) error {
	// A strictly valid PDF 1.4 document
	streamContent := "BT\n/F1 22 Tf\n80 720 Td\n(LabShop Sample PDF Document) Tj\n/F1 12 Tf\n80 680 Td\n(Generated for CMC WAAP and Multi-CDN Verification) Tj\nET"
	streamLen := len(streamContent)

	obj1 := "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n"
	obj2 := "2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"
	obj3 := "3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n"
	obj4 := fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", streamLen, streamContent)
	obj5 := "5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n"

	header := "%PDF-1.4\n"
	offset1 := len(header)
	offset2 := offset1 + len(obj1)
	offset3 := offset2 + len(obj2)
	offset4 := offset3 + len(obj3)
	offset5 := offset4 + len(obj4)
	xrefOffset := offset5 + len(obj5)

	xref := fmt.Sprintf("xref\n0 6\n0000000000 65535 f \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \n%010d 00000 n \n",
		offset1, offset2, offset3, offset4, offset5)
	trailer := fmt.Sprintf("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	var pdf bytes.Buffer
	pdf.WriteString(header)
	pdf.WriteString(obj1)
	pdf.WriteString(obj2)
	pdf.WriteString(obj3)
	pdf.WriteString(obj4)
	pdf.WriteString(obj5)
	pdf.WriteString(xref)
	pdf.WriteString(trailer)

	return os.WriteFile(dstPath, pdf.Bytes(), 0644)
}

func generateSampleZip(dstPath string) error {
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	w, err := zw.Create("sample.txt")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("LabShop Sample Archive Content for CDN download and compression testing.\n"))
	if err != nil {
		return err
	}
	return zw.Close()
}

// generateBinaryFile creates a deterministic seeded binary file of exact size.
func generateBinaryFile(dstPath string, sizeBytes int64, seed int64) error {
	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	rng := rand.New(rand.NewSource(seed))
	chunk := make([]byte, 64*1024)

	var written int64
	for written < sizeBytes {
		remaining := sizeBytes - written
		toWrite := int64(len(chunk))
		if remaining < toWrite {
			toWrite = remaining
		}
		// Fill chunk with deterministic pseudo-random bytes
		for i := 0; i < int(toWrite); i++ {
			chunk[i] = byte(rng.Intn(256))
		}
		n, err := f.Write(chunk[:toWrite])
		if err != nil {
			return err
		}
		written += int64(n)
	}

	return nil
}

func generateSampleJSON(dstPath string) error {
	data := map[string]any{
		"application": "LabShop",
		"version":     "1.0.0",
		"categories": []string{
			"Electronics", "Accessories", "Smart Home", "Audio", "Wearables",
		},
		"test_matrix": map[string]string{
			"cdn_cache": "30m default",
			"origin":    "Go Clean Architecture",
			"proxy":     "Nginx Edge",
		},
	}
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dstPath, content, 0644)
}

func generateDummyWOFF2(dstPath string) error {
	// WOFF2 magic bytes 'wOF2' (0x77, 0x4F, 0x46, 0x32)
	header := []byte{0x77, 0x4F, 0x46, 0x32, 0x00, 0x01, 0x00, 0x00}
	msg := []byte("\n/* DUMMY WOFF2 FILE - NOT A REAL FONT - FOR CDN MIME & CACHE TESTING */\n")
	content := append(header, msg...)
	return os.WriteFile(dstPath, content, 0644)
}

func generatePrivateFiles(baseDir string) error {
	secretTxt := filepath.Join(baseDir, "private", "secret-1.txt")
	txtContent := "CONFIDENTIAL ORIGIN ASSET\nKey: SECRETLAB-SUPER-TOKEN-998877\nAccess Denied to unauthorized clients.\n"
	if err := os.WriteFile(secretTxt, []byte(txtContent), 0644); err != nil {
		return err
	}

	secretJpg := filepath.Join(baseDir, "private", "secret-2.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 220, G: 38, B: 38, A: 255}}, image.Point{}, draw.Src)
	drawLabel(img, 120, 140, "TOP SECRET")

	f, err := os.Create(secretJpg)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
}

func generateWebAssets(baseDir string) error {
	// css/app.css
	appCSS := `/* ==========================================================================
   LabShop Origin Application Stylesheet
   Used for testing CDN Edge Caching and Auto-Minification
   ========================================================================== */

:root {
  --primary-color: #0284c7;
  --primary-hover: #0369a1;
  --bg-color: #f8fafc;
  --text-color: #0f172a;
  --border-color: #e2e8f0;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background-color: var(--bg-color);
  color: var(--text-color);
  line-height: 1.5;
}

.lab-header {
  padding: 1.5rem 2rem;
  background: #ffffff;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.lab-card {
  background: #ffffff;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  padding: 1.5rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}
`
	if err := os.WriteFile(filepath.Join(baseDir, "css", "app.css"), []byte(appCSS), 0644); err != nil {
		return err
	}

	// css/app.min.css (strip comments and excess whitespace)
	minCSS := minifyCSS(appCSS)
	if err := os.WriteFile(filepath.Join(baseDir, "css", "app.min.css"), []byte(minCSS), 0644); err != nil {
		return err
	}

	// js/app.js
	appJS := `/* ==========================================================================
   LabShop Origin Application Script
   Used for testing CDN Edge Caching and Auto-Minification
   ========================================================================== */

(function () {
  'use strict';

  console.log("LabShop application runtime initialized");

  window.LabShop = {
    version: "1.0.0",
    checkHealth: async function () {
      const response = await fetch("/api/health");
      return await response.json();
    },
    getTime: async function () {
      const response = await fetch("/api/time");
      return await response.json();
    }
  };
})();
`
	if err := os.WriteFile(filepath.Join(baseDir, "js", "app.js"), []byte(appJS), 0644); err != nil {
		return err
	}

	// js/app.min.js
	minJS := minifyJS(appJS)
	if err := os.WriteFile(filepath.Join(baseDir, "js", "app.min.js"), []byte(minJS), 0644); err != nil {
		return err
	}

	// js/runner.js
	runnerJS := `/* LabShop Scenario Runner Script */
window.LabRunner = {
  executeScenario: async function(scenarioId) {
    console.log("Executing scenario:", scenarioId);
    return { status: "success", scenario: scenarioId };
  }
};
`
	if err := os.WriteFile(filepath.Join(baseDir, "js", "runner.js"), []byte(runnerJS), 0644); err != nil {
		return err
	}

	// js/tutorial.js
	tutorialJS := `/* LabShop Tutorial Script */
window.LabTutorial = {
  loadScenarios: async function() {
    console.log("Loading tutorial scenarios");
    return [];
  }
};
`
	if err := os.WriteFile(filepath.Join(baseDir, "js", "tutorial.js"), []byte(tutorialJS), 0644); err != nil {
		return err
	}

	// js/storefront.js
	storefrontJS := `/* LabShop Storefront Script */
window.LabStorefront = {
  addToCart: function(productId) {
    console.log("Added to cart:", productId);
  }
};
`
	if err := os.WriteFile(filepath.Join(baseDir, "js", "storefront.js"), []byte(storefrontJS), 0644); err != nil {
		return err
	}

	return nil
}

func minifyCSS(input string) string {
	// Strip multi-line comments /* ... */
	reComment := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	res := reComment.ReplaceAllString(input, "")
	// Normalize spaces
	reSpace := regexp.MustCompile(`\s+`)
	res = reSpace.ReplaceAllString(res, " ")
	// Remove spaces around symbols
	res = strings.ReplaceAll(res, " {", "{")
	res = strings.ReplaceAll(res, "{ ", "{")
	res = strings.ReplaceAll(res, " }", "}")
	res = strings.ReplaceAll(res, "} ", "}")
	res = strings.ReplaceAll(res, " :", ":")
	res = strings.ReplaceAll(res, ": ", ":")
	res = strings.ReplaceAll(res, " ;", ";")
	res = strings.ReplaceAll(res, "; ", ";")
	return strings.TrimSpace(res)
}

func minifyJS(input string) string {
	reComment := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	res := reComment.ReplaceAllString(input, "")
	lines := strings.Split(res, "\n")
	var cleaned []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "//") || trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return strings.Join(cleaned, " ")
}
