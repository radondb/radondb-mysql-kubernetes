export KUBECONFIG=~/.kube/config
export KUBERNETES_SERVICE_HOST= 
export KUBERNETES_SERVICE_PORT=6443
export S3ENDPOINT=
export S3ACCESSKEY=
export S3SECRETKEY=
export ACK_GINKGO_DEPRECATIONS=1.16.5
kubectl delete namespace radondb-mysql-e2e
kubectl get clusterrole|grep mysql|awk  '{print "kubectl delete clusterrole "$1}'|bash
kubectl get clusterrolebindings|grep mysql|awk  '{print "kubectl delete clusterrolebindings "$1}'|bash
kubectl get crd|grep mysql|awk  '{print "kubectl delete crd "$1}'|bash

make e2e-local