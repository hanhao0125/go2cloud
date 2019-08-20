package common

import "testing"

func TestInitDB(t *testing.T) {
	DBScanRootPathWithNonRecur("/", 0)
	// wg.Wait()
}
