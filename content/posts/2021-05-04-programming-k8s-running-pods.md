---
title: "Programming Kubernetes - Running Pods"
date: 2021-05-15
draft: false
---
This is the first post of a series I want to publish about interact with the Kubernetes programmatic instead of using kubectl for example.
In this first post, we'll create an API that runs an individual pod and returns the exit code with the stdout of the execution,
something like that:

```
$ curl -i -d \
    '{"image": "python:3.7.0", "command": "python",\
    "args": ["-c", "print(\"Hello World\")"]}'\
    http://api-endpoint/

HTTP/1.1 200 OK
Content-Length: 41
Content-Type: application/json; charset=utf-8

{"exit_code": 0, "output": "Hello World\n"}
```
Explaining the request payload:
* **image**: the container image.
* **command**: main command the container will execute.
* **args**: the command arguments.

And the response payload:
* **exit code**: the container exit code.
* **output**: the stdout of container

### The Stack
We will use Golang with [client-go](https://github.com/kubernetes/client-go) library to do that.
But why? First of all, Kubernetes is written in Go and the _client-go_ is the official library used on Kubernetes ecosystem, so, there's no
reason to use another library or consuming api-server via REST requests, but, if you hate Golang, there're other options, like
[client-python](https://github.com/kubernetes-client/python) for example.
Also, I'm using minikube with Kubernetes 1.18, but, feel free to choose your Kubernetes installation.

### Code snippets
All the snippets on this post are simplified to archive better readability but don't be angry, all codes are available [here](https://github.com/drgarcia1986/drgarcia1986.github.io/tree/samples/programming-kubernetes-running-pods).

### Adding the Kubernetes Client
Let's prepare our environment to talk to Kubernetes API, starting by add _client-go_ library.
Assuming you start this project by running the `go mod init` command, you can run `go get k8s.io/client-go@latest` on the same directory of the `go.mod` file
to get the last version of the library.

After that let's create an ultra simple Golang code that lists the Kubernetes cluster namespaces (just to validate the communication with the api-server).

```go
package main

import (
    "context"
    "fmt"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

func main() {
    cfg, err := clientcmd.BuildConfigFromFlags(
        "",
        filepath.Join(homedir.HomeDir(), ".kube", "config"),
    )
    checkErr(err)

    k8s, err := kubernetes.NewForConfig(cfg)
    checkErr(err)

    nsList, err := k8s.CoreV1().
        Namespaces().
        List(context.Background(), metav1.ListOptions{})
    checkErr(err)

    for _, n := range nsList.Items {
    	fmt.Println(n.Name)
    }
}
```
Before you run this code, don't forget to run `go mod tidy` to normalize dependencies on the `go.mod` file.

And, the magic happens:
```
$ go run main.go
default
kube-node-lease
kube-public
kube-system
```

Let's take a look at the important parts of that code:

* `clientcmd.BuildConfigFromFlags(...)`: We take the Kubernetes access configuration from the `~/.kube/config` yaml file (the same as the kubectl use).
* `kubernetes.NewForConfig(...)`: We create a new k8s client based on the previous configuration.
* `k8s.CoreV1().Namespaces().List(...)`: And finally, we perform a query on Kubernetes API (same as `kubectl get ns`).

A note here, we're consuming the Kubernetes API outside the cluster, later on, in a new blog post, we'll deploy an application inside the Kubernetes with a [service-account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/).
When we do that, we don't need to inform where's the config yaml file.

### Creating and Running Pod
Now things start to get excited.
We'll create a pod manifest just as you would a yaml, but programmatic.

```go
pod := &v1.Pod{
    ObjectMeta: metav1.ObjectMeta{Name: "rtw"},
    Spec: v1.PodSpec{
        RestartPolicy: v1.RestartPolicyNever,
        Containers: []v1.Container{
            v1.Container{
                Name:    "main",
                Image:   "python:3.8",
                Command: []string{"python"},
                Args:    []string{"-c", "print('hello world')"},
            },
        },
    },
}
```
And use the Kubernetes client we created before to create the pod on the cluster:

```go
_, err = k8s.CoreV1().Pods("default").Create(
    context.Background(),
    pod,
    metav1.CreateOptions{},
)
```

That's it! If you run this code, you'll be able to see a beautifully `hello world` coming from our python container:
```
$ kubectl get pods
NAME   READY   STATUS      RESTARTS   AGE
rtw    0/1     Completed   0          3s

$ kubectl logs rtw main
hello world
```

### Getting Exit Code
To get the container exit code there's no magic or beautiful path, we need to polling Kubernetes API until the container execution terminates.
Luckily we can use the helper `PollImmediate` from `apimachinery` to perform the polling with backoff and timeout.

```go
import (
    "context"
    "fmt"
    "time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/wait"
)

...

var exitCode int32
cli := k8s.CoreV1().Pods("default")
err := wait.PollImmediate(3*time.Second, 2*time.Minute, func() (bool, error) {
    p, err := cli.Get(context.Background(), "rtw", metav1.GetOptions{})
    if err != nil {
        return false, err
    }
    if len(p.Status.ContainerStatuses) == 0 {
        return false, nil
    }
    state := p.Status.ContainerStatuses[0].State
    if state.Terminated != nil {
        exitCode = state.Terminated.ExitCode
        return true, nil
    }
    return false, nil
})
checkErr(err)
fmt.Printf("Container Exit Code: %d\n", exitCode)
```
_PollImmediate_ will execute our anonymous function until return `true`, or an `error`, or reach the timeout (second arg).

So, what's happening here?
We're getting pod information looking for the status of the main container (we created that pod with only one container).
When Kubernetes API returns the status of the container, we check if the container is terminated, if yes, we get the exit code of it.

```
Container Exit Code: 0
```

### Getting StdOut

OK, we already know the container exit code, now it's time to get the stdout, to do that, we'll get the logs from the container.

```go
stdout, err := k8s.CoreV1().
    Pods("default").
    GetLogs("rtw", &v1.PodLogOptions{}).
    Do(context.Background()).
    Raw()

checkErr(err)
fmt.Println(string(stdout))
```
Easy, right? The code above works as same as `kubectl logs`:
```
$ kubectl logs rtw main
hello world
```

### Deleting the Pod
We have all we need to build our API, but before we move, don't forget to delete the pod as we'll no longer need it.

```go
err := k8s.CoreV1().
    Pods("default").
    Delete(
        context.Background(),
        "rtw",
        metav1.DeleteOptions{},
    )

checkErr(err)
```
Done, the code above works as same as `kubectl delete pod rtw`.

### Put all the things together

Now is time to put all the things together in our HTTP Server, starting by the request and response payload:

```go
type RequestBody struct {
    Image   string   `json:"image"`
    Command string   `json:"command"`
    Args    []string `json:"args"`
}

type ResponseBody struct {
    ExitCode int32  `json:"exit_code"`
    Output   string `json:"output"`
}
```

Good, now assuming we created a struct to represent the Kubernetes client with all method we created before, we end up with something like that:

```go
type Client struct {
    k8s kubernetes.Interface
    ns  string // The default namespace we'll work
}

func (c *Client) CreatePod(ctx context.Context, name, image, command string, args []string) error {
...
}

func (c *Client) GetPodExitCode(ctx context.Context, name string) (int32, error) {
...
}

func (c *Client) GetPodStdOut(ctx context.Context, name string) (string, error) {
...
}

func (c *Client) DeletePod(ctx context.Context, name string) error {
...
}

func NewClient(namespace string) (*Client, error) {
    // logic to create the k8s client
    return &Client{k8s: k8s, ns: namespace}, nil
}
```
The last piece of code is the HTTP handler:
```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    var rb RequestBody
    if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    err := cli.CreatePod(
        ctx,
        "rtw",
        rb.Image,
        rb.Command,
        rb.Args,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    exitCode, err := cli.GetPodExitCode(ctx, "rtw")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    stdout, err := cli.GetPodStdOut(ctx, "rtw")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    go func() {
        if err := cli.DeletePod(ctx, "rtw"); err != nil {
            log.Printf("Error deleting pod: %v", err)
        }
    }()
    w.Header().Set("Content-type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(
        ResponseBody{ExitCode: exitCode, Output: stdout},
    )
})
```
Finally, put all the things together, our `main` function looks like:
```go
func main() {
    cli, err := NewClient("default")
    if err != nil {
        panic(err)
    }
    ctx := context.Background()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
       // The HTTP Handler
    })
    log.Println("Starting HTTP Server")
    http.ListenAndServe(":8080", nil)
}
```
It's done, we finally have an API that creates a Kubernetes Pod, and returns the exit code and the stdout of the container executed inside the pod:
```
$  go run main.go &
[1] 25538
2021/05/15 19:28:27 Starting HTTP Server

$ curl -i localhost:8080 -d '{"image": "python:3.8", "command": "python", "args": ["-c", "print(\"hello world\")"]}'
HTTP/1.1 200 OK
Content-Type: application/json
Date: Sat, 15 May 2021 22:28:33 GMT
Content-Length: 41

{"exit_code":0,"output":"hello world\n"}
```

### Conclusion
Consuming the Kubernetes API is easy and adds the maximum power over Kubernetes integrations.
If you want to use Kubernetes most complete and sustainable way, there's no easy path to do that without some code.
