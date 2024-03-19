package ginkgo

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/pkg/log"
	"github.com/portworx/torpedo/tests/ginkgo/ginkgo-dsl"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

//type fieldPathReporter struct {
//	path    cmp.Path // Tracks the current path in the comparison
//	paths   []string // Accumulates paths to fields with differences
//	changes []string
//}
//
//// PushStep is called when descending into a part of the data structure.
//func (r *fieldPathReporter) PushStep(ps cmp.PathStep) {
//	r.path = append(r.path, ps)
//}
//
//// PopStep is called when ascending out of a part of the data structure.
//func (r *fieldPathReporter) PopStep() {
//	r.path = r.path[:len(r.path)-1]
//}

//// Report is called for each part of the data structure that is compared.
//func (r *fieldPathReporter) Report(rs cmp.Result) {
//	if !rs.Equal() {
//		// Join the path elements into a human-readable string and store it.
//		var pathStr []string
//		for _, step := range r.path {
//			if step, ok := step.(cmp.MapIndex); ok {
//				pathStr = append(pathStr, fmt.Sprintf("[%v]", step.Key()))
//				continue
//			}
//			if step, ok := step.(cmp.SliceIndex); ok {
//				pathStr = append(pathStr, fmt.Sprintf("[%v]", step.Key()))
//				continue
//			}
//			pathStr = append(pathStr, fmt.Sprintf(".%v", step))
//		}
//		fullPath := strings.TrimPrefix(strings.Join(pathStr, ""), ".")
//		r.paths = append(r.paths, fullPath)
//	}
//}

//func (r *fieldPathReporter) Report(rs cmp.Result) {
//	if !rs.Equal() {
//		// Convert each PathStep to a string and accumulate them.
//		var pathSteps []string
//		for _, step := range rs.Path() {
//			pathSteps = append(pathSteps, step.String())
//		}
//		pathStr := strings.Join(pathSteps, ".")
//		r.changes = append(r.changes, fmt.Sprintf("Changed: %s", pathStr))
//	}
//}
//
//var reporter = &fieldPathReporter{}

type DiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *DiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *DiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		// Use cmp.Path.String() to get the simplified path
		simplifiedPath := r.path.String()
		// Append the simplified path to the diffs slice
		r.diffs = append(r.diffs, simplifiedPath)
	}
}

func (r *DiffReporter) String() string {
	if len(r.diffs) == 0 {
		return "No differences"
	}
	// Join and format the paths of all detected changes
	return "Changed fields:\n" + strings.Join(r.diffs, "\n")
}

//func (r *DiffReporter) Report(rs cmp.Result) {
//	if !rs.Equal() {
//		vx, vy := r.path.Last().Values()
//		r.diffs = append(r.diffs, fmt.Sprintf("%#v:\n\t-: %+v\n\t+: %+v\n", r.path, vx, vy))
//	}
//}

//func (r *DiffReporter) Report(rs cmp.Result) {
//	if !rs.Equal() {
//		vx, vy := r.path.Last().Values()
//		// Simplify the path for YAML-like output
//		simplifiedPath := simplifyPath(r.path)
//		r.diffs = append(r.diffs, fmt.Sprintf("%s:\n\t-: %+v\n\t+: %+v\n", simplifiedPath, vx, vy))
//	}
//}
//
//func simplifyPath(path cmp.Path) string {
//	var simplified []string
//	for _, step := range path {
//		// Only include steps that contribute to the hierarchical path
//		if ps, ok := step.(cmp.Indirect); ok {
//			simplified = append(simplified, fmt.Sprintf("%v", ps))
//		} else if ps, ok := step.(cmp.Transform); ok {
//			simplified = append(simplified, ps.String())
//		}
//		// Extend with more cases as necessary to handle different types of steps
//	}
//	return strings.Join(simplified, "; ")
//}

//func (r *DiffReporter) Report(rs cmp.Result) {
//	if !rs.Equal() {
//		vx, vy := r.path.Last().Values()
//		pathStr := formatPath(r.path)
//		r.diffs = append(r.diffs, fmt.Sprintf("%s:\n\t-: %+v\n\t+: %+v", pathStr, vx, vy))
//	}
//}

//func formatPath(path cmp.Path) string {
//	var pathParts []string
//	for _, step := range path {
//		switch s := step.(type) {
//		case cmp.MapIndex:
//			pathParts = append(pathParts, fmt.Sprintf("[%v]", s.Key()))
//		case cmp.SliceIndex:
//			pathParts = append(pathParts, fmt.Sprintf("[%d]", s.Key()))
//		case cmp.Indirect:
//			continue // Skip indirect steps to simplify output
//		case cmp.TypeAssertion:
//			continue // Skip type assertions to focus on field names
//		default:
//			pathParts = append(pathParts, fmt.Sprint(s))
//		}
//	}
//	return strings.Join(pathParts, ".")
//}

func (r *DiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

//func (r *DiffReporter) String() string {
//	return strings.Join(r.diffs, "\n")
//}

//func (r *DiffReporter) Report(rs cmp.Result) {
//	if !rs.Equal() {
//		// Dynamically format the path to reflect its structure
//		pathStr := dynamicFormatPath(r.path)
//		// Append the formatted path to the diffs slice
//		r.diffs = append(r.diffs, pathStr)
//	}
//}
//
//// dynamicFormatPath generates a string representation of the path that led to the difference.
//func dynamicFormatPath(path cmp.Path) string {
//	var parts []string
//	for _, p := range path {
//		switch v := p.(type) {
//		case cmp.MapIndex, cmp.SliceIndex:
//			// Handle collection indices
//			parts = append(parts, fmt.Sprintf("[%v]", v))
//		default:
//			// For other types, use the string representation
//			parts = append(parts, fmt.Sprintf("%v", p))
//		}
//	}
//	return strings.Join(parts, ".")
//}
//
//func (r *DiffReporter) String() string {
//	if len(r.diffs) == 0 {
//		return "No differences"
//	}
//	// Join and format the paths of all detected changes
//	return "Changed fields:\n" + strings.Join(r.diffs, "\n")
//}

var reporter = &DiffReporter{}

var (
	initialSyncDone = false
	initialSyncLock = sync.Mutex{}
)

// To run: ginkgo --label-filter="D1I1" .

// Function to set initialSyncDone to true once cache is synced
func onInitialSyncDone(informer cache.SharedIndexInformer) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			initialSyncLock.Lock()
			defer initialSyncLock.Unlock()
			if !initialSyncDone && informer.HasSynced() {
				initialSyncDone = true
				fmt.Println("Initial sync completed.")
			}
		},
	})
}

var _ = ginkgo_dsl.NewDescribe("Describe 1", ginkgo.Label("D1"), func() {
	var (
		A = 2
	)

	ginkgo_dsl.NewIt("It 1A", ginkgo.Label("D1I1"), func() {
		//// Set up a connection to the cluster
		//
		//config, err := clientcmd.BuildConfigFromFlags("", "/Users/krishna/.kube/config")
		//if err != nil {
		//	panic(err.Error())
		//}
		//clientset, err := kubernetes.NewForConfig(config)
		//if err != nil {
		//	panic(err.Error())
		//}
		//
		//// Watching StorageClasses
		//watchSc, err := clientset.StorageV1().StorageClasses().Watch(context.TODO(), metav1.ListOptions{})
		//if err != nil {
		//	panic(err.Error())
		//}
		//defer watchSc.Stop()
		//
		//log.Infof("Watching StorageClasses...")
		//for {
		//	select {
		//	case event := <-watchSc.ResultChan():
		//		switch event.Type {
		//		case watch.Added:
		//			sc, ok := event.Object.(*storagev1.StorageClass)
		//			if !ok {
		//				log.Infof("Error: unexpected type")
		//			}
		//			fmt.Printf("StorageClass Added: %s\n", sc.Name)
		//		case watch.Modified:
		//			sc, ok := event.Object.(*storagev1.StorageClass)
		//			if !ok {
		//				log.Infof("Error: unexpected type")
		//			}
		//			fmt.Printf("StorageClass Modified: %s\n", sc.Name)
		//		case watch.Deleted:
		//			sc, ok := event.Object.(*storagev1.StorageClass)
		//			if !ok {
		//				log.Infof("Error: unexpected type")
		//			}
		//			fmt.Printf("StorageClass Deleted: %s\n", sc.Name)
		//		case watch.Error:
		//			log.Infof("Error watching")
		//		}
		//	case <-time.After(30 * time.Minute):
		//		// Stop watching after a timeout to avoid hanging indefinitely
		//		log.Infof("Timeout reached, stopping watch.")
		//		return
		//	}
		//}

		//var kubeconfig string
		//if home := homedir.HomeDir(); home != "" {
		//	kubeconfig = filepath.Join(home, ".kube", "config")
		//}
		//
		//config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		//if err != nil {
		//	log.Fatalf("Error building kubeconfig: %s", err.Error())
		//}
		//
		//clientset, err := kubernetes.NewForConfig(config)
		//if err != nil {
		//	log.Fatalf("Error building Kubernetes clientset: %s", err.Error())
		//}
		//
		//discoveryClient := clientset.Discovery()
		//apiResourceList, err := discoveryClient.ServerPreferredResources()
		//if err != nil {
		//	log.Fatalf("Error retrieving API Resources: %s", err.Error())
		//}
		//
		//for _, apiResource := range apiResourceList {
		//	gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		//	if err != nil {
		//		continue // or handle error
		//	}
		//
		//	for _, resource := range apiResource.APIResources {
		//		fmt.Printf("Name: %v, Group: %v, Version: %v\n", resource.Name, gv.Group, gv.Version)
		//	}
		//}

		// Build kubeconfig
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatalf("Error building kubeconfig: %s", err.Error())
		}

		// Initialize Kubernetes client
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("Error building Kubernetes clientset: %s", err.Error())
		}

		// Discover API resources
		discoveryClient := clientset.Discovery()
		apiResourceList, err := discoveryClient.ServerPreferredResources()
		if err != nil {
			log.Fatalf("Error retrieving API Resources: %s", err.Error())
		}

		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			log.Fatalf("Error creating dynamic client: %s", err.Error())
		}

		factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0)

		for _, apiResource := range apiResourceList {
			gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
			if err != nil {
				continue // or handle error
			}

			for _, resource := range apiResource.APIResources {

				//if len(resource.Verbs) == 0 {
				//	log.Infof("[len(resource.Verbs) == 0] Skipping resource: %s", resource.Name)
				//	continue
				//}
				//
				//if resource.Namespaced {
				//	log.Infof("[resource.Namespaced == 0] Skipping resource: %s", resource.Name)
				//	continue
				//}
				//
				//if containsString(resource.Verbs, "watch") == false {
				//	log.Infof("[containsString(resource.Verbs, \"watch\") == false] Skipping resource: %s", resource.Name)
				//	continue
				//}

				if len(resource.Verbs) == 0 || containsString(resource.Verbs, "watch") == false {
					log.Errorf("Skipping resource: %s", resource.Name)
					continue
				}
				if resource.Namespaced {
					/*if !(resource.Kind == "Pod") {
						log.Errorf("[resource.Namespaced] Skipping resource: %s", resource.Name)
						continue
					}*/
					continue
				}

				//if len(resource.Verbs) == 0 || resource.Namespaced || containsString(resource.Verbs, "watch") == false {
				//	log.Errorf("Skipping resource: %s", resource.Name)
				//	continue
				//}

				fmt.Printf("Setting up informer for: Name: %v, Group: %v, Version: %v\n", resource.Name, gv.Group, gv.Version)
				resourceGVR := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name}
				informer := factory.ForResource(resourceGVR).Informer()

				informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						//fmt.Println("Add event detected")
						//uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
						//if err != nil {
						//	fmt.Println("Error converting to unstructured object:", err)
						//	return
						//}
						//unstrObj := unstructured.Unstructured{Object: uObj}
						//fmt.Printf("Add event detected for %s in %s\n", unstrObj.GetName(), unstrObj.GetNamespace())

						initialSyncLock.Lock()
						defer initialSyncLock.Unlock()
						if initialSyncDone {
							uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
							if err != nil {
								fmt.Println("Error converting to unstructured object:", err)
								return
							}
							unstrObj := unstructured.Unstructured{Object: uObj}
							fmt.Printf("Add event detected for %s in %s\n", unstrObj.GetName(), unstrObj.GetNamespace())
						}
					},
					UpdateFunc: func(oldObj, newObj interface{}) {
						initialSyncLock.Lock()
						defer initialSyncLock.Unlock()
						if initialSyncDone {
							//uOldObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(oldObj)
							//if err != nil {
							//	fmt.Println("Error converting old object to unstructured:", err)
							//	return
							//}
							//unstrOldObj := unstructured.Unstructured{Object: uOldObj}

							uNewObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newObj)
							if err != nil {
								fmt.Println("Error converting new object to unstructured:", err)
								return
							}
							unstrNewObj := unstructured.Unstructured{Object: uNewObj}

							if unstrNewObj.GetKind() == "Node" {
								//fmt.Printf("Node update detected for %s\n", unstrNewObj.GetName())
								//compareNodes(&unstrOldObj, &unstrNewObj)
							} else {
								fmt.Printf("Update event detected for [%s/%s] in %s\n", unstrNewObj.GetKind(), unstrNewObj.GetName(), unstrNewObj.GetNamespace())

								if oldSC, ok := oldObj.(*storagev1.StorageClass); ok {
									newSC := oldObj.(*storagev1.StorageClass)
									fmt.Println("Old object is of type StorageClass")
									cmp.Equal(oldSC, newSC, cmp.Reporter(reporter))
									fmt.Println("Changed fields:", reporter.diffs)
								} else {
									fmt.Printf("Type of oldObj: %v\n", reflect.TypeOf(oldObj))
									oldSC, err := convertToStorageClass(oldObj.(*unstructured.Unstructured))
									if err != nil {
										log.Errorf("Error converting to OLD StorageClass: %s", err)
									}
									log.Infof("OLD SC LABELS [%#v]", oldSC.Labels)
									newSC, err := convertToStorageClass(newObj.(*unstructured.Unstructured))
									if err != nil {
										log.Errorf("Error converting to OLD StorageClass: %s", err)
									}
									log.Infof("NEW SC LABELS [%#v]", oldSC.ObjectMeta.Labels)
									cmp.Equal(oldSC, newSC, cmp.Reporter(reporter))
									fmt.Println("Changed fields:", reporter.String())
								}
							}

							//diff := cmp.Diff(unstrOldObj.Object, unstrNewObj.Object)
							//if diff != "" {
							//	fmt.Printf("Resource has changed: %s\n", diff)
							//}
						}
					},
					DeleteFunc: func(obj interface{}) {
						//fmt.Println("Delete event detected")
						//uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
						//if err != nil {
						//	fmt.Println("Error converting to unstructured object:", err)
						//	return
						//}
						//unstrObj := unstructured.Unstructured{Object: uObj}
						//fmt.Printf("Delete event detected for %s in %s\n", unstrObj.GetName(), unstrObj.GetNamespace())

						initialSyncLock.Lock()
						defer initialSyncLock.Unlock()
						if initialSyncDone {
							uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
							if err != nil {
								fmt.Println("Error converting to unstructured object:", err)
								return
							}
							unstrObj := unstructured.Unstructured{Object: uObj}
							fmt.Printf("Delete event detected for %s in %s\n", unstrObj.GetName(), unstrObj.GetNamespace())
						}
					},
				})
			}
		}
		stopCh := make(chan struct{})
		factory.Start(stopCh)
		// Start informers
		cachedSynced := factory.WaitForCacheSync(stopCh)
		allSynced := true
		for _, synced := range cachedSynced {
			if !synced {
				allSynced = false
				break
			}
		}

		if allSynced {
			initialSyncLock.Lock()
			initialSyncDone = true
			initialSyncLock.Unlock()
			fmt.Println("All informers have synced; watching for new changes.")
		} else {
			log.Fatalf("Failed to wait for caches to sync")
		}

		// Continue to run your application
		<-stopCh
	})

	ginkgo_dsl.NewIt("It 1B", ginkgo.Label("D1I2"), func() {
		log.Infof("[D1I2] The value of A is: %d", A)
		A = 5
	})
})

var _ = ginkgo_dsl.NewDescribe("Describe 2", ginkgo.Label("D2"), func() {
	var (
		B = 0
	)

	ginkgo_dsl.NewIt("It 2A", ginkgo.Label("D2I1"), func() {
		log.Infof("[D2I1] The value of B is: %d", B)
		B = 1
	})

	ginkgo_dsl.NewIt("It 2B", ginkgo.Label("D2I2"), func() {
		log.Infof("[D2I2] The value of B is: %d", B)
		B = 2
	})
})

var _ = ginkgo.Describe("Web Service Stateful Operations", ginkgo.Label("WebService"), func() {
	ginkgo.BeforeEach(func() {
		// This setup runs before each spec in this Describe block.
		log.Infof("[BeforeEach]")
	})

	// Using OncePerOrdered to run setup steps only once for the entire ordered group.
	ginkgo.BeforeEach(func() {
		// Setup steps that are needed once for the ordered tests below.
		log.Infof("[BeforeEach OncePerOrdered]")
	}, ginkgo.OncePerOrdered)

	ginkgo.Describe("Some ordered container", ginkgo.Ordered, func() {
		ginkgo.It("Step 1: Initialize the service", func() {
			// Test initialization
			log.Infof("[Step 1]")
		})

		ginkgo.It("Step 2: Perform operation A", func() {
			// Test operation A
			log.Infof("[Step 2]")
		})

		ginkgo.It("Step 3: Perform operation B", func() {
			// Test operation B, depends on state from A
			log.Infof("[Step 3]")
		})
	})
})

func containsString(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func compareNodes(oldNode, newNode *unstructured.Unstructured) {
	log.Infof("Old Node update detected: [%#v]", oldNode)

	log.Infof("New Node update detected: [%#v]", newNode)

	//oldLabels := oldNode.GetLabels()
	//newLabels := newNode.GetLabels()
	//
	//for key, oldValue := range oldLabels {
	//	if newValue, ok := newLabels[key]; ok {
	//		if oldValue != newValue {
	//			fmt.Printf("Label %s changed from %s to %s\n", key, oldValue, newValue)
	//		}
	//	} else {
	//		fmt.Printf("Label %s removed\n", key)
	//	}
	//}
	//for key, newValue := range newLabels {
	//	if _, ok := oldLabels[key]; !ok {
	//		fmt.Printf("Label %s added with value %s\n", key, newValue)
	//	}
	//}

	// Extend this function to compare other aspects, like annotations or status conditions
}

func convertToStorageClass(u *unstructured.Unstructured) (*storagev1.StorageClass, error) {
	var sc storagev1.StorageClass
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &sc)
	if err != nil {
		return nil, err
	}
	return &sc, nil
}
