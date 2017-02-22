package prow

import "errors"

var (
	// ErrChartNotExist is returned when no chart/ directory exists during "prow up."
	ErrChartNotExist = errors.New("chart/ does not exist. Please create it using 'prow create' before calling 'prow up'")
	// ErrDockerfileNotExist is returned when no Dockerfile exists during "prow up."
	ErrDockerfileNotExist = errors.New("Dockerfile does not exist. Please create it using 'prow create' before calling 'prow up'")
)
