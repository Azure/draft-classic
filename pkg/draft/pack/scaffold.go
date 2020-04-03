package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const DockerignoreTemplate = `
Dockerfile
draft.toml
charts/
`

const DockerfileTemplate = `
FROM ubuntu # Replace ubuntu with the base image of your choice

MAINTAINER Jane Doe <jane.doe@example.com> # Replace with maintainer info

# Use the RUN command to install tools required to build the project
RUN apt-get update && apt-get install -y package-bar package-baz package-foo

# Copy all project and build it
COPY . /path/to/project
# RUN instruction to build

# Use the EXPOSE instruction to indicate the ports on which your container
# will listen for connections
EXPOSE 80

# Use the CMD instruction to run software contained by your image with
# along with any arguments
# CMD ["/path/to/binary", "args"]
`

const TasksFileTemplate = `
# Use this file to specify tasks or commands to hook into the draft lifecycle

[pre-up]
# description = "command to run right before $ draft up (build/deploy steps)"

[post-deploy]
# description = "command to run after deploying application to cluster"

[post-delete]
# description = "command to run after deleting application from cluster"
`

// Scaffold creates a generic draft pack with a given name
func Scaffold(name string) error {
	path, err := filepath.Abs(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(name); os.IsNotExist(err) {
		if err = os.MkdirAll(name, os.FileMode(0755)); err != nil {
			return fmt.Errorf("Problem creating %s: %s", name, err)
		}
	} else {
		return fmt.Errorf("%s already exists", name)
	}

	files := []struct {
		path    string
		content []byte
	}{
		{
			// create README
			path:    filepath.Join(path, ReadmeFileName),
			content: []byte(fmt.Sprintf("#%s Draft Pack", path)),
		},
		{
			// create Dockerfile
			path:    filepath.Join(path, DockerfileName),
			content: []byte(fmt.Sprintf("#%s Draft Pack", path)),
		},
	}

	for _, file := range files {
		if _, err := os.Stat(file.path); err == nil {
			// File exists and is okay. Skip it.
			continue
		}
		if err := ioutil.WriteFile(path, []byte(createReadmeContent(name)), 0644); err != nil {
			return err
		}
	}

	//createFuncs := []func(string) error{
	//createReadme(name),
	//createDockerignore(name),
	//createDockerfile(name),
	//createTasksFile(name),
	//}

	//for _, funct := range createFuncs {
	//if err := funct(); err != nil {
	//return err
	//}
	//}

	//TODO; create a chart
	return nil
}

func createReadmeContent(name string) string {
	return fmt.Sprintf("%v pack README", name)
}
func createDockerignore(pack string) error {
	path := filepath.Join(pack, DockerignoreFileName)
	return ioutil.WriteFile(path, []byte(DockerignoreTemplate), 0644)
}

func createDockerfile(pack string) error {
	path := filepath.Join(pack, DockerfileName)
	return ioutil.WriteFile(path, []byte(DockerfileTemplate), 0644)
}

func createTasksFile(pack string) error {
	path := filepath.Join(pack, TasksFileName)
	return ioutil.WriteFile(path, []byte(TasksFileTemplate), 0644)
}

func createFile(path string) error {
	if _, err := os.Create(path); err != nil {
		return fmt.Errorf("Could not create %s: %s", path, err)
	}
	return nil
}
