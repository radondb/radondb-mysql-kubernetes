# CLUSTER_NAME=sample
BASE=/juicefs
JSONFILE=$BASE/$CLUSTER_NAME-backup.json

function checkfile() {
    if ! [ -r  $JSONFILE ] ; then 
     jq -n --arg cluster $CLUSTER_NAME --arg namespace $NAMESPACE '{"cluster_name": $cluster, "namespace": $namespace,"backup_chains": []}' >$JSONFILE
    else 
     echo exist the json file
    fi
}
function read() {
    max=0
    IFS_OLD=$IFS
    IFS=$(echo -en "\n\b")
    for i in $(jq -c '.backup_chains[]' $JSONFILE);
    do 
        #echo $i | jq '.type'
        val=$(echo $i | jq '."target-dir"|match("\\d+")|.string|tonumber')
        #echo $val
       if [[ $max < $val ]] ; then
        max=$val
       fi
    done
    IFS=$IFS_OLD
    echo $max
}

function getDate() {
    date '+%Y-%m-%d %H:%M:%S'
}

function parseDateToUnix() {
    local t=$1
    
    echo date -d $t '+%s'|sh
}
function checkTime() {
    local time=$1 # get the parameter
    val=0
    IFS_OLD=$IFS
    IFS=$(echo -en "\n\b")
    for i in $(jq -c '.backup_chains[]' $JSONFILE);
    do 
        traw=$(echo $i|jq '."time"')
        val=$(echo $i | jq '."target-dir"|match("\\d+")|.string|tonumber')
        t=$(echo date -d $traw '+%s'|sh)
        cmptime=$(echo date -d "\"$time\"" '+%s'|sh)
        if [ $t -ge $cmptime ]; then 
            break
        fi
    done
    IFS=$IFS_OLD
    echo $val

}

function appendinc() {
    num=$1
    incbase="$BASE/backups/base"
    #echo $BASE/backups/inc$(echo $num + 1|bc)
    if ! [ $num -eq 0 ]; then
        incbase=$BASE/backups/inc$num
    fi
    jq ".backup_chains += [{\"type\": \"incr-backup\",  \"time\": \"$(getDate)\",  \"target-dir\": \"$BASE/backups/inc$(echo $num + 1|bc)\", 
        \"incremental-basedir\": \"$incbase\" }]" $JSONFILE >"tmp.json" && mv ./tmp.json $JSONFILE
}

function appendbase() {
    jq  ".backup_chains += [{\"type\": \"full-backup\",   \"time\": \"$(getDate)\",  \"target-dir\": \"$BASE/backups/base\"}]" $JSONFILE >"tmp.json" && mv ./tmp.json $JSONFILE
    sleep 2
}

function fullbackup() {
    mkdir -p /$BASE/backups/base
    xtrabackup --backup --host=127.0.0.1 --user=root --password=''  --datadir=/var/lib/mysql/ --target-dir=/$BASE/backups/base
    success=$?
    if [ $success ]; then 
        appendbase
    fi
}

function incrbackup() {
    num=$1
    incbase="$BASE/backups/base"
    #echo $BASE/backups/inc$(echo $num + 1|bc)
    if ! [ $num -eq 0 ]; then
        incbase=$BASE/backups/inc$num
    fi
    xtrabackup --backup --host=127.0.0.1 --user=root --password=''  --datadir=/var/lib/mysql/ --target-dir=$BASE/backups/inc$(echo $num + 1|bc) \
    --incremental-basedir=$incbase
    success=$?
    if [ $success ]; then 
        appendinc $num
    fi
}

function backup() {
    if ! [ -r  $JSONFILE ] ; then 
      jq -n --arg cluster $CLUSTER_NAME --arg namespace $NAMESPACE '{"cluster_name": $cluster, "namespace": $namespace,"backup_chains": []}' >$JSONFILE
      sleep 3
      echo now do the fullbackup
      fullbackup
    else 
     num=$(read)
     incrbackup $num
    fi

}


function restore() {
    local restorTime=$1
    local from=$2
    jsonfile=$BASE/$from-backup.json
    if [ $# -ne 2 ] ; then
        echo you can use it as restore date cluster-from
    fi
    local total=$(checkTime $restorTime)
    for index in $(seq 0 $total); do 
        # at restore, base always use /backups/base
        base=$(jq -c ".backup_chains[0][\"target-dir\"]" $jsonfile)
        type=$(jq -c ".backup_chains[$index][\"type\"]" $jsonfile)
        inc=$(jq -c ".backup_chains[$index][\"target-dir\"]" $jsonfile)
        cmd=""
        # echo $i, $base, $type,$inc
        case $type in
            "\"full-backup\"")
            cmd=$(echo xtrabackup --prepare --apply-log-only --target-dir=$base)
            echo $cmd
            echo $cmd|sh
            ;;
            "\"incr-backup\"")
                if [ $index -eq  $total ]; then
                    cmd=$(echo xtrabackup --prepare  --target-dir=$base --incremental-dir=$inc)
                else
                    cmd=$(echo xtrabackup --prepare --apply-log-only --target-dir=$base --incremental-dir=$inc)
                fi
                echo $cmd
                echo $cmd|sh
            ;;
            *)
            echo nothing
            ;;
        esac
    done
    # check /var/lib/mysql is emtpty
    if ! [ -d "/var/lib/mysql/mysql" ]; then
        base=$(jq -c ".backup_chains[0][\"target-dir\"]" $JSONFILE)
        cmd=$(echo xtrabackup --copy-back --target-dir=$base --datadir=/var/lib/mysql)
        echo $cmd
        echo $cmd|sh
        chown -R mysql.mysql /var/lib/mysql
    else
        echo the dir is not empty, cannot copy back
    fi
}

