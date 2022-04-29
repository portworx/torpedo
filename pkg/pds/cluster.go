package pds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	pdsSystemNamespace = "pds-system"
)

type cluster struct {
	kubeconfig string
	clientset  kubernetes.Interface
}

type namespacedName struct {
	namespace string
	name      string
}

func (n namespacedName) String() string {
	return fmt.Sprintf("%s/%s", n.namespace, n.name)
}

func newCluster(kubeconfig string) (*cluster, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &cluster{
		kubeconfig: kubeconfig,
		clientset:  clientset,
	}, nil
}

func (c *cluster) getLogsForComponents(ctx context.Context, components []namespacedName, since time.Time) (string, error) {
	var logs strings.Builder
	for _, component := range components {
		log, err := c.getDeploymentLogs(ctx, component, since)
		if err != nil {
			return "", err
		}
		logs.WriteString(fmt.Sprintf("%s:\n", component))
		logs.WriteString(log)
	}
	return logs.String(), nil
}

func (c *cluster) getDeploymentLogs(ctx context.Context, deployment namespacedName, since time.Time) (string, error) {
	opts := metav1.ListOptions{
		LabelSelector: "component=" + deployment.name,
	}
	podList, err := c.clientset.CoreV1().Pods(deployment.namespace).List(ctx, opts)
	if err != nil {
		return "", err
	}
	var logs strings.Builder
	for _, pod := range podList.Items {
		podLogs, err := c.getPodLogs(ctx, pod, since)
		if err != nil {
			return "", err
		}
		logs.WriteString(fmt.Sprintf("%s:\n", pod.Name))
		logs.WriteString(podLogs)
	}
	return logs.String(), nil
}

func (c *cluster) getPodLogs(ctx context.Context, pod corev1.Pod, since time.Time) (string, error) {
	metaSince := metav1.NewTime(since)
	logOpts := &corev1.PodLogOptions{
		SinceTime: &metaSince,
	}
	req := c.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Namespace, logOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, podLogs); err != nil {
		return "", err
	}
	return buf.String(), nil
}
