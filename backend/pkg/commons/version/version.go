package version

import (
	"net/http"
	"runtime"
)

// Build information. Populated at build-time
var (
	Version   = "undefined"
	GitDate   = "undefined"
	GitCommit = "undefined"
	BuildDate = "undefined"
	GoVersion = runtime.Version()
)

func HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Beaconchain-Version", Version)
		next.ServeHTTP(w, r)
	})
}
