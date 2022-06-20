package metrics

import (
	"runtime/debug"
	"strings"
)

// GetVersion optimistically gets current client version.
//
// Does not handle replace directives.
func GetVersion() string {
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
