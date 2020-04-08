package workspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/ZupIT/ritchie-cli/pkg/env"
)

var (
	home string
)

func TestMain(m *testing.M) {
	env.Edition = env.Single
	home = fmt.Sprintf("%s/.rit", os.TempDir())
	os.Exit(m.Run())
}

func TestCheckWorkingDir(t *testing.T) {
	workman := NewChecker(home)
	if err := workman.Check(); err != nil {
		t.Errorf("Check got %v, want %v", err, nil)
	}
}
