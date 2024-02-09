package tektoncd

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/common"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	v1 "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sync"
)

var (
	instance Ops
	once     sync.Once
)

// Ops is an interface to perform kubernetes related operations on the core resources.
type Ops interface {
	taskOps
	taskRunOps
	pipelineOps
	pipelineRunOps
	// SetConfig sets the config and resets the client
	SetConfig(config *rest.Config)
}

// Instance returns a singleton instance of the client.
func Instance() Ops {
	once.Do(func() {
		if instance == nil {
			instance = &Client{}
		}
	})
	return instance
}

// SetInstance replaces the instance with the provided one. Should be used only for testing purposes.
func SetInstance(i Ops) {
	instance = i
}

// New creates a new client.
func new(tc v1.TaskInterface, tp v1.PipelineInterface, trc v1.TaskRunInterface, tpr v1.PipelineRunInterface) *Client {
	return &Client{
		V1TaskClient:        tc,
		V1PipelineClient:    tp,
		V1TaskRunClient:     trc,
		V1PipelineRunClient: tpr,
	}
}

// NewForConfig creates a new client for the given config.
func NewForConfig(c *rest.Config) (*Client, error) {
	_, err := v1.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	cs, err := versioned.NewForConfig(c)
	return &Client{
		V1TaskClient:        cs.TektonV1().Tasks(""),
		V1PipelineClient:    cs.TektonV1().Pipelines(""),
		V1TaskRunClient:     cs.TektonV1().TaskRuns(""),
		V1PipelineRunClient: cs.TektonV1().PipelineRuns(""),
	}, nil
}

// NewInstanceFromConfigFile returns new instance of client by using given
// config file
func NewInstanceFromConfigFile(config string) (Ops, error) {
	newInstance := &Client{}
	err := newInstance.loadClientFromKubeconfig(config, "")
	if err != nil {
		return nil, err
	}
	return newInstance, nil
}

// Client is a wrapper for the tekton client Ops
type Client struct {
	config              *rest.Config
	V1PipelineClient    v1.PipelineInterface
	V1TaskClient        v1.TaskInterface
	V1TaskRunClient     v1.TaskRunInterface
	V1PipelineRunClient v1.PipelineRunInterface
}

// SetConfig sets the config and resets the client
func (c *Client) SetConfig(cfg *rest.Config) {
	c.config = cfg
	c.V1PipelineClient = nil
	c.V1TaskClient = nil
	c.V1TaskRunClient = nil
	c.V1PipelineRunClient = nil
}

// initClient the k8s client if uninitialized
func (c *Client) initClient(namespace string) error {
	fmt.Println("Inside initClient")
	if c.V1PipelineClient != nil || c.V1TaskClient != nil || c.V1TaskRunClient != nil || c.V1PipelineRunClient != nil {
		fmt.Println("clients were not nil")
		return nil
	}
	fmt.Println("One of the client was nil going to set client")
	return c.setClient(namespace)
}

// setClient instantiates a client.
func (c *Client) setClient(namespace string) error {
	fmt.Println("Inside setClient")
	var err error

	if c.config != nil {
		fmt.Println("config is not nil")
		fmt.Println("calling loadClient")
		err = c.loadClient(namespace)
	} else {
		kubeconfig := os.Getenv("KUBECONFIG")
		if len(kubeconfig) > 0 {
			fmt.Println("kubeconfig is not empty")
			fmt.Println("calling loadClientFromKubeconfig")
			err = c.loadClientFromKubeconfig(kubeconfig, namespace)
		} else {
			fmt.Println("calling loadClientFromServiceAccount")
			err = c.loadClientFromServiceAccount(namespace)
		}
	}
	return err
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount specified in the pod running px
func (c *Client) loadClientFromServiceAccount(namespace string) error {
	fmt.Println("Inside loadClientFromServiceAccount")
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	c.config = config
	fmt.Println("calling loadClient")
	return c.loadClient(namespace)
}

func (c *Client) loadClientFromKubeconfig(kubeconfig string, namespace string) error {
	fmt.Println("Inside loadClientFromKubeconfig")
	fmt.Println("calling BuildConfigFromFlags")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	fmt.Println("Setting config file")
	c.config = config
	fmt.Println("calling loadClient")
	return c.loadClient(namespace)
}

func (c *Client) loadClient(namespace string) error {
	fmt.Println("Inside loadClient")
	if c.config == nil {
		fmt.Println("rest config is not provided")
		return fmt.Errorf("rest config is not provided")
	}

	var err error
	fmt.Println("calling setRateLimiter")
	err = common.SetRateLimiter(c.config)
	if err != nil {
		return err
	}
	fmt.Println("calling NewForConfig")
	cs, err := versioned.NewForConfig(c.config)
	if err != nil {
		return err
	}
	fmt.Println("setting clients")
	c.V1PipelineClient = cs.TektonV1().Pipelines(namespace)
	c.V1TaskClient = cs.TektonV1().Tasks(namespace)
	fmt.Println("cs.TektonV1().TaskRuns(namespace)", cs.TektonV1().TaskRuns(namespace))
	c.V1TaskRunClient = cs.TektonV1().TaskRuns(namespace)
	c.V1PipelineRunClient = cs.TektonV1().PipelineRuns(namespace)

	return nil
}
