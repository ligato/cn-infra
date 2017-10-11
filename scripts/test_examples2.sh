#!/usr/bin/env bash

TMP_FILE="/tmp/out"
TMP_FILE2="/tmp/unprocessed"
processedLines=0 
exitCode=0
PREV_IFS="$IFS"

# test whether output of the command contains expected lines
# arguments
# 1-st command to run
# 2-nd array of expected strings in the command output
# 3-rd argument is an optional array of unexpected strings in the command output
# 4-th argument is an optional command runtime limit (value 0 does not cause stopping of the run.)
function testOutput {
IFS="${PREV_IFS}"

    #run the command
    if [ $# -ge 4 ]; then
        if [ -e ${TMP_FILE} ]; then
            # if exists file /tmp/out we assume that the command still runs, do not start it again
            echo > /dev/null
        else
            $1 > ${TMP_FILE} 2>&1 &
            CMD_PID=$!
        fi
        # do we set runtime limit?
        if [ $4 -ne 0 ]; then
            sleep $4
            kill $CMD_PID
        else
            # The command continues to run - his PID is in the variable CMD_PID
            sleep 20
        fi
        cat ${TMP_FILE} | awk "NR > $processedLines" > ${TMP_FILE2}
    else
        $1 > ${TMP_FILE} 2>&1
        cat ${TMP_FILE} > ${TMP_FILE2}
    fi

IFS="
"
    echo "Testing $1"
    rv=0
    # loop through expected lines
    for i in $2
    do
        if grep -- "${i}" ${TMP_FILE2} > /dev/null ; then
            echo "OK - '$i'"
        else
            echo "Not found - '$i'"
            rv=1
        fi
    done
    # loop through unexpected lines
    if [[ ! -z $3 ]] ; then
        for i in $3
        do
            if grep -- "${i}" ${TMP_FILE2} > /dev/null ; then
                echo "IS NOT OK - '$i'"
                rv=1
            fi
        done
    fi

    # if an error occurred print the output
    if [[ ! $rv -eq 0 ]] ; then
        cat ${TMP_FILE2}
        exitCode=1
    fi

    echo "================================================================"
    if [ $# -ge 4 ]; then
        if [ $4 -ne 0 ]; then
            rm ${TMP_FILE}
            rm ${TMP_FILE2}
        else
            # The command continues to run - the output is still redirected to the file ${TMP_FILE}
            # read -n1 -r -p "Press any key to continue..." key
            processedLines=`wc -l ${TMP_FILE} | cut --delimiter=" " -f1`
        fi
    else
        rm ${TMP_FILE}
        rm ${TMP_FILE2}
    fi
    return ${rv}
}

source scripts/docker_start_stop_functions.sh

#### Redis ###########################################################

# startRedis
# 
# expected=("config: redis.NodeConfig:
# INFO[0004] GetValue(key1) = true ; val = val 1 ; revision = 0  loc=\"simple/simple.go(208)\" logger=defaultLogger tag=00000000
# INFO[0004] GetValue(key2) = true ; val = val 2 ; revision = 0  loc=\"simple/simple.go(208)\" logger=defaultLogger tag=00000000
# ==> NOTE: key3 should have expired
# INFO[0004] GetValue(key3) = false                        loc=\"simple/simple.go(210)\" logger=defaultLogger tag=00000000
# ==> NOTE: get(key) should return false
# INFO[0004] GetValue(key) = false                         loc=\"simple/simple.go(210)\" logger=defaultLogger tag=00000000
# INFO[0004] ListKeys(key):  key2 (rev 0)                  loc=\"simple/simple.go(228)\" logger=defaultLogger tag=00000000
# INFO[0004] ListKeys(key):  key1 (rev 0)                  loc=\"simple/simple.go(228)\" logger=defaultLogger tag=00000000
# INFO[0004] ListKeys(key): count = 2                      loc=\"simple/simple.go(231)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key):  key2 = val 2 (rev 0)        loc=\"simple/simple.go(249)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key):  key1 = val 1 (rev 0)        loc=\"simple/simple.go(249)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key): count = 2                    loc=\"simple/simple.go(252)\" logger=defaultLogger tag=00000000
# INFO[0004] doKeyInterator(): Expected 100 keys; Found 100  loc=\"simple/simple.go(277)\" logger=defaultLogger tag=00000000
# INFO[0004] doKeyValInterator(): Expected 100 keyVals; Found 100  loc=\"simple/simple.go(315)\" logger=defaultLogger tag=00000000
# INFO[0004] Delete(key): found = true                     loc=\"simple/simple.go(341)\" logger=defaultLogger tag=00000000
# ==> NOTE: All keys should have been deleted
# INFO[0004] GetValue(key1) = false                        loc=\"simple/simple.go(210)\" logger=defaultLogger tag=00000000
# INFO[0004] GetValue(key2) = false                        loc=\"simple/simple.go(210)\" logger=defaultLogger tag=00000000
# INFO[0004] ListKeys(key): count = 0                      loc=\"simple/simple.go(231)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key): count = 0                    loc=\"simple/simple.go(252)\" logger=defaultLogger tag=00000000
# INFO[0004] txn(): keys = [key101 key102 key103 key104]   loc=\"simple/simple.go(353)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key):  key102 = 2 (rev 0)          loc=\"simple/simple.go(249)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key):  key104 = 4 (rev 0)          loc=\"simple/simple.go(249)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key):  key103 = 3 (rev 0)          loc=\"simple/simple.go(249)\" logger=defaultLogger tag=00000000
# INFO[0004] ListValues(key): count = 3                    loc=\"simple/simple.go(252)\" logger=defaultLogger tag=00000000
# INFO[0004] Sleep for 5 seconds                           loc=\"simple/simple.go(175)\" logger=defaultLogger tag=00000000
# INFO[0009] Closing connection                            loc=\"simple/simple.go(179)\" logger=defaultLogger tag=00000000
# ==> NOTE: Call on a closed connection should fail.
# ERRO[0009] Delete(key) called on a closed connection     loc=\"simple/simple.go(338)\" logger=defaultLogger tag=00000000
# INFO[0009] Sleep for 10 seconds                          loc=\"simple/simple.go(186)\" logger=defaultLogger tag=00000000
# ")
# 
# cmd="examples/redis-lib/simple/simple -redis-config=node-client.yaml "
# testOutput "${cmd}" "${expected}"
# 
# stopRedis

#### Simple-agent ########################################################

expected=("etcd config not found  - skip loading this plugin
kafka config not found  - skip loading this plugin
redis config not found  - skip loading this plugin
cassandra client config not found  - skip loading this plugin
All plugins initialized successfully
")

unexpected=("")

testOutput examples/simple-agent/simple-agent "${expected}" "${unexpected}" 5

#### Simple-agent with Kafka and ETCD ####################################

startEtcd
startKafka

expected=("Plugin etcdv3: status check probe registered
Plugin kafka: status check probe registered
redis config not found  - skip loading this plugin
cassandra client config not found  - skip loading this plugin
All plugins initialized successfully
")

unexpected=("")

cmd="examples/simple-agent/simple-agent --etcdv3-config=examples/datasync-plugin/etcd.conf --kafka-config examples/kafka-plugin/hash-partitioner/kafka.conf"
testOutput "${cmd}" "${expected}" "${unexpected}" 5

stopEtcd
stopKafka

#### Simple-agent with Cassandra and Redis and Kafka and ETCD ####################################

startEtcd
startCustomizedKafka examples/kafka-plugin/manual-partitioner/server.properties
startRedis
startCassandra

expected=("Plugin etcdv3: status check probe registered
Plugin redis: status check probe registered
Plugin cassandra: status check probe registered
Plugin kafka: status check probe registered
All plugins initialized successfully
Agent plugin state update.*plugin=etcdv3 state=ok
Agent plugin state update.*plugin=redis state=ok
Agent plugin state update.*plugin=cassandra state=ok
Agent plugin state update.*plugin=kafka state=ok
")

unexpected=("redis config not found  - skip loading this plugin
cassandra client config not found  - skip loading this plugin
")

cmd="examples/simple-agent/simple-agent --etcdv3-config=examples/etcdv3-lib/etcd.conf --kafka-config=examples/kafka-plugin/manual-partitioner/kafka.conf  --redis-config=examples/redis-lib/node-client.yaml --cassandra-config=examples/cassandra-lib/client-config.yaml"
testOutput "${cmd}" "${expected}" "${unexpected}" 0 # the cmd continues to run - we will kill it later

stopRedis >> /dev/null
sleep 10
docker exec etcd etcdctl get --prefix "" | grep  redis

expected=("Agent plugin state update.*Get(/probe-redis-connection) failed: EOF.*status-check.*plugin=redis state=error
")

unexpected=("Agent plugin state update.*plugin=redis state=ok
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT disconnected

startRedis >> /dev/null
sleep 10
docker exec etcd etcdctl get --prefix "" | grep redis

expected=("Agent plugin state update.*plugin=redis state=ok
")

unexpected=("Agent plugin state update.*Get(/probe-redis-connection) failed: EOF.*status-check.*plugin=redis state=error 
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT connected AGAIN

kill $CMD_PID > /dev/null
rm ${TMP_FILE} > /dev/null
rm ${TMP_FILE2} > /dev/null

stopEtcd
stopKafka
stopRedis
stopCassandra

##########################################################################

exit ${exitCode}
