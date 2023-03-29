package pkg

import "fmt"

var (
	Version string
	Build   string
)

func VersionString() string {
	return fmt.Sprintf("signalman %s (%s)", Version, Build)
}
