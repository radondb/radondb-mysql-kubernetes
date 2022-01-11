# How to run e2e test

## Prerequisites

- Prepare a client connected to K8S.
- Make sure [Ginkgo V2](https://onsi.github.io/ginkgo/MIGRATING_TO_V2) is installed.

## Hands-on Lab

### Step 1: Configure environment variables

```
export KUBECONFIG=$HOME/.kube/config
```

### Step 2: Run test

> The Ginkgo version of the following examples is V2.

- Running all cases.

```
go test test/e2e/e2e_test.go -test.v
```

- Running all cases labeled `simplecase`.

```
go test test/e2e/e2e_test.go -test.v --ginkgo.label-filter=simplecase
```

- Skip the cases of describing information contains `Namespace`.

```
go test test/e2e/e2e_test.go -test.v --ginkgo.skip="list namespace"
```

- Just run the description information contains `Namespace`'s cases.

```
go test test/e2e/e2e_test.go -test.v --ginkgo.focus="list namespace"
```

### Supported args

| Args                | Default                        | Description                                                 |
| ------------------- | ------------------------------ | ----------------------------------------------------------- |
| kubernetes-host     | ""                             | The kubernetes host, or apiserver, to connect to.           |
| kubernetes-config   | $KUBECONFIG                    | Path to config containing embedded authinfo for kubernetes. |
| kubernetes-context  | current-context                | config context to use for kuberentes.                       |
| log-dir-prefix      | ""                             | Prefix of the log directory.                                |
| chart-path          | ../../charts/mysql-operator    | The chart name or path for mysql operator.                  |
| operator-image-path | radondb/mysql-operator:v2.2.0  | Image path of mysql operator.                               |
| sidecar-image-path  | radondb/mysql57-sidecar:v2.2.0 | Image full path of mysql sidecar.                           |
| pod-wait-timeout    | 1200                           | Timeout to wait for a pod to be ready.                      |
| dump-logs           | false                          | Dump logs when test case failed.                            |