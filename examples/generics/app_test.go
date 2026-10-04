package generics

import (
	"slices"
	"testing"

	"github.com/Medzoner/whyor/examples/generics/providers"
)

func TestInitApp(t *testing.T) {
	var events providers.Events
	app, cleanup, err := InitApp(User{Name: "Ada"}, Order{ID: 42}, &events)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if app.Users.Value.Name != "Ada" || app.Orders.Value.ID != 42 || app.Pair.First.Name != "Ada" || app.Pair.Second.ID != 42 {
		t.Fatalf("wrong repository instantiations: %+v", app)
	}
	if len(app.Labels) != 3 || app.Labels[0].Name != "User" || app.Labels[1].Name != "Order" ||
		app.Labels[2].Name != "User" || app.Label.Name != "User" {
		t.Fatalf("same-signature instantiations were incorrectly merged: %+v", app)
	}
}

func TestGenericCleanup(t *testing.T) {
	var events providers.Events
	_, cleanup, err := InitApp(User{}, Order{}, &events)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	if !slices.Equal(events.Closed, []string{"Order", "User"}) {
		t.Fatalf("wrong cleanup order: %v", events.Closed)
	}
}

func TestGenericCleanupOnError(t *testing.T) {
	events := providers.Events{Fail: true}
	app, cleanup, err := InitApp(User{}, Order{}, &events)
	if err == nil || app != nil || cleanup != nil || !slices.Equal(events.Closed, []string{"User"}) {
		t.Fatalf("initialization did not clean up the previous instantiation: app=%v error=%v events=%v", app, err, events.Closed)
	}
}
