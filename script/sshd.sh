cp  /etc/secret-ssh/* /root/.ssh
chmod 600 /root/.ssh/authorized_keys
/usr/sbin/sshd -D -e -f /etc/ssh/sshd_config &
echo "start..."
