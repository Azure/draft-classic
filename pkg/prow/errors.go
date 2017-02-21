package prow

import "errors"

var (
	ChartNotExistError      = errors.New("chart/ does not exist. Please create it using 'prow create' before calling 'prow up'.")
	DockerfileNotExistError = errors.New("Dockerfile does not exist. Please create it using 'prow create' before calling 'prow up'.")
)
