package main

import (
	"log"
	"os"

	. "github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	ns      string
	podName string
)

func init() {
	ns = os.Getenv("NAMESPACE")
	podName = os.Getenv("POD_NAME")
}

func main() {
	switch os.Args[1] {
	case "leaderStart":
		if err := leaderStart(); err != nil {
			log.Fatalf("leaderStart failed: %v", err)
		}
	case "leaderStop":
		if err := leaderStop(); err != nil {
			log.Fatalf("leaderStop failed: %v", err)
		}
	case "liveness":
		if err := liveness(); err != nil {
			log.Fatalf("liveness failed: %v", err)
		}
	case "readiness":
		if err := readiness(); err != nil {
			log.Fatalf("readiness failed: %v", err)
		}
	case "postStart":
		if err := postStart(); err != nil {
			log.Fatalf("postStart failed: %v", err)
		}
	case "preStop":
		if err := preStop(); err != nil {
			log.Fatalf("postStop failed: %v", err)
		}
	default:
		log.Fatalf("Usage: %s leaderStart|leaderStop|liveness|readiness|postStart|preStop", os.Args[0])
	}
}

// TODO
func leaderStart() error {
	return nil
}

func leaderStop() error {
	return PatchRoleLabelTo(myself(string(Follower)))
}

func liveness() error {
	return XenonPingMyself()
}

func readiness() error {
	role := GetRole()
	if role != string(Leader) {
		return PatchRoleLabelTo(myself(role))
	}
	return nil
}

// TODO
func postStart() error {
	return nil
}

// TODO
func preStop() error {
	return nil
}

func myself(role string) MySQLNode {
	return MySQLNode{
		PodName:   podName,
		Namespace: ns,
		Role:      role,
	}
}
