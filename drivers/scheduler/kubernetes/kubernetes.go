
package kubernetes

import (
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	sv1 "k8s.io/client-go/pkg/apis/storage/v1beta1"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"time"
)

type k8sdriver struct {
}

const (
	defaultServer string = "http://127.0.0.1:8001"
	serverEnvVar  string = "SERVER"
	tokenEnvVar   string = "TOKEN"
	caFileEnvVar  string = "CA_FILE"
)

var logger *log.Logger

//k8sConnect connects to kubernetes master with ip,bearer token
//and certificate
func k8sConnect(ip string) (*kubernetes.Clientset, error) {
	token := os.Getenv("tokenEnvVar")
	caFile := os.Getenv("caFileEnvVar")
	var err error
	var caData []byte
	if len(caFile) > 0 {
		caData, err = ioutil.ReadFile(caFile)
	}
	if err != nil {
		logger.Println(err)
		return nil,err
	}
	config := &rest.Config{
		Host:            "https://" + ip + ":6443",
		BearerToken:     token,
		TLSClientConfig: rest.TLSClientConfig{CAData: caData},
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

//createPersistentVolumeClaimSpec returns representation
//of PersistentVolumeClaim
func createPersistentVolumeClaimSpec(task Task) *v1.PersistentVolumeClaim {
	pvc := v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "torpedo",
			Annotations: map[string]string{
				"volume.beta.kubernetes.io/storage-class": "torpedo",
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("2Gi"),
				},
			},
		},
	}
	return &pvc
}

//createStorageClassSpec returns representation
//of StorageClass
func createStorageClassSpec(task Task) *sv1.StorageClass {
	sc := sv1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "torpedo",
		},
		Provisioner: "kubernetes.io/portworx-volume",
		Parameters: map[string]string{
			"repl": "1",
		},
	}
	return &sc
}

func (d *k8sdriver) Init() error {
	log.Printf("Using the Kuberntes scheduler driver.\n")
	log.Printf("The following hosts are in the cluster: %v.\n", nodes)
	return nil
}

func (d *k8sdriver) GetNodes() ([]string, error) {
	return nodes, nil
}

func (d *k8sdriver) Create(task Task) (*Context, error) {
	clientset, err := k8sConnect(task.IP)
	if err != nil {
		return nil, err
	}
	context := Context{}
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "POD",
			APIVersion: "extensions/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "torpedo",
			Labels: map[string]string{"app": task.Name},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				v1.Volume{
					Name: "torpedo",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "torpedo",
						},
					},
				},
			},
			Containers: []v1.Container{
				v1.Container{
					Name:            "torpedo",
					Image:           task.Img,
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         task.Cmd,
					VolumeMounts: []v1.VolumeMount{
						v1.VolumeMount{
							MountPath: task.Vol.Path,
							Name:      "torpedo",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
		Status: v1.PodStatus{},
	}
	context.Task = task
	context.ID = "torpedo"
	context.PodSpec = &pod
	pvc := createPersistentVolumeClaimSpec(task)
	_, err = clientset.Core().
		PersistentVolumeClaims("kube-system").
		Create(pvc)
	if err != nil {
		return nil, err
	}
	sc := createStorageClassSpec(task)
	_, err = clientset.StorageV1beta1().
		StorageClasses().Create(sc)
	if err != nil {
		return nil, err
	}
	return &context, nil
}

//Run to completion
func (d *k8sdriver) Run(ctx *Context) error {
	clientset, err := k8sConnect(ctx.Task.IP)
	if err != nil {
		return err
	}
	_, err = clientset.Core().Pods("kube-system").
		Create(ctx.PodSpec)
	if err != nil {
		return err
	}
	log.Printf("Pod created")
	time.Sleep(60 * time.Second)
	podOptions := metav1.GetOptions{}
	pod, errr := clientset.Core().Pods("kube-system").
		Get("torpedo", podOptions)
	if errr != nil {
		return err
	}
	for {
		if pod.Status.Phase == v1.PodSucceeded {
			ctx.Status = 0
			break
		}

		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodUnknown {
			ctx.Status = 1
			break
		}
		pod, err = clientset.Core().Pods("kube-system").
			Get("torpedo", podOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *k8sdriver) Start(ctx *Context) error {
	clientset, err := k8sConnect(ctx.Task.IP)
	if err != nil {
		return err
	}
	_, err = clientset.Core().Pods("kube-system").
		Create(ctx.PodSpec)

	if err != nil {
		return err
	}
	return nil
}

func (d *k8sdriver) WaitDone(ctx *Context) error {
	clientset, err := k8sConnect(ctx.Task.IP)
	if err != nil {
		return err
	}
	time.Sleep(60 * time.Second)
	podOptions := metav1.GetOptions{}
	pod, errr := clientset.Core().Pods("kube-system").
		Get("torpedo", podOptions)
	if errr != nil {
		return err
	}
	for {
	        if pod.Status.Phase == v1.PodSucceeded {
			ctx.Status = 0
			break
		}

		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodUnknown {
			ctx.Status = 1
			break
		}
		pod, err = clientset.Core().Pods("kube-system").
			Get("torpedo", podOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *k8sdriver) InspectVolume(ip, name string) (*Volume, error) {
	clientset, err := k8sConnect(ip)
	if err != nil {
		return nil, err
	}
	vol, err := clientset.StorageV1().
		StorageClasses().
		Get("torpedo", metav1.GetOptions{})
	var driver string
	if vol.Provisioner == "kubernetes.io/portworx-volume" {
		driver = "pxd_k8s"
	}
	v := Volume{
		Driver: driver,
	}
	return &v, nil
}

func (d *k8sdriver) DeleteVolume(ip, name string) error {
	clientset, err := k8sConnect(ip)
	if err != nil {
		return err
	}
	err = clientset.PersistentVolumeClaims("kube-system").
		Delete("torpedo", &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = clientset.
		StorageV1beta1().
		StorageClasses().
		Delete("torpedo", &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (d *k8sdriver) DestroyByName(ip, name string) error {
	clientset, err := k8sConnect(ip)
	if err != nil {
		return err
	}
	err = clientset.Core().
		PersistentVolumeClaims("kube-system").
		Delete("torpedo", &metav1.DeleteOptions{})

	if err != nil {
		log.Println(err)
	} else {
		log.Println("PVC deleted")
	}
	err = clientset.
		StorageV1beta1().
		StorageClasses().
		Delete("torpedo", &metav1.DeleteOptions{})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("SC deleted")
	}
	err = clientset.Pods("kube-system").
		Delete("torpedo", &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.Printf("Pod torpedo deleted")
	log.Printf("Deleted task: %v\n", name)
	return nil
}

func (d *k8sdriver) Destroy(ctx *Context) error {
	clientset, err := k8sConnect(ctx.Task.IP)
	if err != nil {
		return err
	}
	err = clientset.Pods("kube-system").
		Delete(ctx.ID, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	log.Println("Pod Deleted")
	err = clientset.Core().PersistentVolumeClaims("kube-system").
		Delete("torpedo", &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	err = clientset.StorageV1beta1().StorageClasses().
		Delete("torpedo", &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("Deleted task: %v\n", ctx.Task.Name)
	return nil
}

func init() {
	logger = log.New(os.Stdout, "", 0)
	register("kubernetes", &k8sdriver{})
}