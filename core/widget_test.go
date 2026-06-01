package core_test

import (
	"testing"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

// Compile-time interface assertions.
// These fail to compile until the types satisfy their interfaces.
var (
	_ core.Widget    = (*coretest.MockWidget)(nil)
	_ core.KeyHandler = (*coretest.MockKeyWidget)(nil)
)

func TestAlignConstants(t *testing.T) {
	if core.AlignLeft == core.AlignCenter || core.AlignCenter == core.AlignRight {
		t.Error("align constants must be distinct")
	}
}
