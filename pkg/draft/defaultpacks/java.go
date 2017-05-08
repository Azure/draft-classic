package defaultpacks

import (
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/pack"
)

const javaValues = `# Default values for Java.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 2
image:
  registry: docker.io
  org: library
  name: maven
  tag: onbuild
  pullPolicy: IfNotPresent
service:
  name: java
  type: ClusterIP
  externalPort: 80
  internalPort: 4567
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
`

const javaDetect = `#!/usr/bin/env bash
# bin/use <build-dir>

if [ -f $1/pom.xml ]; then
   echo "Java" && exit 0
else
  echo "no" && exit 1
fi
`

const javaDockerfile = `FROM maven:onbuild
EXPOSE 4567
ENTRYPOINT java
CMD ["-jar", "target/helloworld-jar-with-dependencies.jar"]
`

// JavaFiles returns all of the files needed for the Java default pack
// Paths are relative to the pack root
func JavaFiles() []*pack.File {
	return []*pack.File{
		{
			// values.yaml
			Path:    filepath.Join(pack.ChartDir, pack.ValuesfileName),
			Content: []byte(javaValues),
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
			Content: []byte(javaDetect),
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
			Content: []byte(javaDockerfile),
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
