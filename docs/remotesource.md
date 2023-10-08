English 

# Quickstart remote migration from remote source

## 1. perpare the remote source node

We suppose you have got a existence mysql database, such as percona server, or other mysql release.
and mysql version is 8.0.25 or mysql 5.7, this document use 8.0.25 for example. we suppose the ip of the node is `172.16.0.29` , and the root's password is `rootpass`. Make sure the sshd is ok, and it can login by `ssh` 

## 2. create secret in k8s
Now, what's need you to do is that create secret file in k8s, and it must contain the keys ,`host` and `passwd`. for example, we can create a secret named `remotesecret` as follow:
```
 kubectl create secret generic remotesecret  --from-literal=host=172.16.0.29    --from-literal=passwd=rootpass
```
## 3. fill the fields in mysqlcluster.yaml
fill the dataSouce's `remote` field, you should fill the name which is the secret's name, in this example, name is `remotesecret`, and fill `items` as same as below:

```yaml
spec:
  ...
  dataSource:
    ...
    remote: 
      sourceConfig:
          name: remotesecret
          items:
          - key: passwd
            path: passwd
          - key: host
            path: host
    
```
## 4. apply the yaml file, run it in k8s

```sh
kubectl apply -f mysql_v1beta1_mysqlcluster.yaml
```