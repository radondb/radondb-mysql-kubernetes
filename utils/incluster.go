package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type raftStatus struct {
	Leader string   `json:"leader"`
	State  string   `json:"state"`
	Nodes  []string `json:"nodes"`
}

type MySQLNode struct {
	PodName   string
	Namespace string
	Role      string
}

func GetClientSet() (*kubernetes.Clientset, error) {
	// Creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %v", err)
	}
	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	return clientset, nil
}

func PatchRoleLabelTo(n MySQLNode) error {
	// Creates the clientset
	clientset, err := GetClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}
	patch := fmt.Sprintf(`{"metadata":{"labels":{"role":"%s"}}}`, n.Role)
	_, err = clientset.CoreV1().Pods(n.Namespace).Patch(context.TODO(), n.PodName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch pod role label: %v", err)
	}
	return nil
}

func XenonPingMyself() error {
	args := []string{"xenon", "ping"}
	cmd := exec.Command("xenoncli", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exec xenoncli xenon ping: %v", err)
	}
	return nil
}

func GetRaftStatus() *raftStatus {
	args := []string{"raft", "status"}
	cmd := exec.Command("xenoncli", args...)
	res, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed to exec xenoncli raft status: %v", err)
	}
	raftStatus := raftStatus{}
	if err := json.Unmarshal(res, &raftStatus); err != nil {
		log.Fatalf("failed to unmarshal raft status: %v", err)
	}
	return &raftStatus
}

func GetRole() string {
	return GetRaftStatus().State
}
