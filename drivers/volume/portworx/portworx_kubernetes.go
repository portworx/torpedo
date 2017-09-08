

package volume

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"io/ioutil"
	

	"github.com/libopenstorage/openstorage/api"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/volume"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/pkg/api/v1"
	dv1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	//av1 "k8s.io/client-go/pkg/apis/rbac/v1beta1"

)

type portworxK8s struct {
	
	clusterManager cluster.Cluster
	volDriver      volume.VolumeDriver
}

var logger *log.Logger

func convertYamltoStruct(yaml string) runtime.Object {
	
		d := scheme.Codecs.UniversalDeserializer()
		obj, _, err := d.Decode([]byte(yaml), nil, nil)
		if err != nil {
			log.Fatalf("could not decode yaml: %s\n%s", yaml, err)
		}
	
		return obj
}

func kconnect(ip string)(*kubernetes.Clientset,error){
	
	 token, caData, err := parseConnectionParams()
	 if err != nil {
		 logger.Fatalf("failed to parse configuration parameters: %s", err)
	 }
 
 
	 config := &rest.Config{
		 Host:            "https://"+ip+":6443",
		 BearerToken:     token,
		 TLSClientConfig: rest.TLSClientConfig{CAData: caData},
	 }
 
 
	 clientset, err := kubernetes.NewForConfig(config)
 
	 
	 return clientset,err
 }

func parseConnectionParams() (token string, caData []byte, err error) {
	

	token = os.Getenv("tokenEnvVar")
   // fmt.Println(token)
	caFile := os.Getenv("caFileEnvVar")
	if len(caFile) > 0 {
		caData, err = ioutil.ReadFile(caFile)
	}

	return token, caData, err
}



func (d *portworxK8s) String() string {
	return "pxd_k8s"
}

func (d *portworxK8s) Init() error {
	log.Printf("Using the Portworx volume portworx.\n")

	n := "127.0.0.1"
	if len(nodes) > 0 {
		n = nodes[1]
	}

	clnt, err := clusterclient.NewClusterClient("http://"+n+":9001", "v1")
	if err != nil {
		return err
	}
	d.clusterManager = clusterclient.ClusterManager(clnt)

	clnt, err = volumeclient.NewDriverClient("http://"+n+":9001", "pxd", "","")
	if err != nil {
		return err
	}
	d.volDriver = volumeclient.VolumeDriver(clnt)

	cluster, err := d.clusterManager.Enumerate()
	if err != nil {
		return err
	}

	log.Printf("The following Portworx nodes are in the cluster:\n")
	for _, n := range cluster.Nodes {
		log.Printf(
			"\tNode ID: %v\tNode IP: %v\tNode Status: %v\n",
			n.Id,
			n.DataIp,
			n.Status,
		)
	}

	return nil
}

func (d *portworxK8s) RemoveVolume(name string) error {
	locator := &api.VolumeLocator{}

	volumes, err := d.volDriver.Enumerate(locator, nil)
	if err != nil {
		return err
	}

	for _, v := range volumes {
		if v.Locator.Name == name {
			// First unmount this volume at all mount paths...
			for _, path := range v.AttachPath {
				if err = d.volDriver.Unmount(v.Id, path); err != nil {
					err = fmt.Errorf(
						"Error while unmounting %v at %v because of: %v",
						v.Id,
						path,
						err,
					)
					log.Printf("%v", err)
					return err
				}
			}

			if err = d.volDriver.Detach(v.Id,true); err != nil {
				err = fmt.Errorf(
					"Error while detaching %v because of: %v",
					v.Id,
					err,
				)
				log.Printf("%v", err)
				return err
			}

			if err = d.volDriver.Delete(v.Id); err != nil {
				err = fmt.Errorf(
					"Error while deleting %v because of: %v",
					v.Id,
					err,
				)
				log.Printf("%v", err)
				return err
			}

			log.Printf("Succesfully removed Portworx volume %v\n", name)

			return nil
		}
	}

	return nil
}


func (d *portworxK8s) Stop(ip string) error{
 
	clientset,err:=kconnect(ip)
	if(err!=nil){
		log.Println(err)
		return err
	}

   
	fal:=false
	
    daemonDelete :=&metav1.DeleteOptions{
	  OrphanDependents:   &fal,
	}
	err=clientset.DaemonSets("kube-system").Delete("portworx",daemonDelete)
	if(err!=nil){
		log.Println(err)
		return err
	}

	log.Println("Portworx Driver Down")

	return nil
}



func (d *portworxK8s) Start(ip string) error{
	clientset,err:=kconnect(ip)
	if(err!=nil){
		log.Println(err)
		return err
	}	

	dat, err := ioutil.ReadFile("daemonset.yaml")
	if err != nil {
		log.Fatalf(err.Error())
	}

	//hack
  obj1 := convertYamltoStruct(fmt.Sprintf("%s", dat))
		
	
	dms := obj1.(*dv1.DaemonSet)
	dms, err = clientset.DaemonSets("kube-system").Create(dms)	
	
	log.Printf("Portworx Driver is running")
	return nil
}

func (d *portworxK8s) WaitStart(ip string) error {
	// Wait for Portworx to become usable.
	status, _ := d.clusterManager.NodeStatus()
	for i := 0; status != api.Status_STATUS_OK; i++ {
		if i > 60 {
			return fmt.Errorf(
				"Portworx did not start up in time: Status is %v",
				status,
			)
		}

		time.Sleep(1 * time.Second)
		status, _ = d.clusterManager.NodeStatus()
	}

	return nil
}


func init() {
	nodes = strings.Split(os.Getenv("CLUSTER_NODES"), ",")
  logger = log.New(os.Stdout, "", 0)

	register("pxd_k8s", &portworxK8s{})
}

