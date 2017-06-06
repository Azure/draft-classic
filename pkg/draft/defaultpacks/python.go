package defaultpacks

import (
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/pack"
)

const pythonValues = `# Default values for python.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
basedomain: example.com
replicaCount: 2
image:
  registry: docker.io
  org: library
  name: python
  tag: onbuild
  pullPolicy: IfNotPresent
service:
  name: python
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

const pythonDetect = `#!/usr/bin/env bash

BUILD_DIR=$1

# Exit early if app is clearly not Python.
if [ ! -f $BUILD_DIR/requirements.txt ] && [ ! -f $BUILD_DIR/setup.py ] && [ ! -f $BUILD_DIR/Pipfile ]; then
  exit 1
fi

echo Python
`

const pythonDockerfile = `FROM python:onbuild
EXPOSE 8080
ENTRYPOINT ["python"]
CMD ["app.py"]
`

// PythonFiles returns all of the files needed for the python default pack
// Paths are relative to the pack root
func PythonFiles() []*pack.File {
	return []*pack.File{
		{
			// values.yaml
			Path:    filepath.Join(pack.ChartDir, pack.ValuesfileName),
			Content: []byte(pythonValues),
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
			Content: []byte(pythonDetect),
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
			Content: []byte(pythonDockerfile),
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
