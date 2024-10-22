package utilities

import (
	"fmt"
	"github.com/pure-px/torpedo/pkg/log"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"strconv"
	"strings"
	"time"
)

const (
	postgresql   = "PostgreSQL"
	mysql        = "MySQL"
	pdsTpccImage = "portworx/torpedo-tpcc-automation:v1"
)

const (
	defaultCommandRetry  = 5 * time.Second
	defaultRetryInterval = 10 * time.Minute
)

// This module creates TPCC Schema for a given Deployment and then Runs TPCC Workload
func RunTpccWorkload(dbUser string, pdsPassword string, dnsEndpoint string, dbName string,
	timeToRun string, numOfThreads string, numOfCustomers string, numOfWarehouses string,
	deploymentName string, namespace string, dataServiceName string) error {
	var fileToRun string
	if dataServiceName == postgresql {
		dbName = "pds"
		fileToRun = "tpcc-pg-run.sh" // file to run in case of Postgres workload
	}
	if dataServiceName == mysql {
		err := SetupMysqlDatabaseForTpcc(dbUser, pdsPassword, dnsEndpoint, namespace)
		if err != nil {
			return err
		}
		dbName = "tpcc"
		fileToRun = "tpcc-mysql-run.sh" // File to run in case of MySQL workload
	}
	if dbUser == "" {
		dbUser = "pds"
	}
	if timeToRun == "" {
		timeToRun = "120" // Default time to run is 2 minutes
	}
	if numOfThreads == "" {
		numOfThreads = "64" // Default threads is 64
	}
	if numOfCustomers == "" {
		numOfCustomers = "2" // Default number of customer and districts is 4
	}
	if numOfWarehouses == "" {
		numOfWarehouses = "1" // Default number of warehouses to simulate is 2
	}
	// Create a Deployment to Prepare and Run TPCC Workload
	var replicas int32 = 1
	deploymentSpec := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: deploymentName + "-",
			Namespace:    namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": deploymentName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "tpcc-run",
							Image: pdsTpccImage,
							Command: []string{"/bin/sh", "-C", fileToRun, dbUser, pdsPassword, dnsEndpoint, dbName,
								timeToRun, numOfThreads, numOfCustomers, numOfWarehouses, "run"},
							WorkingDir: "/sysbench-tpcc",
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "tpcc-prepare",
							Image: pdsTpccImage,
							Command: []string{"/bin/sh", "-C", fileToRun, dbUser, pdsPassword, dnsEndpoint, dbName,
								timeToRun, numOfThreads, numOfCustomers, numOfWarehouses, "prepare"},
							WorkingDir:      "/sysbench-tpcc",
							ImagePullPolicy: corev1.PullAlways,
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
	log.InfoD("Going to Trigger TPCC Workload for the Deployment")
	deployment, err := k8sApps.CreateDeployment(deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("An Error Occured while creating deployment %v", err)
		return fmt.Errorf("An Error Occured while creating deployment %s", err.Error())
	}

	timeAskedToRun, err := strconv.Atoi(timeToRun)
	flag := false
	// Hard sleep for 10 seconds for deployment to come up
	time.Sleep(10 * time.Second)
	var newPods []corev1.Pod
	for i := 1; i <= 200; i++ {
		newPodList, _ := GetPods(namespace)
		newPods = append(newPods, newPodList.Items...)
		for _, pod := range newPods {
			if strings.Contains(pod.Name, deployment.Name) {
				log.InfoD("Will check for status of Init Container Once......")
				for _, c := range pod.Status.InitContainerStatuses {
					if c.State.Terminated != nil {
						flag = true
					}
				}
			}
		}
		if flag {
			log.InfoD("TPCC Schema Prepared successfully. Moving ahead to run the TPCC Workload now.....")
			break
		} else {
			log.InfoD("Init Container is still running means TPCC Schema is being prepared. Will wait for further 30 Seconds.....")
			time.Sleep(30 * time.Second)
		}
	}
	if !flag {
		log.Errorf("TPCC Schema couldn't be prepared in 100 minutes. Timing Out. Please check manually.")
		return fmt.Errorf("TPCC Schema couldn't be prepared in 100 minutes. Timing Out. Please check manually.")
	}
	flag = false
	for i := 1; i <= int((timeAskedToRun+300)/60); i++ {
		newPodList, _ := GetPods(namespace)
		newPods = append(newPods, newPodList.Items...)
		for _, pod := range newPods {
			if strings.Contains(pod.Name, deployment.Name) {
				log.InfoD("Waiting for TPCC Workload Container to finish")
				for _, c := range pod.Status.ContainerStatuses {
					if int32(c.RestartCount) != 0 {
						flag = true
						if c.State.Terminated != nil && c.State.Terminated.ExitCode != 0 && c.State.Terminated.Reason != "Completed" {
							log.Errorf("Something went wrong and Run Container Exited abruptly. Leaving the TPCC deployment as is - pls check manually")
							log.InfoD("Printing TPCC Deployment Describe Status here .....")
							depStatus, err := k8sApps.DescribeDeployment(deployment.Name, namespace)
							if err != nil {
								log.Errorf("Could not print TPCC Deployment status due to some reason. Please check manually.")
								return fmt.Errorf("Could not print TPCC Deployment status due to some reason. Please check manually.")
							}
							log.InfoD("%+v\n", *depStatus)
							return fmt.Errorf("TPCC Deployment Status: %v", *depStatus)
						}
						break
					}
				}
			}
		}
		if flag {
			log.InfoD("TPCC Workload run finished. Finishing this Test Case")
			break
		} else {
			log.InfoD("TPCC Workload is still running. Will wait for further 1 minute to check again.....")
			time.Sleep(1 * time.Minute)
		}
	}
	log.InfoD("Will delete TPCC Worklaod Deployment now.....")
	k8sApps.DeleteDeployment(deployment.Name, namespace)
	return nil
}

// This module sets up MySQL Database for Running TPCC. There is some specific requirement that needs to be
// done for MySQL before running MySQL.
func SetupMysqlDatabaseForTpcc(dbUser string, pdsPassword string, dnsEndpoint string, namespace string) error {
	log.InfoD("Trying to configure Mysql deployment for TPCC Workload")
	if dbUser == "" {
		dbUser = "pds"
	}
	podSpec := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "configure-mysql-",
			Namespace:    namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "configure-mysql",
					Image:           pdsTpccImage,
					Command:         []string{"/bin/sh", "-C", "setup-mysql-for-tpcc.sh", dbUser, pdsPassword, dnsEndpoint},
					WorkingDir:      "/sysbench-tpcc",
					ImagePullPolicy: corev1.PullAlways,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	configureMysqlPod, err := k8sCore.CreatePod(podSpec)
	if err != nil {
		log.Errorf("An Error Occured while creating %v", err)
		return fmt.Errorf("Error occurred while creating configure-mysql pod: [%s]", err.Error())
	}
	configureMysqlPodName := configureMysqlPod.ObjectMeta.Name
	//Static sleep to let DB changes settle in
	time.Sleep(20 * time.Second)

	err = wait.Poll(defaultCommandRetry, defaultRetryInterval, func() (bool, error) {
		pod, err := k8sCore.GetPodByName(configureMysqlPodName, namespace)
		if err != nil {
			return false, err
		}
		if k8sCore.IsPodRunning(*pod) {
			log.Infof("Looks like configure-mysql pod is running. Waiting for it to complete.")
			return false, nil
		} else {
			log.Infof("configure-mysql pod is now not running. Moving ahead.")
			return true, nil
		}
	})
	if err != nil {
		return fmt.Errorf("Error occurred while creating configure-mysql pod: [%s]", err.Error())
	}
	var newPods []corev1.Pod
	newPodList, err := GetPods(namespace)
	if err != nil {
		return fmt.Errorf("Error occurred while creating configure-mysql pod: [%s]", err.Error())
	}
	//reinitializing the pods
	newPods = append(newPods, newPodList.Items...)

	// Validate if MySQL pod is configured successfully or not for running TPCC
	for _, pod := range newPods {
		if strings.Contains(pod.Name, configureMysqlPodName) {
			log.InfoD("pds system pod name %v", pod.Name)
			for _, c := range pod.Status.ContainerStatuses {
				if c.State.Terminated != nil {
					if c.State.Terminated.ExitCode == 0 && c.State.Terminated.Reason == "Completed" {
						log.InfoD("Successfully Configured Mysql for TPCC Run. Exiting")
						k8sCore.DeletePod(pod.Name, namespace, true)
						return nil
					} else {
						k8sCore.DeletePod(pod.Name, namespace, true)
					}
				} else {
					log.Infof("configure-mysql pod seems to be still running. This is not expected.")
					return fmt.Errorf("Error occurred while creating configure-mysql pod")
				}
			}
		}
	}
	return fmt.Errorf("Error occurred while creating configure-mysql pod")
}
