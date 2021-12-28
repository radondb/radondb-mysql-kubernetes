# How to run e2e test

## Prerequisites

Prepare a client connected to K8S.

## Hands-on Lab

### Step 1: Configure environment variables

```
export KUBECONFIG=$HOME/.kube/config
```

### Step 2: Run test

```
make e2e-local
```
Example output of simplecase:
```
$ make e2e-local
=== RUN   TestE2E
STEP: Creating framework with timeout: 1200
Running Suite: MySQL Operator E2E Suite
=======================================
Random Seed: 1640785115 - Will randomize all specs
Will run 1 of 1 specs

Namespece test 
  test list namespace
  /home/runkecheng/goCode/src/radondb-mysql-kubernetes/test/e2e/simplecase/list_namespace.go:38
[BeforeEach] Namespece test
  /home/runkecheng/goCode/src/radondb-mysql-kubernetes/test/e2e/framework/framework.go:62
STEP: creating a kubernetes client
STEP: create a namespace api object (e2e-mc-1-cnkbs)
[BeforeEach] Namespece test
  /home/runkecheng/goCode/src/radondb-mysql-kubernetes/test/e2e/simplecase/list_namespace.go:34
STEP: before each
[It] test list namespace
  /home/runkecheng/goCode/src/radondb-mysql-kubernetes/test/e2e/simplecase/list_namespace.go:38
default
kube-public
kube-system
kubesphere-controls-system
kubesphere-devops-system
kubesphere-devops-worker
kubesphere-monitoring-federated
kubesphere-monitoring-system
kubesphere-system
radondb-mysql
radondb-mysql-kubernetes-system
[AfterEach] Namespece test
  /home/runkecheng/goCode/src/radondb-mysql-kubernetes/test/e2e/framework/framework.go:63
STEP: Collecting logs
STEP: Run cleanup actions
STEP: Delete testing namespace
•
Ran 1 of 1 Specs in 0.743 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
```
