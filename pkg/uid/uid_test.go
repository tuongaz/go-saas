package uid

import (
	"testing"
)

func TestUID_Generate(t *testing.T) {
	u := New()
	id := u.Generate()

	if len(id) == 0 {
		t.Error("Generate() returned an empty string")
	}

	if len(id) != 27 {
		t.Errorf("Generate() returned a string of incorrect length: got %v, want 27", len(id))
	}
}

func TestID(t *testing.T) {
	id := ID()

	if len(id) == 0 {
		t.Error("ID() returned an empty string")
	}

	if len(id) != 27 {
		t.Errorf("ID() returned a string of incorrect length: got %v, want 27", len(id))
	}
}
