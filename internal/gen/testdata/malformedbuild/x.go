//go:build whyor

package x

import "github.com/Medzoner/whyor"

func Init() int { whyor.Build(); return 0 }
