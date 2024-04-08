package version

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "undefined"
	}
	vcsRev := ""
	vcsTime := ""
	for _, s := range buildInfo.Settings {
		switch s.Key {
		case "vcs.revision":
			if len(s.Value) > 8 {
				vcsRev = s.Value[:8]
			} else {
				vcsRev = s.Value
			}
		case "vcs.time":
			vcsTime = s.Value
		default:
		}
		// fmt.Println(s.Key, s.Value)
	}
	if vcsRev != "" && vcsTime != "" {
		return fmt.Sprintf("%v-%v", vcsTime, vcsRev)
	}
	return "undefined"
}
