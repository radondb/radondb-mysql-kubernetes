 if [ $# != 1 ] ; then
  echo "need source and dest"
  exit 1
fi
dest=$1
ctrl=kubectl
$ctrl delete pod/backup-access 2>/dev/null || :

	cat - <<-EOF | $ctrl apply -f -
		apiVersion: v1
		kind: Pod
		metadata:
		  name: backup-access
		spec:
		  containers:
		  - name: xtrabackup
		    image: 'acekingke/sidecar:v01'
		    command: [ "/bin/bash", "-c", "--" ]
		    args: [ "rm -rf /backup/*;curl --user sys_backups:sys_backups sample-mysql-0.sample-mysql.default:8082/download|xbstream -x -C /backup" ]
		    volumeMounts:
		    - name: backup
		      mountPath: /backup        
		  restartPolicy: Never
		  volumes:
		  - name: backup
		    persistentVolumeClaim:
		      claimName: ${dest}
	EOF

	echo -n Starting pod.
	until $ctrl get pod/backup-access -o jsonpath='{.status.containerStatuses[0].ready}' 2>/dev/null | grep -q 'true'; do
		sleep 1
		echo -n .
	done
	echo "[done]"
    # kubectl cp backup-access:backup /root/backup

# curl --user sys_backups:sys_backups sample-mysql-0.sample-mysql.default:8082/xbackup

# curl --user sys_backups:sys_backups sample-mysql-0.sample-mysql.default:8082/download|xbstream -x -C /backup