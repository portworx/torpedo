package utilities

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jinzhu/copier"
	"github.com/portworx/sched-ops/k8s/kubevirt"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/pkg/log"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/scheduler"
	corev1 "k8s.io/api/core/v1"
)

type AppInfo struct {
	StartDataSupport bool
	User             string
	Password         string
	Port             int
	NodePort         int
	DBName           string
	Hostname         string
	AppType          string
	Namespace        string
	IPAddress        string
}

type AwsCompatibleStorageClient struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
}

type AwsStorageClient struct {
	AccessKey string
	SecretKey string
	Region    string
}

type AzureStorageClient struct {
	AccountName string
	AccountKey  string
}

type GcpStorageClient struct {
	ProjectId string
}

const (
	svcAnnotationKey                = "startDataSupported"
	userAnnotationKey               = "username"
	passwordAnnotationKey           = "password"
	databaseAnnotationKey           = "databaseName"
	portAnnotationKey               = "port"
	appTypeAnnotationKey            = "appType"
	defaultCmdTimeout               = 20 * time.Second
	defaultCmdRetryInterval         = 5 * time.Second
	defaultKubeconfigMapForKubevirt = "kubevirt-creds"
	postgresqlStressImage           = "portworx/torpedo-pgbench:pdsloadTest"
)

// RandomString generates a random lowercase string of length characters.
func RandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = letters[rand.Intn(len(letters))]
	}
	randomString := string(randomBytes)
	return randomString
}

func (azureObj *AzureStorageClient) CreateAzureBucket(bucketName string) error {
	log.Debugf("Creating azure bucket with name [%s]", bucketName)
	urlStr := fmt.Sprintf("https://%s.blob.core.windows.net/%s", azureObj.AccountName, bucketName)
	log.Infof("Create container url %s", urlStr)
	// Create a ContainerURL object that wraps a soon-to-be-created container's URL and a default pipeline.
	u, _ := url.Parse(urlStr)
	credential, err := azblob.NewSharedKeyCredential(azureObj.AccountName, azureObj.AccountKey)
	if err != nil {
		return fmt.Errorf("Failed to create shared key credential [%v]", err)
	}

	containerURL := azblob.NewContainerURL(*u, azblob.NewPipeline(credential, azblob.PipelineOptions{}))
	ctx := context.Background()

	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		return fmt.Errorf("Failed to create container. Error: [%v]", err)
	}
	return nil
}

func (awsObj *AwsStorageClient) CreateS3Bucket(bucketName string) error {
	log.Debugf("Creating s3 bucket with name [%s]", bucketName)
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(awsObj.Region),
			Credentials: credentials.NewStaticCredentials(awsObj.AccessKey, awsObj.SecretKey, ""),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to initialize new session: %v", err)
	}

	client := s3.New(sess)
	bucketObj, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if (aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou) || (aerr.Code() == s3.ErrCodeBucketAlreadyExists) {
				log.Infof("Bucket: %v ,already exist.", bucketName)
				return nil
			} else {
				return fmt.Errorf("couldn't create bucket: %v", err)
			}

		}
	}

	log.Infof("[AWS]Successfully created the bucket. Info: %v", bucketObj)
	return nil
}

func (awsObj *AwsCompatibleStorageClient) CreateS3CompBucket(bucketName string) error {
	log.Debugf("Creating s3 bucket with name [%s]", bucketName)
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:         aws.String(awsObj.Endpoint),
			Region:           aws.String(awsObj.Region),
			Credentials:      credentials.NewStaticCredentials(awsObj.AccessKey, awsObj.SecretKey, ""),
			S3ForcePathStyle: aws.Bool(true),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to initialize new session: %v", err)
	}

	client := s3.New(sess)
	bucketObj, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if (aerr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou) || (aerr.Code() == s3.ErrCodeBucketAlreadyExists) {
				log.Infof("Bucket: %v ,already exist.", bucketName)
				return nil
			} else {
				return fmt.Errorf("couldn't create bucket: %v", err)
			}

		}
	}

	log.Infof("[AWS]Successfully created the bucket. Info: %v", bucketObj)
	return nil
}

func (awsObj *AwsCompatibleStorageClient) ListFolders(after time.Time, bucketName string) ([]string, error) {
	var folders []string
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Endpoint:         aws.String(awsObj.Endpoint),
			Region:           aws.String(awsObj.Region),
			Credentials:      credentials.NewStaticCredentials(awsObj.AccessKey, awsObj.SecretKey, ""),
			S3ForcePathStyle: aws.Bool(true),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new session: %v", err)
	}
	svc := s3.New(sess)
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("error while listing the folders in S3 bucket: %v", err)
	}
	for _, obj := range resp.Contents {
		log.Infof("Last modified status: %v", obj.LastModified.After(after))
		if obj.LastModified.After(after) {
			parts := strings.Split(*obj.Key, "/")
			log.Infof("Parts: %v", parts)
			if len(parts) >= 1 {
				folders = append(folders, parts[0])
			}
		}
	}
	return folders, nil
}

func ConvertJsonFileToString(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	// Read the file contents
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Convert byte slice to string
	jsonString := string(byteValue)

	return jsonString, nil

}

// CreateHostNameForApp creates a hostname using service name and namespace
func CreateHostNameForApp(serviceName string, nodePort int32, namespace string) (string, error) {
	var hostname string

	if nodePort != 0 {
		k8sCore := core.Instance()
		nodes, err := k8sCore.GetNodes()
		if err != nil {
			return "", err
		}
		// hostname = nodes.Items[0].Name
		hostname = nodes.Items[0].Status.Addresses[0].Address
	} else {
		hostname = fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
	}

	return hostname, nil
}

// ExtractConnectionInfo Extracts the connection information from the service yaml
func ExtractConnectionInfo(ctx *scheduler.Context, context context.Context) (AppInfo, error) {

	// TODO: This needs to be enhanced to support multiple application in one ctx
	var appInfo AppInfo

	for _, specObj := range ctx.App.SpecList {
		if obj, ok := specObj.(*kubevirtv1.VirtualMachine); ok {
			k8sKubevirt := kubevirt.Instance()
			appInfo.Namespace = obj.Namespace
			log.Infof("%+v", obj.Annotations)
			vmInstance, err := k8sKubevirt.GetVirtualMachineInstance(context, obj.Name, obj.Namespace)
			if err != nil {
				return appInfo, err
			}
			if svcAnnotationValue, ok := obj.Annotations[svcAnnotationKey]; ok {
				appInfo.StartDataSupport = svcAnnotationValue == "true"
				if !appInfo.StartDataSupport {
					break
				}
			} else {
				appInfo.StartDataSupport = false
				break
			}
			if userAnnotationValue, ok := obj.Annotations[userAnnotationKey]; ok {
				appInfo.User = userAnnotationValue
			} else {
				return appInfo, fmt.Errorf("Username not found")
			}
			if appTypeAnnotationValue, ok := obj.Annotations[appTypeAnnotationKey]; ok {
				appInfo.AppType = appTypeAnnotationValue
			} else {
				return appInfo, fmt.Errorf("AppType not found")
			}
			appInfo.Hostname = obj.Name

			appInfo.IPAddress = vmInstance.Status.Interfaces[0].IP
			cm, err := core.Instance().GetConfigMap(defaultKubeconfigMapForKubevirt, "default")
			if err != nil {
				return appInfo, err
			}
			appInfo.Password = cm.Data[obj.Name]
			return appInfo, nil

		}
		if obj, ok := specObj.(*corev1.Service); ok {
			appInfo.Namespace = obj.Namespace
			// TODO: This needs to be fetched from spec once CloneAppContextAndTransformWithMappings is fixed
			svc, err := core.Instance().GetService(obj.Name, obj.Namespace)
			if err != nil {
				return appInfo, err
			}
			nodePort := svc.Spec.Ports[0].NodePort
			hostname, err := CreateHostNameForApp(obj.Name, nodePort, obj.Namespace)
			if err != nil {
				return appInfo, fmt.Errorf("Some error occurred while generating hostname")
			}
			appInfo.Hostname = hostname
			appInfo.NodePort = int(nodePort)
			if svcAnnotationValue, ok := obj.Annotations[svcAnnotationKey]; ok {
				appInfo.StartDataSupport = svcAnnotationValue == "true"
				if !appInfo.StartDataSupport {
					continue
				}
			} else {
				appInfo.StartDataSupport = false
				continue
			}
			if userAnnotationValue, ok := obj.Annotations[userAnnotationKey]; ok {
				appInfo.User = userAnnotationValue
			} else {
				return appInfo, fmt.Errorf("Username not found")
			}
			if appTypeAnnotationValue, ok := obj.Annotations[appTypeAnnotationKey]; ok {
				appInfo.AppType = appTypeAnnotationValue
			} else {
				return appInfo, fmt.Errorf("AppType not found")
			}
			if passwordAnnotationValue, ok := obj.Annotations[passwordAnnotationKey]; ok {
				appInfo.Password = passwordAnnotationValue
			} else {
				return appInfo, fmt.Errorf("Password not found")
			}
			if portAnnotationValue, ok := obj.Annotations[portAnnotationKey]; ok {
				appInfo.Port, _ = strconv.Atoi(portAnnotationValue)
			}
			if databaseAnnotationValue, ok := obj.Annotations[databaseAnnotationKey]; ok {
				appInfo.DBName = databaseAnnotationValue
			}
		}
	}

	if appInfo.StartDataSupport {
		log.Infof("Running sync on all pods in namespace [%s]", appInfo.Namespace)
		syncData(appInfo.Namespace)
	}

	return appInfo, nil
}

// RunCmdGetOutputOnNode runs the command on a particular node and returns output
func RunCmdGetOutputOnNode(cmd string, n node.Node, nodeDriver node.Driver) (string, error) {
	output, err := nodeDriver.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         defaultCmdTimeout,
		TimeBeforeRetry: defaultCmdRetryInterval,
		Sudo:            true,
	})
	if err != nil {
		log.Warnf("failed to run cmd: %s. err: %v", cmd, err)
	}
	return output, err
}

// GetEnv gets environment variable and fall back to default value if not found
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Infof("Unable to find [%v] in the env variables, falling back to [%s]", fallback)
		return fallback
	}
	log.Infof("ENV values %s", value)
	return value
}

// CopyStruct copies one struct to another and raise error if failed
func CopyStruct(fromValue interface{}, toValue interface{}) error {
	// log.Infof("Copying from [%+v]", fromValue)
	// log.Infof("Copying to [%+v]", toValue)
	err := copier.CopyWithOption(toValue, fromValue, copier.Option{CaseSensitive: false})
	return err
}

// RandString generates random string
func RandString(length int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// syncData will execute 'sync' command on pod post start
func syncData(namespace string) {
	var k8sCore = core.Instance()
	allPods, err := k8sCore.GetPods(namespace, make(map[string]string))

	for _, pod := range allPods.Items {
		_, err = k8sCore.RunCommandInPod([]string{"sync"}, pod.Name, pod.Spec.Containers[0].Name, namespace)
		if err != nil {
			log.Warnf("Some error occurred while running Sync - Error - [%s]", err.Error())
		} else {
			log.Infof("Sync ran successfully")
		}
	}

}

// DeleteElementFromSlice deletes an element from a slice
func DeleteElementFromSlice(slice []string, element string) ([]string, error) {
	// Find the index of the element to delete
	index := -1
	for i, v := range slice {
		if v == element {
			index = i
			break
		}
	}

	// If element not found, return the original slice
	if index == -1 {
		return slice, fmt.Errorf("[%s] not found in [%s]", element, slice)
	}

	// Delete the element by slicing the original slice
	return append(slice[:index], slice[index+1:]...), nil
}

func GetBasePodName(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "-")
	}
	return podName
}

// ParseInterfaceAndGetDetails takes interface as input and checks for the particular type and extracts the host and port information
// Returns the host and port as dnsEndpoints
func ParseInterfaceAndGetDetails(connectionDetails interface{}, dataServiceName string) (string, error) {
	var (
		defaultPort string
		dsNode      string
	)

	connDetailsMap, ok := connectionDetails.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Error: connectionDetails is not of type map[string]interface{}")
	}

	nodesInterface, ok := connDetailsMap["instances"]
	if !ok {
		return "", fmt.Errorf("Error: instances not found in connectionDetails")
	}
	nodes, ok := nodesInterface.([]interface{})
	if !ok {
		return "", fmt.Errorf("Error: instances is not of type []interface{}")
	}

	log.Debugf("Available nodes")
	for _, nodeInterface := range nodes {
		node, err := ConvertInterfacetoString(nodeInterface)
		if err != nil {
			return "", err
		}
		log.Debugf("[%s]", node)
		if strings.Contains(node, "vip") {
			dsNode = node
		}
	}

	// Extract ports from the map
	defaultPort, err := GetDeploymentPort(connectionDetails, dataServiceName)
	if err != nil {
		return "", fmt.Errorf("Unable to fetch port details - [%s]", err.Error())
	}

	if dsNode == "" || string(defaultPort) == "" {
		return "", fmt.Errorf("Node or Port value is empty..\n")
	}

	dnsEndPoint := dsNode + ":" + defaultPort

	return dnsEndPoint, nil
}

// GetDeploymentPort fetches the port for given deployment
func GetDeploymentPort(connectionDetails interface{}, dataServiceName string) (string, error) {
	var (
		defaultPort string
	)

	connDetailsMap, ok := connectionDetails.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Error: connectionDetails is not of type map[string]interface{}")
	}

	// Extract ports from the map
	portsInterface, ok := connDetailsMap["ports"]
	if !ok {
		return "", fmt.Errorf("Error: ports not found in connectionDetails")
	}
	ports, ok := portsInterface.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Error: ports is not of type map[string]interface{}")
	}

	log.Debugf("Available ports")
	for portName, portInterface := range ports {
		port, err := ConvertInterfacetoString(portInterface)
		if err != nil {
			return "", err
		}
		log.Debugf("[%s]:[%s]", portName, port)
		switch strings.ToLower(dataServiceName) {
		case "postgresql":
			if portName == "postgresql" {
				defaultPort = port
			}
		case "cassandra":
			if portName == "cql" {
				defaultPort = port
			}
		case "couchbase":
			if portName == "rest" {
				defaultPort = port
			}
		case "redis":
			if portName == "client" {
				defaultPort = port
			}
		case "rabbitmq":
			if portName == "amqp" {
				defaultPort = port
			}
		case "kafka":
			if portName == "client" {
				defaultPort = port
			}
		case "elasticsearch":
			if portName == "rest" {
				defaultPort = port
			}
		case "mongodb":
			if portName == "mongos" {
				defaultPort = port
			}
		case "consul":
			if portName == "http" {
				defaultPort = port
			}
		case "mysql":
			if portName == "mysql-router" {
				defaultPort = port
			}
		case "sqlserver":
			if portName == "client" {
				defaultPort = port
			}
		case "neo4j":
			if portName == "bolt" {
				defaultPort = port
			}
		}
	}

	return defaultPort, nil
}

func CreatePostgresqlWorkload(dnsEndpoint, pdsPassword, scaleFactor, iterations, deploymentName, namespace string) (*appsv1.Deployment, error) {
	var replicas int32 = 1
	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: deploymentName + "-",
			Namespace:    namespace,
		},
		Spec: appsv1.DeploymentSpec{
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
							Name:    "pgbench",
							Image:   postgresqlStressImage,
							Command: []string{"/pgloadgen.sh"},
							Args:    []string{dnsEndpoint, pdsPassword, scaleFactor, iterations},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
	deployment, err := k8sApps.CreateDeployment(deploymentSpec, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("An Error Occured while creating deployment %v", err)
		return nil, err
	}
	err = k8sApps.ValidateDeployment(deployment, timeOut, timeInterval)
	if err != nil {
		log.Errorf("An Error Occured while validating the pod %v", err)
		return nil, err
	}

	//TODO: Remove static sleep and verify the injected data
	time.Sleep(2 * time.Minute)

	return deployment, err
}

func ConvertInterfacetoString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported type: %T", value)
	}
}

// Contains checks if a string slice contains a specific string
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// GetAndExpectStringEnvVar parses a string from env variable.
func GetAndExpectStringEnvVar(varName string) string {
	varValue := os.Getenv(varName)
	return varValue
}
