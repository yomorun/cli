// Command-line tools for YoMo

package cli

import (
	"path"
	"runtime"
)

// GetRootPath get root path
func GetRootPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		return path.Dir(filename)
	}
	return ""
}
