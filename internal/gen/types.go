package gen

import (
	"go/token"
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

// providerKey distinguishes explicit instantiations, including type arguments
// which do not occur in the provider's input or output signature.
type providerKey struct {
	fn   *types.Func
	args typeID
}

func (i *typeIndex) provider(fn *types.Func, args []types.Type) providerKey {
	key := providerKey{fn: fn, args: -1}
	if len(args) == 0 {
		return key
	}
	params := make([]*types.Var, len(args))
	for n, t := range args {
		params[n] = types.NewVar(token.NoPos, nil, "", t)
	}
	// A synthetic signature gives typeutil.Map an ordered, identity-aware key
	// for the arguments without serializing their names or comparing pointers.
	sig := types.NewSignatureType(nil, nil, nil, types.NewTuple(params...), types.NewTuple(), false)
	key.args = i.key(sig)
	return key
}

func (i *typeIndex) key(t types.Type) typeID {
	if id := i.types.At(t); id != nil {
		return id.(typeID)
	}
	id := typeID(i.types.Len())
	i.types.Set(t, id)
	return id
}
