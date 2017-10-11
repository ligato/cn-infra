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

# redis start/stop test
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

# cassandra start/stop test
stopCassandra >> /dev/null
sleep 10
docker exec etcd etcdctl get --prefix "" | grep  gocql

expected=("Agent plugin state update.*gocql: no hosts available in the pool.*status-check plugin=cassandra state=error
")

unexpected=("Agent plugin state update.*plugin=cassandra state=ok
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT disconnected

startCassandra >> /dev/null
sleep 10
docker exec etcd etcdctl get --prefix "" | grep gocql

expected=("Agent plugin state update.*plugin=cassandra state=ok
")

unexpected=("Agent plugin state update.*gocql: no hosts available in the pool.*status-check plugin=cassandra state=error 
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT connected AGAIN

# kafka start/stop test
stopKafka >> /dev/null
sleep 10
docker exec etcd etcdctl get --prefix "" | grep  kafka

expected=("Agent plugin state update.*kafka: client has run out of available brokers to talk to (Is your cluster reachable?).*status-check plugin=kafka state=error
")

unexpected=("Agent plugin state update.*plugin=kafka state=ok
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT disconnected

startKafka >> /dev/null
sleep 10
docker exec etcd etcdctl get --prefix "" | grep kafka

expected=("Agent plugin state update.*plugin=kafka state=ok
")

unexpected=("Agent plugin state update.*kafka: client has run out of available brokers to talk to (Is your cluster reachable?).*status-check plugin=kafka state=error 
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
