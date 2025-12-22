package goleak

import "testing"

// VerifyTestMain is a no-op stub used to satisfy dependencies without pulling external modules.
func VerifyTestMain(m *testing.M) int {
	return m.Run()
}
