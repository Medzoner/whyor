package typeidentity

import "testing"

func TestInitResult(t *testing.T) {
	got := InitResult(struct{ Value int }{Value: 10})
	if got.Value != 25 || got.Box == nil || got.Box.Value != 3 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestCleanupThroughGenericAlias(t *testing.T) {
	b, cleanup := InitBox()
	cleanup()
	if b.Value != 0 {
		t.Fatal("cleanup was not registered for the alias-equivalent type")
	}
}

func TestManyThroughGenericAlias(t *testing.T) {
	boxes := InitBoxes()
	if len(boxes) != 1 || boxes[0].Value != 3 {
		t.Fatalf("unexpected boxes: %+v", boxes)
	}
}
