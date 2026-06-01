package testutil

import (
	"image"

	"github.com/atterpac/dado/snapshots"
)

// RenderToPNG converts the current screen state to an RGBA image.
// Each terminal cell becomes a snapshots.CellW × snapshots.CellH pixel block.
func (ts *TestScreen) RenderToPNG() *image.RGBA {
	return snapshots.RenderToPNG(ts.SimulationScreen)
}

// SavePNG writes img to path as a PNG file, creating parent directories as needed.
func SavePNG(img *image.RGBA, path string) error {
	return snapshots.SavePNG(img, path)
}
