#!/usr/bin/bash
CLUSTER=
IMAGE=
update () { 
# set innodb_fast_shutdown=0 in yaml file
kubectl patch mysql $CLUSTER --type=merge --patch '{"spec":
            {"mysqlOpts": 
                {"mysqlConf": {"innodb_fast_shutdown":"0"}
                }
            }
        }'

checkReadyOrClosed() {
     kubectl get mysqlclusters.mysql.radondb.com	 -o jsonpath='{.items[0].status.state}' 2>/dev/null |grep -q 'Ready'||kubectl get mysqlclusters.mysql.radondb.com	 -o jsonpath='{.items[0].status.state}' 2>/dev/null |grep -q 'Closed'
}
checkStatefulSetUpdated() {
    # it has ready
    # get statefulset ready replicas
    ready=$(kubectl get statefulsets.apps -o jsonpath='{.items[0].status.readyReplicas}' 2>/dev/null)
    updateVersion=$(kubectl get statefulsets.apps -o jsonpath='{.items[0].status.updateRevision}' 2>/dev/null)  
    count=$(kubectl get pod -l "app.kubernetes.io/managed-by"="mysql.radondb.com" -l "controller-revision-hash"=$updateVersion |awk  'NR>1{print $1}'|wc -l)
    if [ $count -eq $ready ];then
        return 0
    else
        return 1
    fi
}
waitReady () {
    timeout=$1
    SECONDS=0
    echo -n Wating
    until  checkReadyOrClosed; do
        sleep 2
        echo -n "."
        if [ $SECONDS -gt $timeout ]; then
            echo "timeout"
            break
        fi
    done
    echo
}

waitReady 30
until checkStatefulSetUpdated; do
    echo "waiting for statefulset updated"
    echo -n "."
    sleep 5
done
echo
pods=$(kubectl get pod -l "app.kubernetes.io/managed-by"="mysql.radondb.com"|awk  'NR>1{print $1}')
for p in $pods ; do
    kubectl exec -it $p -c mysql -- bash <<EOF 
    rm -rf /var/lib/mysql/ib_logfile*
EOF
done
kubectl patch mysql $CLUSTER --type=merge --patch "{\"spec\":
            {\"mysqlVersion\": \"8.0\",
            \"podPolicy\": {\"sidecarImage\": \"$IMAGE\"}
             }
    }"

}

check() {
  kubectl exec -it svc/$CLUSTER-leader -c mysql -- bash <<EOF 
        mysqlcheck -u root -pRadonDB@123 --all-databases --check-upgrade 
EOF
    kubectl exec -it svc/$CLUSTER-leader -c mysql -- bash <<EOF 
    mysql -uroot -pRadonDB@123 <<MOF
    SELECT TABLE_SCHEMA, TABLE_NAME
    FROM INFORMATION_SCHEMA.TABLES
    WHERE ENGINE NOT IN ('innodb', 'ndbcluster')
    AND CREATE_OPTIONS LIKE '%partitioned%';
    MOF
EOF
    # if has partition , alter them to innodb
    # ALTER TABLE table_name ENGINE = INNODB;

    # check the special table name in mysql57 information_schema
    kubectl exec -it svc/$CLUSTER-leader -c mysql -- bash <<EOF 
    mysql -uroot -pRadonDB@123 <<MOF
    SELECT TABLE_SCHEMA, TABLE_NAME
    FROM INFORMATION_SCHEMA.TABLES
    WHERE LOWER(TABLE_SCHEMA) = 'mysql'
    and LOWER(TABLE_NAME) IN
    (
    'catalogs',
    'character_sets',
    'check_constraints',
    'collations',
    'column_statistics',
    'column_type_elements',
    'columns',
    'dd_properties',
    'events',
    'foreign_key_column_usage',
    'foreign_keys',
    'index_column_usage',
    'index_partitions',
    'index_stats',
    'indexes',
    'parameter_type_elements',
    'parameters',
    'resource_groups',
    'routines',
    'schemata',
    'st_spatial_reference_systems',
    'table_partition_values',
    'table_partitions',
    'table_stats',
    'tables',
    'tablespace_files',
    'tablespaces',
    'triggers',
    'view_routine_usage',
    'view_table_usage'
    );
MOF
EOF

    # if exist RENAME TABLE 
    # check the constraints name over 64
    kubectl exec -it svc/$CLUSTER-leader -c mysql -- bash <<EOF 
    mysql -uroot -pRadonDB@123 <<MOF
    SELECT TABLE_SCHEMA, TABLE_NAME
    FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_NAME IN
    (SELECT LEFT(SUBSTR(ID,INSTR(ID,'/')+1),
                INSTR(SUBSTR(ID,INSTR(ID,'/')+1),'_ibfk_')-1)
    FROM INFORMATION_SCHEMA.INNODB_SYS_FOREIGN
    WHERE LENGTH(SUBSTR(ID,INSTR(ID,'/')+1))>64);
MOF
EOF
    # There must be no views with explicitly defined columns names that exceed 64 characters

    # partitioned table 
    kubectl exec -it svc/$CLUSTER-leader -c mysql -- bash <<EOF 
    mysql -uroot -pRadonDB@123 <<MOF
    SELECT DISTINCT NAME, SPACE, SPACE_TYPE FROM INFORMATION_SCHEMA.INNODB_SYS_TABLES
    WHERE NAME LIKE '%#P#%' AND SPACE_TYPE NOT LIKE 'Single';
MOF
EOF

    ## ALTER TABLE table_name REORGANIZE PARTITION partition_name
    ##  INTO (partition_definition TABLESPACE=innodb_file_per_table); 
    ##

}

main() {
    CHOICE=$1
    CLUSTER=$2
    IMAGE=$3
    # choose the action
    case $CHOICE in
        "update")
            update 
            ;;
        "check")
            check 
            ;;
        *)
            echo "Usage: $0 [update|check] [cluster name] [image name]"
            ;;
    esac

}
main "$@"