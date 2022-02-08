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
ginkgo test/e2e/
```

- Running all cases labeled `simplecase`.

```
ginkgo --label-filter=simplecase test/e2e/
```

- Skip the cases of describing information contains `Namespace`.

```
ginkgo --skip "list namespace" test/e2e/
```

- Just run the description information contains `Namespace`'s cases.

```
ginkgo --focus "list namespace" test/e2e/
```
