package installer

import (
	"fmt"
	"path"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/deis/prow/pkg/version"
)

const prowChart = `name: prowd
description: The prow server
version: %s
apiVersion: v1
`

const prowValues = `# Default values for prowd.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 1
image:
  registry: quay.io
  org: deis
  name: prowd
  tag: %s
  pullPolicy: Always
debug: false
service:
  http:
    externalPort: 80
    internalPort: 44135
registry:
  url: quay.io
  org: deis
  # This field follows the format of Docker's X-Registry-Auth header.
  #
  # See https://github.com/docker/docker/blob/master/docs/api/v1.22.md#push-an-image-on-the-registry
  #
  # For credential-based logins, use
  #
  # $ echo '{"username":"jdoe","password":"secret","email":"jdoe@acme.com"}' | base64 -w 0
  #
  # For token-based logins, use
  #
  # $ echo '{"registrytoken":"9cbaf023786cd7"}' | base64 -w 0
  authtoken: changeme
`

const prowIgnore = `# Patterns to ignore when building packages.
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

const prowDeployment = `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: prowd
  labels:
    chart: "{{ .Chart.Name }}-{{ .Chart.Version }}"
spec:
  replicas: {{ .Values.replicaCount }}
  template:
    metadata:
      labels:
        app: prow
        name: prowd
    spec:
      containers:
      - name: prowd
        image: "{{ .Values.image.registry }}/{{ .Values.image.org }}/{{ .Values.image.name }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        args:
        - start
        - --registry-url={{ .Values.registry.url }}
        - --registry-org={{ .Values.registry.org }}
        - --registry-auth={{ .Values.registry.authtoken }}
        {{- if .Values.debug }}
        - --debug
        {{- end }}
        ports:
        - containerPort: {{ .Values.service.http.internalPort }}
        env:
        - name: PROW_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        livenessProbe:
          httpGet:
            path: /ping
            port: {{ .Values.service.http.internalPort }}
        readinessProbe:
          httpGet:
            path: /ping
            port: {{ .Values.service.http.internalPort }}
        volumeMounts:
        - mountPath: /var/run/docker.sock
          name: docker-socket
      volumes:
      - name: docker-socket
        hostPath:
          path: /var/run/docker.sock
`

const prowService = `apiVersion: v1
kind: Service
metadata:
  name: {{ .Chart.Name }}
spec:
  ports:
    - name: http
      port: {{ .Values.service.http.externalPort }}
      targetPort: {{ .Values.service.http.internalPort }}
  selector:
    app: {{ .Chart.Name }}
`

const prowNotes = `Now you can deploy an app using prow!

	$ cd my-app
	$ prow up --namespace=foo
	--> Building Dockerfile
	--> Pushing my-app:latest
	--> Deploying to Kubernetes
	--> Deployed!

That's it! You're now running your app in a Kubernetes cluster.
`

// this file left intentionally blank.
const prowHelpers = ``

var DefaultChartFiles = []*chartutil.BufferedFile{
	&chartutil.BufferedFile{
		Name: chartutil.ChartfileName,
		Data: []byte(fmt.Sprintf(prowChart, version.Release)),
	},
	&chartutil.BufferedFile{
		Name: chartutil.ValuesfileName,
		Data: []byte(fmt.Sprintf(prowValues, version.Release)),
	},
	&chartutil.BufferedFile{
		Name: chartutil.IgnorefileName,
		Data: []byte(prowIgnore),
	},
	&chartutil.BufferedFile{
		Name: path.Join(chartutil.TemplatesDir, chartutil.DeploymentName),
		Data: []byte(prowDeployment),
	},
	&chartutil.BufferedFile{
		Name: path.Join(chartutil.TemplatesDir, chartutil.ServiceName),
		Data: []byte(prowService),
	},
	&chartutil.BufferedFile{
		Name: path.Join(chartutil.TemplatesDir, chartutil.NotesName),
		Data: []byte(prowNotes),
	},
	&chartutil.BufferedFile{
		Name: path.Join(chartutil.TemplatesDir, chartutil.HelpersName),
		Data: []byte(prowHelpers),
	},
}

// Install uses the helm client to install Prowd with the given config.
//
// Returns an error if the command failed.
func Install(client *helm.Client, chartConfig *chart.Config, namespace string) error {
	chart, err := chartutil.LoadFiles(DefaultChartFiles)
	if err != nil {
		return err
	}
	_, err = client.InstallReleaseFromChart(
		chart,
		namespace,
		helm.ReleaseName("prow"),
		helm.ValueOverrides([]byte(chartConfig.Raw)))
	return err
}

//
// Upgrade uses the helm client to upgrade Prowd using the given config.
//
// Returns an error if the command failed.
func Upgrade(client *helm.Client, chartConfig *chart.Config) error {
	chart, err := chartutil.LoadFiles(DefaultChartFiles)
	if err != nil {
		return err
	}
	_, err = client.UpdateReleaseFromChart(
		"prow",
		chart,
		helm.UpdateValueOverrides([]byte(chartConfig.Raw)))
	return err
}
