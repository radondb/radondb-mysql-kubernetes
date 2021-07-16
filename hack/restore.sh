
#!/bin/bash
# do not has RESTORE_FROM env variable
if [ -z "$RESTORE_FROM" ]; then
    echo "nothing to do"
	exit 0
fi
if [ -z "$S3_ENDPOINT" ] || [ -z "$S3_ACCESSKEY" ] || [ -z "$S3_SECRETKEY" ] || [ -z "$S3_BUCKET" ]; then
	echo "nothing to do "
    exit 0
fi
if [ ! -d "/var/lib/mysql" ] ; then
    echo "is not exist the var lib mysql"
    mkdir /var/lib/mysql
    chown -R mysql.mysql /var/lib/mysql
fi
mkdir /root/backup
xbcloud get --storage=S3 \
--s3-endpoint="${S3_ENDPOINT}" \
--s3-access-key="${S3_ACCESSKEY}" \
--s3-secret-key="${S3_SECRETKEY}" \
--s3-bucket="${S3_BUCKET}" \
--parallel=10 $RESTORE_FROM \
--insecure |xbstream -xv -C /root/backup
# prepare redolog
xtrabackup --defaults-file=/etc/mysql/my.cnf --use-memory=3072M --prepare --apply-log-only --target-dir=/root/backup
# prepare data
xtrabackup --defaults-file=/etc/mysql/my.cnf --use-memory=3072M --prepare --target-dir=/root/backup
chown -R mysql.mysql /root/backup
xtrabackup --defaults-file=/etc/mysql/my.cnf --datadir=/var/lib/mysql --copy-back --target-dir=/root/backup
chown -R mysql.mysql /var/lib/mysql
rm -rf /root/backup
