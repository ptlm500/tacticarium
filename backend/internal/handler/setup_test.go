package handler_test

import (
	"os"
	"testing"

	"github.com/peter/tacticarium/backend/internal/testutil"
)

func TestMain(m *testing.M) {
	env := testutil.MustSetupTestEnv()
	code := m.Run()
	env.Teardown()
	os.Exit(code)
}
