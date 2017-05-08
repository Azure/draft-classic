package defaultpacks

import (
	"path/filepath"

	"github.com/Azure/draft/pkg/draft/pack"
)

const phpValues = `# Default values for PHP.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 2
image:
  registry: docker.io
  org: library
  name: php
  tag: 7.1-apache
  pullPolicy: IfNotPresent
service:
  name: php
  type: ClusterIP
  externalPort: 80
  internalPort: 80
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
`

const phpDetect = `#!/usr/bin/env bash

if [[ -f "$1/composer.json" || -f "$1/index.php" ]]; then
  echo "PHP" && exit 0
else
  exit 1
fi

`

const phpDockerfile = `FROM php:7.1-apache
COPY . /var/www/html/
EXPOSE 80
`

// PHPFiles returns all of the files needed for the PHP default pack
// Paths are relative to the pack root
func PHPFiles() []*pack.File {
	return []*pack.File{
		{
			// values.yaml
			Path:    filepath.Join(pack.ChartDir, pack.ValuesfileName),
			Content: []byte(phpValues),
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
			Content: []byte(phpDetect),
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
			Content: []byte(phpDockerfile),
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
