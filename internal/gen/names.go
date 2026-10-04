package gen

import (
	"fmt"
	"strconv"
	"strings"
)

// Imports must not acquire names from the generator's local-variable namespace.
func generatedName(name string) bool {
	if name == "err" || name == "_cleanupCtx" {
		return true
	}
	for _, prefix := range []string{"v", "cleanup"} {
		if digits, ok := strings.CutPrefix(name, prefix); ok && digits != "" {
			if _, err := strconv.Atoi(digits); err == nil {
				return true
			}
		}
	}
	return false
}

func (r *resolver) nameTaken(name string) bool {
	return r.names[name] || r.f.reserved[name] || r.f.used[name] != ""
}

func (r *resolver) localName(base string) string {
	name := base
	for i := 2; r.nameTaken(name); i++ {
		name = fmt.Sprintf("%s%d", base, i)
	}
	r.names[name] = true
	return name
}
