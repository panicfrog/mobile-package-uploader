package api


type ApiStatus = int32
var (
	APISuccess ApiStatus = 0
	// common fail
	APIFailed ApiStatus = 1
)

type Platform = int
const (
	IOS Platform = 1
	Android Platform = 2
)