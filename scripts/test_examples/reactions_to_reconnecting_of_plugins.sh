#!/usr/bin/env bash
source scripts/test_examples/docker_start_stop_functions.sh

# In this file there are tested the CN-infra examples.
# These examples are located in the folder examples.
# The function for testing output of executed example - testOutput - can be used
# in three modes - depending on the way how executed exampes works:
# - the executed example will stop its run itself (no need to use 4th parameter)
# - the executed example does not stop itself (it has to be killed after some
#   time - for this is used the 4th parameter of the function testOutput).
# - the executed example does not stop itself (it has to be killed after some
#   time). In contrast with previous case where the executed example is killed
š   and the output is processed only once time we use the special value
#   for the 4th parameter - value 0 - to postpone the killing of process which
#   will alow to process output of executed example several times to monitor
#   reactions of executed example to some outer influences (e.g. stop/start of
#   some services executed before the tested example.)
# In this file are preferentially stored the tests which are run in the third
# mode.

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
sleep 3
docker exec -it etcd etcdctl get /vnf-agent/vpp1/check/status/v1/plugin/redis

expected=("Agent plugin state update.*Get(/probe-redis-connection) failed: EOF.*status-check.*plugin=redis state=error
")

unexpected=("Agent plugin state update.*plugin=redis state=ok
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT disconnected

startRedis >> /dev/null
sleep 3
docker exec -it etcd etcdctl get /vnf-agent/vpp1/check/status/v1/plugin/redis

expected=("Agent plugin state update.*plugin=redis state=ok
")

unexpected=("Agent plugin state update.*Get(/probe-redis-connection) failed: EOF.*status-check.*plugin=redis state=error
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT connected AGAIN

# cassandra start/stop test
stopCassandra >> /dev/null
sleep 3
docker exec -it etcd etcdctl get /vnf-agent/vpp1/check/status/v1/plugin/cassandra

expected=("Agent plugin state update.*gocql: no hosts available in the pool.*status-check plugin=cassandra state=error
")

unexpected=("Agent plugin state update.*plugin=cassandra state=ok
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT disconnected

startCassandra >> /dev/null
sleep 3
docker exec -it etcd etcdctl get /vnf-agent/vpp1/check/status/v1/plugin/cassandra

expected=("Agent plugin state update.*plugin=cassandra state=ok
")

unexpected=("Agent plugin state update.*gocql: no hosts available in the pool.*status-check plugin=cassandra state=error
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT connected AGAIN

# kafka start/stop test
stopKafka >> /dev/null
sleep 3
docker exec -it etcd etcdctl get /vnf-agent/vpp1/check/status/v1/plugin/kafka

expected=("Agent plugin state update.*kafka: client has run out of available brokers to talk to (Is your cluster reachable?).*status-check plugin=kafka state=error
")

unexpected=("Agent plugin state update.*plugin=kafka state=ok
")

testOutput "${cmd}" "${expected}" "${unexpected}" 0 # cmd unchanged - ASSERT disconnected

startKafka >> /dev/null
sleep 3
docker exec -it etcd etcdctl get /vnf-agent/vpp1/check/status/v1/plugin/kafka

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
