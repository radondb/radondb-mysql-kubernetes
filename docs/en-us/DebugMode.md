# Debug Mode
When you want avoid the restart-on-fail loop for mysql container, You should use Debug Mode.
it just use create a empty file  `/var/lib/mysql/sleep-forever`
for example:
```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'touch /var/lib/mysql/sleep-forever'
```
it make pod sample-mysql-0's mysql container will never restart when mysqld is crashed.

# Remove Debug Mode

```bash
kubectl exec -it sample-mysql-0 -c mysql -- sh -c 'rm /var/lib/mysql/sleep-forever'
```
restart the container