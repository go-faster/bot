package app

import (
	"runtime/debug"
	"strings"
)

// GetGotdVersion optimistically gets current gotd version.
//
// Does not handle replace directives.
func GetGotdVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	const pkg = "github.com/gotd/td"
	for _, d := range info.Deps {
		if strings.HasPrefix(d.Path, pkg) {
			return d.Version
		}
	}
	return ""
}
