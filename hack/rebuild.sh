#!/bin/bash
## This script is used to rebuild the pod.
## useage : ./rebuild.sh pod_name
function check_pod_status(){
    POD_STATUS=$(kubectl get pods $POD_NAME -o jsonpath='{.status.phase}')
    if [ "$POD_STATUS" == "Running" ]; then
        echo "Pod $POD_NAME is running."
    else
        echo "Pod $POD_NAME is not running."
        exit 1
    fi
}
function rebuild_pod(){
    INVALID=$(kubectl exec -it $POD_NAME -c xenon -- xenoncli raft status | grep "INVALID")
    if [ -z $INVALID ]; then
        echo "Pod in Raft is not in invalid state, please check your cluster status."
        exit 1
    fi

    kubectl exec -it $POD_NAME -c mysql  -- rm  -rf /var/lib/mysql/*

    sleep 5
    kubectl delete pod $POD_NAME
    echo "Rebuild mysql success."
}

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 POD_NAME." >&2
  exit 1
fi
POD_NAME=$1
echo "Rebuilding is danagerous job, Before doing this, please make sure your cluster data is safe and consistent."
while true; do
    read -p "Do you want to do it (y/n)?" yn
    case $yn in
        [Yy]* ) check_pod_status; rebuild_pod;break;;
        [Nn]* ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

