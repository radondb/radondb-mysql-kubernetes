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

- Run `simplecase`.

```
go test test/e2e/e2e_test.go -test.v --ginkgo.focus="simplecase"
```

or

```
go test test/e2e/e2e_test.go -test.v --ginkgo.label-filter=simplecase
```

- Skip `simplecase`.

```
go test test/e2e/e2e_test.go -test.v --ginkgo.skip="simplecase"
```

- Clean up

```
go test test/e2e/e2e_test.go -test.v --ginkgo.focus="clean" -clean-up=true
```

### Supported args

| Args               | Default                        | Description                                                 |
| ------------------ | ------------------------------ | ----------------------------------------------------------- |
| kubernetes-host    | ""                             | The kubernetes host, or apiserver, to connect to.           |
| kubernetes-config  | $KUBECONFIG                    | Path to config containing embedded authinfo for kubernetes. |
| kubernetes-context | current-context                | config context to use for kuberentes.                       |
| log-dir-prefix     | ""                             | Prefix of the log directory.                                |
| sidecar-image-path | radondb/mysql57-sidecar:v2.2.1 | Image full path of mysql sidecar.                           |
| pod-wait-timeout   | 1200                           | Timeout to wait for a pod to be ready.                      |
| dump-logs          | false                          | Dump logs when test case failed.                            |
