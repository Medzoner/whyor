package gen

import (
	"go/types"

	"golang.org/x/tools/go/types/typeutil"
)

// typeID is local to a generated package, never persisted or displayed.
type typeID int

// typeIndex interns types using Go identity (types.Identical), including aliases
// nested in signatures, interfaces, structs and generic type arguments.
// typeutil.Map hashes types without turning their printed names into identities.
type typeIndex struct {
	types typeutil.Map
}

func (i *typeIndex) key(t types.Type) typeID {
	if id := i.types.At(t); id != nil {
		return id.(typeID)
	}
	id := typeID(i.types.Len())
	i.types.Set(t, id)
	return id
}
