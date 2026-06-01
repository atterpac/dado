// Package snapshots renders tcell simulation screens to PNG images.
// It has no dependency on components or testutil so it can be imported
// from component tests without creating an import cycle.
package snapshots

import (
	"embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"

	gofont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/gdamore/tcell/v2"
)

//go:embed fonts/FiraCodeNerdFontMono-Regular.ttf
var fontFS embed.FS

const (
	// CellW / CellH are the pixel dimensions of one terminal cell.
	CellW = 8
	CellH = 16
)

var (
	faceOnce sync.Once
	cellFace gofont.Face
)

// face returns the shared font face, calibrated so one character advance = CellW.
func face() gofont.Face {
	faceOnce.Do(func() {
		data, err := fontFS.ReadFile("fonts/FiraCodeNerdFontMono-Regular.ttf")
		if err != nil {
			panic("snapshots: load font: " + err.Error())
		}
		parsed, err := opentype.Parse(data)
		if err != nil {
			panic("snapshots: parse font: " + err.Error())
		}

		// Start at a trial size, measure the advance width of 'M',
		// then scale to hit exactly CellW pixels.
		const trialSize = 13.0
		const dpi = 72.0
		trial, err := opentype.NewFace(parsed, &opentype.FaceOptions{
			Size:    trialSize,
			DPI:     dpi,
			Hinting: gofont.HintingFull,
		})
		if err != nil {
			panic("snapshots: create trial face: " + err.Error())
		}
		advance, ok := trial.GlyphAdvance('M')
		trial.Close()
		if !ok || advance <= 0 {
			advance = fixed.I(CellW) // fallback
		}
		advancePx := float64(advance) / 64.0 // fixed.Int26_6 → pixels
		calibrated := trialSize * float64(CellW) / advancePx

		cellFace, err = opentype.NewFace(parsed, &opentype.FaceOptions{
			Size:    calibrated,
			DPI:     dpi,
			Hinting: gofont.HintingFull,
		})
		if err != nil {
			panic("snapshots: create calibrated face: " + err.Error())
		}
	})
	return cellFace
}

// CanvasW and CanvasH are the standard component image dimensions in cells.
// All component images are rendered at this size with the component centered.
const (
	CanvasW = 60
	CanvasH = 20
)

// RenderToPNG converts a tcell.SimulationScreen to an RGBA image at its
// natural size (screen dimensions × CellW/CellH).
func RenderToPNG(screen tcell.SimulationScreen) *image.RGBA {
	cells, w, h := screen.GetContents()
	return renderCells(cells, w, h, 0, 0, w, h)
}

// RenderCentered renders a tcell.SimulationScreen of size (compW × compH) onto
// a standard CanvasW × CanvasH canvas, centered, with canvasBg as the fill color
// for the surrounding padding area.
func RenderCentered(screen tcell.SimulationScreen, compW, compH int, canvasBg color.RGBA) *image.RGBA {
	cells, _, _ := screen.GetContents()
	offsetX := (CanvasW - compW) / 2
	offsetY := (CanvasH - compH) / 2

	img := image.NewRGBA(image.Rect(0, 0, CanvasW*CellW, CanvasH*CellH))
	draw.Draw(img, img.Bounds(), image.NewUniform(canvasBg), image.Point{}, draw.Src)

	comp := renderCells(cells, compW, compH, 0, 0, compW, compH)
	destRect := image.Rect(offsetX*CellW, offsetY*CellH, (offsetX+compW)*CellW, (offsetY+compH)*CellH)
	draw.Draw(img, destRect, comp, image.Point{}, draw.Src)
	return img
}

func renderCells(cells []tcell.SimCell, screenW, _ , x0, y0, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w*CellW, h*CellH))

	f := face()
	metrics := f.Metrics()
	ascent := metrics.Ascent.Ceil()

	for y := range h {
		for x := range w {
			cell := cells[(y0+y)*screenW+(x0+x)]
			fg, bg, _ := cell.Style.Decompose()

			bgC := tcellToRGBA(bg, color.RGBA{A: 255})
			fgC := tcellToRGBA(fg, color.RGBA{R: 255, G: 255, B: 255, A: 255})

			cellRect := image.Rect(x*CellW, y*CellH, (x+1)*CellW, (y+1)*CellH)
			draw.Draw(img, cellRect, image.NewUniform(bgC), image.Point{}, draw.Src)

			ch := ' '
			if len(cell.Runes) > 0 {
				ch = cell.Runes[0]
			}
			if ch != ' ' && ch != 0 {
				if ch >= 0x2800 && ch <= 0x28FF {
					drawBraille(img, ch, x*CellW, y*CellH, fgC)
				} else {
					drawGlyph(img, f, ch, x*CellW, y*CellH+ascent, fgC)
				}
			}
		}
	}
	return img
}

// SavePNG writes img to path as a PNG, creating parent directories as needed.
func SavePNG(img *image.RGBA, path string) error {
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// glyphFallbacks substitutes visually similar ASCII/Latin characters for
// Unicode glyphs that common Nerd Fonts omit. When a rune is missing from the
// face and has no fallback, we skip drawing it (blank cell) rather than
// rendering the .notdef red-box glyph.
var glyphFallbacks = map[rune]rune{
	'✕': '×', // U+2715 MULTIPLICATION X  → U+00D7 MULTIPLICATION SIGN
	'✗': '×', // U+2717 BALLOT X
	'✘': '×', // U+2718 HEAVY BALLOT X
	'✓': 'v', // U+2713 CHECK MARK
	'✔': 'v', // U+2714 HEAVY CHECK MARK
	'→': '>', // U+2192 RIGHTWARDS ARROW
	'←': '<', // U+2190 LEFTWARDS ARROW
	'↑': '^', // U+2191 UPWARDS ARROW
	'↓': 'v', // U+2193 DOWNWARDS ARROW
}

// drawBraille renders a Unicode braille character (U+2800–U+28FF) as pixel dots.
// The 8 dots are arranged in a 2-column × 4-row grid within the cell.
// Bit layout per Unicode standard:
//
//	dot1=bit0  dot4=bit3
//	dot2=bit1  dot5=bit4
//	dot3=bit2  dot6=bit5
//	dot7=bit6  dot8=bit7
func drawBraille(img *image.RGBA, r rune, cellX, cellY int, fg color.RGBA) {
	bits := r - 0x2800
	// dot positions: (col 0 or 1, row 0-3) for each of the 8 bit positions
	dotPos := [8][2]int{
		{0, 0}, {0, 1}, {0, 2}, // dots 1-3: left col, rows 0-2
		{1, 0}, {1, 1}, {1, 2}, // dots 4-6: right col, rows 0-2
		{0, 3}, {1, 3},         // dots 7-8: both cols, row 3
	}
	// Each dot is a 2×2 block of pixels; dots are spread across the cell
	dotW := CellW / 2 // 4px per column
	dotH := CellH / 4 // 4px per row
	for i, pos := range dotPos {
		if bits>>i&1 == 0 {
			continue
		}
		px := cellX + pos[0]*dotW + dotW/4
		py := cellY + pos[1]*dotH + dotH/4
		for dy := range 2 {
			for dx := range 2 {
				img.SetRGBA(px+dx, py+dy, fg)
			}
		}
	}
}

func drawGlyph(img *image.RGBA, face gofont.Face, r rune, px, py int, fg color.RGBA) {
	r = resolveGlyph(face, r)
	if r == 0 {
		return // no drawable glyph — leave cell background as-is
	}
	d := gofont.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fg),
		Face: face,
		Dot:  fixed.P(px, py),
	}
	d.DrawString(string(r))
}

// resolveGlyph returns r if the face has it, checks the fallback table,
// and returns 0 if nothing drawable is found.
func resolveGlyph(face gofont.Face, r rune) rune {
	if glyphPresent(face, r) {
		return r
	}
	if sub, ok := glyphFallbacks[r]; ok && glyphPresent(face, sub) {
		return sub
	}
	return 0
}

func glyphPresent(face gofont.Face, r rune) bool {
	_, _, ok := face.GlyphBounds(r)
	return ok
}

func tcellToRGBA(c tcell.Color, fallback color.RGBA) color.RGBA {
	if c == tcell.ColorDefault {
		return fallback
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	}
	tc := c.TrueColor()
	if tc == tcell.ColorDefault {
		return fallback
	}
	r, g, b := tc.RGB()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
