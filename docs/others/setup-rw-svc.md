# setup service for client to read/write

## 1. Get the pod about leader

```
# kubectl exec -ti $pod -c xenon xenoncli raft status | jq .leader | cut -d . -f 1 | tail -c +2
demo-krypton-1
```
the leader is the result: `demo-krypton-1`

## 2. Label role to the pod of the leader
```
# kubectl label pods $leader role=master
```
## 3. Edit the demo.yaml
```
# kubectl get svc $svc -oyaml > demo.yaml
```
$svc is demo-krypton, to edit it, for example

```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: demo-krypton
  name: demo-svc
  namespace: default
spec:
  ports:
  - name: mysql
    port: 3306
    protocol: TCP
    targetPort: mysql
  - name: metrics
    port: 9104
    protocol: TCP
    targetPort: metrics
  selector:
    app: demo-krypton
    role: master
  sessionAffinity: None
  type: ClusterIP
```

## 4. Apply the service

```# kubectl apply -f demo.yaml```

## 5. Check the service

```
# kubectl get svc
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
demo-krypton           NodePort    10.96.139.75    <none>        3306:30347/TCP,9104:32430/TCP   26h
demo-svc               ClusterIP   10.96.101.64    <none>        3306/TCP,9104/TCP               23h
```

```
# mysql -h10.96.101.64 -P 3306 -uxxx -pxxxx
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 19068
Server version: 5.7.33-36-log Percona Server (GPL), Release '36', Revision '7e403c5'

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like "read_only";
+---------------+-------+
| Variable_name | Value |
+---------------+-------+
| read_only     | OFF   |
+---------------+-------+
1 row in set (0.02 sec)
```