package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	k8s kubernetes.Interface
	ns  string
}

func (c *Client) CreatePod(ctx context.Context, name, image, command string, args []string) error {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				v1.Container{
					Name:    "main",
					Image:   image,
					Command: []string{command},
					Args:    args,
				},
			},
		},
	}

	_, err := c.k8s.CoreV1().
		Pods(c.ns).
		Create(ctx, pod, metav1.CreateOptions{})

	return err
}

func (c *Client) GetPodExitCode(ctx context.Context, name string) (int32, error) {
	var exitCode int32
	podCli := c.k8s.CoreV1().Pods(c.ns)
	err := wait.PollImmediate(3*time.Second, 2*time.Minute, func() (bool, error) {
		p, err := podCli.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(p.Status.ContainerStatuses) == 0 {
			return false, nil
		}
		if status := p.Status.ContainerStatuses[0].State.Terminated; status != nil {
			exitCode = status.ExitCode
			return true, nil
		}
		return false, nil
	})
	return exitCode, err
}

func (c *Client) GetPodStdOut(ctx context.Context, name string) (string, error) {
	stdout, err := c.k8s.CoreV1().
		Pods(c.ns).
		GetLogs(name, &v1.PodLogOptions{}).
		Do(ctx).
		Raw()

	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func (c *Client) DeletePod(ctx context.Context, name string) error {
	return c.k8s.CoreV1().
		Pods(c.ns).
		Delete(ctx, name, metav1.DeleteOptions{})
}

func NewClient(namespace string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags(
		"",
		filepath.Join(homedir.HomeDir(), ".kube", "config"),
	)
	if err != nil {
		return nil, err
	}
	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{k8s: k8s, ns: namespace}, nil
}

type RequestBody struct {
	Image   string   `json:"image"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type ResponseBody struct {
	ExitCode int32  `json:"exit_code"`
	Output   string `json:"output"`
}

func main() {
	cli, err := NewClient("default")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

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
	log.Println("Starting HTTP Server")
	http.ListenAndServe(":8080", nil)
}
