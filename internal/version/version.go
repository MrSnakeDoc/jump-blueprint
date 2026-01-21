package version

import (
	"runtime"
	"time"
)

var (
	Version   = "dev"                           // ex: v0.1.0
	Commit    = "none"                          // ex: abcd123
	BuildDate = time.Now().Format(time.RFC3339) // ex: 2025-08-11T18:42:00Z
	GoVersion = runtime.Version()               // go version
)
