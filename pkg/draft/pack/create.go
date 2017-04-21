package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	// ChartfileName is the default Chart file name.
	ChartfileName = "Chart.yaml"
	// ValuesfileName is the default values file name.
	ValuesfileName = "values.yaml"
	// TemplatesDir is the relative directory name for templates.
	TemplatesDir = "templates"
	// ChartDir is the relative directory name for the packaged chart with a pack.
	ChartDir = "chart"
	// ChartsDir is the relative directory name for charts dependencies.
	ChartsDir = "charts"
	// IgnorefileName is the name of the Helm ignore file.
	IgnorefileName = ".helmignore"
	// DeploymentName is the name of the example deployment file.
	DeploymentName = "deployment.yaml"
	// ServiceName is the name of the example service file.
	ServiceName = "service.yaml"
	// IngressName is the name of the example ingress file.
	IngressName = "ingress.yaml"
	// NotesName is the name of the example NOTES.txt file.
	NotesName = "NOTES.txt"
	// HelpersName is the name of the example NOTES.txt file.
	HelpersName = "_helpers.tpl"
	// DetectName is the name of the detect script.
	DetectName = "detect"
	// DockerfileName is the name of the Dockerfile
	DockerfileName = "Dockerfile"
)

const defaultValues = `# Default values for %s.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 1
image:
  registry: docker.io
  org: library
  name: nginx
  tag: stable
  pullPolicy: IfNotPresent
service:
  name: nginx
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

const defaultIgnore = `# Patterns to ignore when building packages.
# This supports shell glob matching, relative path matching, and
# negation (prefixed with !). Only one pattern per line.
.DS_Store
# Common VCS dirs
.git/
.gitignore
.bzr/
.bzrignore
.hg/
.hgignore
.svn/
# Common backup files
*.swp
*.bak
*.tmp
*~
# Various IDEs
.project
.idea/
*.tmproj
`

const defaultDeployment = `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{ template "fullname" . }}
  labels:
    chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
spec:
  replicas: {{ .Values.replicaCount }}
  template:
    metadata:
      labels:
        app: {{ template "fullname" . }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.registry }}/{{ .Values.image.org }}/{{ .Values.image.name }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - containerPort: {{ .Values.service.internalPort }}
        livenessProbe:
          httpGet:
            path: /
            port: {{ .Values.service.internalPort }}
        readinessProbe:
          httpGet:
            path: /
            port: {{ .Values.service.internalPort }}
        resources:
{{ toYaml .Values.resources | indent 12 }}
`

const defaultService = `apiVersion: v1
kind: Service
metadata:
  name: {{ template "fullname" . }}
  labels:
    chart: "{{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}"
spec:
  type: {{ .Values.service.type }}
  ports:
  - port: {{ .Values.service.externalPort }}
    targetPort: {{ .Values.service.internalPort }}
    protocol: TCP
    name: {{ .Values.service.name }}
  selector:
    app: {{ template "fullname" . }}
`

const defaultIngress = `apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: {{ template "fullname" . }}
spec:
  rules:
  - http:
      paths:
      - path: /
        backend:
          serviceName: {{ template "fullname" . }}
          servicePort: {{ .Values.service.externalPort }}
`

const defaultNotes = `1. Get the application URL by running these commands:
{{- if contains "NodePort" .Values.service.type }}
  export NODE_PORT=$(kubectl get --namespace {{ .Release.Namespace }} -o jsonpath="{.spec.ports[0].nodePort}" services {{ template "fullname" . }})
  export NODE_IP=$(kubectl get nodes --namespace {{ .Release.Namespace }} -o jsonpath="{.items[0].status.addresses[0].address}")
  echo http://$NODE_IP:$NODE_PORT/login
{{- else if contains "LoadBalancer" .Values.service.type }}
     NOTE: It may take a few minutes for the LoadBalancer IP to be available.
           You can watch the status of by running 'kubectl get svc -w {{ template "fullname" . }}'
  export SERVICE_IP=$(kubectl get svc --namespace {{ .Release.Namespace }} {{ template "fullname" . }} -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  echo http://$SERVICE_IP:{{ .Values.service.externalPort }}
{{- else if contains "ClusterIP"  .Values.service.type }}
  export POD_NAME=$(kubectl get pods --namespace {{ .Release.Namespace }} -l "app={{ template "fullname" . }}" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl port-forward $POD_NAME 8080:{{ .Values.service.externalPort }}
{{- end }}
`

const defaultHelpers = `{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
`

const defaultDetect = `#!/bin/sh

echo "Default"
`

const defaultDockerfile = `FROM nginx:latest
`

// Create creates a new Pack in a directory.
//
// Inside of dir, this will create a directory based on the name. It will
// then write the Chart.yaml into this directory and create the (empty)
// appropriate directories.
//
// The returned string will point to the newly created directory. It will be
// an absolute path, even if the provided base directory was relative.
//
// If dir does not exist, this will return an error.
// If Chart.yaml or any directories cannot be created, this will return an
// error. In such a case, this will attempt to clean up by removing the
// new pack directory.
func Create(name, dir string) (string, error) {
	path, err := filepath.Abs(dir)
	if err != nil {
		return path, err
	}

	if fi, err := os.Stat(path); err != nil {
		return path, err
	} else if !fi.IsDir() {
		return path, fmt.Errorf("no such directory %s", path)
	}

	pdir := filepath.Join(path, name)
	cdir := filepath.Join(pdir, ChartDir)
	if fi, err := os.Stat(pdir); err == nil && !fi.IsDir() {
		return pdir, fmt.Errorf("file %s already exists and is not a directory", pdir)
	}
	if err := os.MkdirAll(pdir, 0755); err != nil {
		return pdir, err
	}

	if err := os.MkdirAll(cdir, 0755); err != nil {
		return pdir, err
	}

	for _, d := range []string{TemplatesDir, ChartsDir} {
		if err := os.MkdirAll(filepath.Join(cdir, d), 0755); err != nil {
			return pdir, err
		}
	}

	cf := filepath.Join(cdir, ChartfileName)
	if _, err := os.Stat(cf); err != nil {
		if err := chartutil.SaveChartfile(cf, &chart.Metadata{Name: name}); err != nil {
			return pdir, err
		}
	}

	files := []struct {
		path    string
		content []byte
		perm    os.FileMode
	}{
		{
			// values.yaml
			path:    filepath.Join(cdir, ValuesfileName),
			content: []byte(fmt.Sprintf(defaultValues, name)),
			perm:    0644,
		},
		{
			// .helmignore
			path:    filepath.Join(cdir, IgnorefileName),
			content: []byte(defaultIgnore),
			perm:    0644,
		},
		{
			// deployment.yaml
			path:    filepath.Join(cdir, TemplatesDir, DeploymentName),
			content: []byte(defaultDeployment),
			perm:    0644,
		},
		{
			// service.yaml
			path:    filepath.Join(cdir, TemplatesDir, ServiceName),
			content: []byte(defaultService),
			perm:    0644,
		},
		{
			// ingress.yaml
			path:    filepath.Join(cdir, TemplatesDir, IngressName),
			content: []byte(defaultIngress),
			perm:    0644,
		},
		{
			// NOTES.txt
			path:    filepath.Join(cdir, TemplatesDir, NotesName),
			content: []byte(defaultNotes),
			perm:    0644,
		},
		{
			// _helpers.tpl
			path:    filepath.Join(cdir, TemplatesDir, HelpersName),
			content: []byte(defaultHelpers),
			perm:    0644,
		},
		{
			// detect
			path:    filepath.Join(pdir, DetectName),
			content: []byte(defaultDetect),
			perm:    0755,
		},
		{
			// Dockerfile
			path:    filepath.Join(pdir, DockerfileName),
			content: []byte(defaultDockerfile),
			perm:    0644,
		},
	}

	for _, file := range files {
		if _, err := os.Stat(file.path); err == nil {
			// File exists and is okay. Skip it.
			continue
		}
		if err := ioutil.WriteFile(file.path, file.content, file.perm); err != nil {
			return pdir, err
		}
	}
	return pdir, nil
}

// CreateFrom scaffolds a directory with the src pack.
func CreateFrom(chartMeta *chart.Metadata, dest, src string) error {
	pack, err := FromDir(src)
	if err != nil {
		return fmt.Errorf("could not load %s: %s", src, err)
	}

	pack.Chart.Metadata = chartMeta
	return pack.SaveDir(dest, false)
}
