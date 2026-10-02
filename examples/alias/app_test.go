package alias

import "testing"

func TestInitRepo(t *testing.T) {
	if InitRepo().DB == nil {
		t.Fatal("nil db")
	}
}
