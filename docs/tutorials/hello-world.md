# Deploying a "Hello World!" app with Draft

Let’s learn by example. In this tutorial, we’ll walk you through the creation of a basic Python application.

We’ll assume you have Draft installed already. You can tell Draft is installed and which version by running the following command in a shell prompt (indicated by the $ prefix):

```shell
$ draft version
```

## Creating a Project

If this is your first time using Draft, you’ll have to take care of some initial setup. Namely, you’ll need to auto-generate some code that establishes a Draft project – a collection of settings for an instance of Draft, including database configuration, Draft-specific options and project-specific settings.

From the command line, `cd` into a directory where you’d like to store your code, then run the following command:

```shell
$ draft new mysite
--> Ready to sail
```

This will create a `mysite` directory in your current directory.

After you create the `mysite` project, switch to its folder:

```shell
$ cd mysite
```

The `mysite` directory has a number of auto-generated files and folders that make up the structure of a Draft application. Here's a basic rundown on the function of each of the files and folders that Draft created by default:

| File/Folder | Purpose                                                                                                                                                    |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| bin/        | Contains the Draft scripts that starts your application and can contain other scripts you use to setup, update, deploy or run your app.                    |
| config/     | Configure your project's routes, database, and more. This is covered in more detail in [Configuring Draft Projects][].                                     |
| lib/        | Extended modules for your application.                                                                                                                     |
| log/        | Application log files.                                                                                                                                     |
| static/     | The only folder seen by the world as-is. Contains static files and compiled assets.                                                                        |
| Taskfile    | This file locates and loads tasks that can be run from the command line. This is covered in mode detail in [Configuring Draft Tasks][].                    |
| README.md   | This is a brief instruction manual for your application. You should edit this file to tell others what your application does, how to set it up, and so on. |
| test/       | Unit tests, fixtures, and other test apparatus. These are covered in [Testing Draft Applications][].

## Hello, Draft!

To begin with, let's get some text up on our screen quickly. To do this, you need to get your Draft application server running.

### Fire up a Web Server

When you ran `draft new`, you actually have a functional Draft application ready to deploy to Kubernetes. To see it, you need to fire up a web server on your development machine. You can do this by running the following in the `mysite` directory:

```shell
$ draft up
Draft Up Started: 'mysite': 01CDXCTNQAV00SVGWAM1782N16
mysite: Building Docker Image: SUCCESS ⚓  (38.0044s)
mysite: Releasing Application: SUCCESS ⚓  (2.1914s)
Inspect the logs with `draft logs 01CDXCTNQAV00SVGWAM1782N16`
```

### Interact with the Deployed App

Now that the application has been deployed, we can connect to our app.

```shell
$ draft connect
Connect to mysite:8080 on localhost:8080
172.17.0.1 - - [13/Sep/2017 19:10:09] "GET / HTTP/1.1" 200 -
```

To see your application in action, open a browser window and navigate to http://localhost:8080. You should see the Draft logo:

![Draft default hello world page](../img/draft-helloworld.png)

Once you're done playing with this app, cancel out of the `draft connect` session using CTRL+C.

### Say "Hello, Draft!"

To get Draft saying "Hello", you need to create at minimum a controller and a route.

A controller is the business logic layer. Each controller you write with Draft consists of a "micro-service" in the language of your choice that follows a certain convention. Draft comes with a utility that automatically generates the basic directory structure of an app, so you can focus on writing code rather than creating directories.

A route's purpose is to route requests to controllers. An important distinction to make is that it is the controller, not the route, where information is collected. The route justs displays that information.

There's an additional unused component called a service. What a service represents can vary by type; it could be a database, a cluster, or even just an account from a SaaS (software-as-a-service) offering. While it isn't necessary for the sake of a barebones "Hello World" app, it's important to point out how controllers can store state.

To create a new controller, you will need to run the "generate" task and tell it you want a controller called "hello".

Make sure you’re in the same directory as tasks.toml and type this command:

```shell
$ draft tasks generate controller hello --language python
--> Ready to sail
```

That’ll create a directory `hello`, which is laid out like this:

```shell
hello/
    app.py
    Dockerfile
```

This directory structure will house the `hello` application.

### Writing our First Route

Now that we have made the controller, we need to tell Draft where we want "Hello, Draft!" to show up. In our case, we want it to show up when we navigate to the root URL of our site, http://localhost:8080. At the moment, the Draft welcome page is occupying that spot.

We'll have to tell Draft where your actual home page is located.

Open the file `config/routes.lua` in your editor.

```
routes = {
    
}
```

This is your application's routing file which holds entries in a special DSL (domain-specific language) that tells Draft how to connect incoming requests to controllers.

When we call `draft up` and `draft connect` again, you'll see the "Hello, Draft!" message you put into `hello/app.py`, indicating that this new route is indeed going to the `hello` controller.

You should also notice a significant faster build time here. This is because Draft is caching unchanged layers and only compiling layers that need to be re-built in the background.

```shell
$ draft up
Draft Up Started: 'mysite': 01BSYA4MW4BDNAPG6VVFWEPTH8
mysite: Building Docker Image: SUCCESS ⚓  (4.0044s)
mysite: Releasing Application: SUCCESS ⚓  (1.1914s)
Inspect the logs with `draft logs 01BSYA4MW4BDNAPG6VVFWEPTH8`
$ draft connect
Connect to mysite:8080 on localhost:8080
172.17.0.1 - - [13/Sep/2017 19:10:09] "GET / HTTP/1.1" 200 -
```


[Installation Guide]: ../README.md#installation
[Helm]: https://github.com/kubernetes/helm
[Kubernetes]: https://kubernetes.io/
[Python]: https://www.python.org/
[dep007]: reference/dep-007.md
