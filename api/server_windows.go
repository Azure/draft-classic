package api

import (
	"fmt"
)

func setupUnixHTTP(addr string) (*Server, error) {
	return nil, fmt.Errorf("Unix sockets are not supported on Windows")
}
