package defaultpacks

import (
	"path/filepath"

	"github.com/deis/draft/pkg/draft/pack"
)

const golangValues = `# Default values for golang.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 2
image:
  registry: docker.io
  org: library
  name: golang
  tag: onbuild
  pullPolicy: IfNotPresent
service:
  name: golang
  type: ClusterIP
  externalPort: 80
  internalPort: 8080
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
`

const golangDetect = `
#!/usr/bin/env bash
# bin/detect <build-dir>
set -e

build=$(cd "$1/" && pwd)

if test -f "${build}/Godeps/Godeps.json" || # godeps
   test -f "${build}/vendor/vendor.json" || # govendor
   test -f "${build}/glide.yaml" || # glide
   (test -d "${build}/src" && test -n "$(find "${build}/src" -mindepth 2 -type f -name '*.go' | sed 1q)") # gb
then
  echo Go
else
  exit 1
fi
`

const golangDockerfile = `FROM golang:onbuild
EXPOSE 8080
`

// GolangFiles returns all of the files needed for the golang default pack
// Paths are relative to the pack root
func GolangFiles() []*pack.File {
	return []*pack.File{
		{
			// values.yaml
			Path:    filepath.Join(pack.ChartDir, pack.ValuesfileName),
			Content: []byte(golangValues),
			Perm:    0644,
		},
		{
			// .helmignore
			Path:    filepath.Join(pack.ChartDir, pack.IgnorefileName),
			Content: []byte(commonIgnore),
			Perm:    0644,
		},
		{
			// deployment.yaml
			Path:    filepath.Join(pack.ChartDir, pack.TemplatesDir, pack.DeploymentName),
			Content: []byte(commonDeployment),
			Perm:    0644,
		},
		{
			// service.yaml
			Path:    filepath.Join(pack.ChartDir, pack.TemplatesDir, pack.ServiceName),
			Content: []byte(commonService),
			Perm:    0644,
		},
		{
			// ingress.yaml
			Path:    filepath.Join(pack.ChartDir, pack.TemplatesDir, pack.IngressName),
			Content: []byte(commonIngress),
			Perm:    0644,
		},
		{
			// NOTES.txt
			Path:    filepath.Join(pack.ChartDir, pack.TemplatesDir, pack.NotesName),
			Content: []byte(commonNotes),
			Perm:    0644,
		},
		{
			// _helpers.tpl
			Path:    filepath.Join(pack.ChartDir, pack.TemplatesDir, pack.HelpersName),
			Content: []byte(commonHelpers),
			Perm:    0644,
		},
		{
			// detect
			Path:    filepath.Join(pack.DetectName),
			Content: []byte(golangDetect),
			Perm:    0755,
		},
		{
			// NOTICE
			Path:    filepath.Join(pack.HerokuLicenseName),
			Content: []byte(commonHerokuLicense),
			Perm:    0644,
		},
		{
			// Dockerfile
			Path:    filepath.Join(pack.DockerfileName),
			Content: []byte(golangDockerfile),
			Perm:    0644,
		},
		{
			// .dockerignore
			Path:    filepath.Join(pack.DockerignoreName),
			Content: []byte(commonDockerignore),
			Perm:    0644,
		},
	}
}
