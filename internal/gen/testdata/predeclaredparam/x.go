//go:build whyor

package x

import "github.com/Medzoner/whyor"

func Init(error int) int { panic(whyor.Build()) }
