package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	v1alhpa1 "github.com/radondb/radondb-mysql-kubernetes/api/v1alpha1"
)

func Create(hostPath string, cluster string, hostName string, capacity string, backupImage string, nfsServerImage string) error {
	cl, err := getClient()
	if err != nil {
		return err
	}

	c, err := getMysqlClusterByName(cl, cluster)
	if err != nil {
		return err
	}
	namespace := c.Namespace

	if hostName == "" {
		hostName, err = getClusterLeader(c)
		if err != nil {
			return err
		}
	}

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	_, err = clientset.StorageV1().StorageClasses().Create(context.TODO(), stc(), metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		log.Println(err)
	}

	_, err = clientset.CoreV1().PersistentVolumes().Create(context.TODO(), pv(hostPath, capacity), metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		log.Println(err)
	}

	_, err = clientset.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), pvc(), metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		log.Println(err)
	}

	_, err = clientset.CoreV1().ReplicationControllers(namespace).Create(context.TODO(), rc(nfsServerImage), metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		log.Println(err)
	}

	service, err := clientset.CoreV1().Services(namespace).Create(context.TODO(), svc(), metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		} else {
			log.Println(err)
			service, err = clientset.CoreV1().Services(namespace).Get(context.TODO(), "radondb-nfs-server", metav1.GetOptions{})
			if err != nil {
				return err
			}
		}
	}

	err = cl.Create(context.TODO(), backup(namespace, cluster, hostName, service.Spec.ClusterIP, backupImage))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		log.Println(err)
		return nil
	}
	log.Println("Backup created successfully")

	return nil
}

func getMysqlClusterByName(c client.Client, name string) (*v1alhpa1.MysqlCluster, error) {
	clusterList := &v1alhpa1.MysqlClusterList{}
	err := c.List(context.TODO(), clusterList)
	if err != nil {
		return nil, err
	}
	cluster := v1alhpa1.MysqlCluster{}
	for i := 0; i < len(clusterList.Items); i++ {
		if clusterList.Items[i].Name == name {
			cluster = clusterList.Items[i]
			return &cluster, nil
		}
	}
	return nil, fmt.Errorf("failed to find cluster: %s", name)
}

func getClient() (client.Client, error) {
	var scheme = runtime.NewScheme()
	_ = v1alhpa1.SchemeBuilder.AddToScheme(scheme)

	cl, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return cl, nil
}

func getClientSet() (*kubernetes.Clientset, error) {
	kubeConfig := getKubeconfig()
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config: %v", err)
	}
	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	return clientset, nil
}

func getKubeconfig() string {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		kubeconfig = filepath.Join(u.HomeDir, ".kube", "config")
		if _, err := os.Stat(kubeconfig); err != nil && !os.IsNotExist(err) {
			kubeconfig = ""
		}
	}
	return kubeconfig
}
