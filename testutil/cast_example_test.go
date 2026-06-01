package testutil_test

import (
	"os"
	"testing"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/testutil"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// TestBadgeCast demonstrates the CastRecorder pattern:
//   - Step through component states with simulated actions
//   - Assert the cast is byte-for-byte stable (golden file)
//   - Optionally write a .cast file for the asciinema player
//
// Run UPDATE_CAST=1 go test ./testutil/ -run TestBadgeCast to regenerate.
// Run WRITE_CAST=1 go test ./testutil/ -run TestBadgeCast to emit badge.cast.
func TestBadgeCast(t *testing.T) {
	theme.Default().SetAutoRefresh(false)
	defer theme.Default().SetAutoRefresh(true)
	theme.Default().SetTheme(themes.Nord)

	badge := components.NewBadge("NEW").SetVariant(components.BadgeDefault)

	rec := testutil.NewCastRecorder(20, 3)

	// Frame 1: default variant
	rec.Capture(badge, 1.0)

	// Frame 2: success variant
	rec.Step(badge, func() {
		badge.SetVariant(components.BadgeSuccess)
	}, 1.0)

	// Frame 3: warning variant
	rec.Step(badge, func() {
		badge.SetVariant(components.BadgeWarning)
	}, 1.0)

	// Frame 4: error variant
	rec.Step(badge, func() {
		badge.SetVariant(components.BadgeError)
	}, 1.5)

	rec.AssertGolden(t, "testdata/cast/badge.cast", "Badge Variants")

	if os.Getenv("WRITE_CAST") != "" {
		if err := rec.WriteTo("badge.cast", "Badge Variants"); err != nil {
			t.Fatal(err)
		}
		t.Log("wrote badge.cast")
	}
}
