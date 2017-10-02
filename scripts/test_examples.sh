#!/usr/bin/env bash

TMP_FILE="/tmp/out"
exitCode=0
PREV_IFS="$IFS"

# test whether output of the command contains expected lines
# arguments
# 1-st command to run
# 2-nd array of expected strings in the command output
# 3-rd argument is an optional command runtime limit
function testOutput {
IFS="${PREV_IFS}"

    #run the command
    if [ $# -ge 3 ]; then
        $1 > ${TMP_FILE} 2>&1 &
        CMD_PID=$!
        sleep $3
        kill $CMD_PID
    else
        $1 > ${TMP_FILE} 2>&1
    fi

IFS="
"
    echo "Testing $1"
    rv=0
    # loop through expected lines
    for i in $2
    do
        if grep -- "${i}" /tmp/out > /dev/null ; then
            echo "OK - '$i'"
        else
            echo "Not found - '$i'"
            rv=1
        fi
    done

    # if an error occurred print the output
    if [[ ! $rv -eq 0 ]] ; then
        cat ${TMP_FILE}
        exitCode=1
    fi

    echo "================================================================"
    rm ${TMP_FILE}
    return ${rv}
}

function startEtcd {
    docker run -p 2379:2379 --name etcd -d -e ETCDCTL_API=3 \
        quay.io/coreos/etcd:v3.1.0 /usr/local/bin/etcd \
             -advertise-client-urls http://0.0.0.0:2379 \
                 -listen-client-urls http://0.0.0.0:2379 > /dev/null
    # dump etcd content to make sure that etcd is ready
    docker exec etcd etcdctl get --prefix ""
    # sometimes etcd needs a bit more time to fully initialize
    sleep 2
}

function stopEtcd {
    docker stop etcd > /dev/null
    docker rm etcd > /dev/null
}

function startKafka {
    docker run -p 2181:2181 -p 9092:9092 --name kafka -d \
        --env ADVERTISED_HOST=0.0.0.0 --env ADVERTISED_PORT=9092 spotify/kafka > /dev/null
    KAFKA_VERSION=$(docker exec kafka /bin/bash -c 'echo $KAFKA_VERSION')
    SCALA_VERSION=$(docker exec kafka /bin/bash -c 'echo $SCALA_VERSION')
    # list kafka topics to ensure that kafka is ready
    docker exec kafka  /opt/kafka_${SCALA_VERSION}-${KAFKA_VERSION}/bin/kafka-topics.sh --list --zookeeper localhost:2181 > /dev/null 2> /dev/null
    # sometimes Kafka needs a bit more time to fully initialize
    sleep 2
}

# startCustomizedKafka takes path to server.properties as the only argument.
function startCustomizedKafka {
    docker create -p 2181:2181 -p 9092:9092 --name kafka \
        --env ADVERTISED_HOST=0.0.0.0 --env ADVERTISED_PORT=9092 spotify/kafka > /dev/null
    KAFKA_VERSION=$(docker inspect -f '{{ .Config.Env }}' kafka |  tr ' ' '\n' | grep KAFKA_VERSION | sed 's/^.*=//')
    SCALA_VERSION=$(docker inspect -f '{{ .Config.Env }}' kafka |  tr ' ' '\n' | grep SCALA_VERSION | sed 's/^.*=//')
    docker cp $1 kafka:/opt/kafka_${SCALA_VERSION}-${KAFKA_VERSION}/config/server.properties
    docker start kafka > /dev/null
    # list kafka topics to ensure that kafka is ready
    docker exec kafka  /opt/kafka_${SCALA_VERSION}-${KAFKA_VERSION}/bin/kafka-topics.sh --list --zookeeper localhost:2181 > /dev/null 2> /dev/null
    # sometimes Kafka needs a bit more time to fully initialize
    sleep 2
}

function stopKafka {
    docker stop kafka > /dev/null
    docker rm kafka > /dev/null
}

function startCassandra {
    docker run -p 9042:9042 --name cassandra01 -d cassandra > /dev/null 2> /dev/null
    # Wait until cassandra is ready to accept a connection.
    for attemptps in {1..20} ; do
        NODEINFO=$(docker exec -it cassandra01 nodetool info)
        if [ $? -eq 0 ]; then
            if [[ ${NODEINFO} == *"Native Transport active: true"* ]]; then
                break
            fi
        fi
    done
    # sometimes Cassandra needs a bit more time to fully initialize
    sleep 2
}

function stopCassandra {
    docker stop cassandra01 > /dev/null
    docker rm cassandra01 > /dev/null
}

#### Cassandra ###########################################################

startCassandra

expected=("Successfully written
Successfully queried
Successfully queried with AND
Successfully queried with Multiple AND
Successfully queried with IN
")

cmd="examples/cassandra-lib/cassandra-lib examples/cassandra-lib/client-config.yaml"
testOutput "${cmd}" "${expected}"

stopCassandra

#### Configs #############################################################

expected=("Loaded plugin config - found external configuration examples/configs-plugin/example.conf
Plugin Config {Field1:external value, Sleep:0s}
")

cmd="examples/configs-plugin/configs-plugin --config-dir=examples/configs-plugin --example-config=example.conf"
testOutput "${cmd}" "${expected}"

#### Datasync ############################################################

startEtcd

expected=("KeyValProtoWatcher subscribed
Write data to /vnf-agent/vpp1/api/v1/example/db/simple/index
Update data at /vnf-agent/vpp1/api/v1/example/db/simple/index
Event arrived to etcd eventHandler, key /vnf-agent/vpp1/api/v1/example/db/simple/index, update: false
Event arrived to etcd eventHandler, key /vnf-agent/vpp1/api/v1/example/db/simple/index, update: true
")

cmd="examples/datasync-plugin/datasync-plugin --etcdv3-config=examples/datasync-plugin/etcd.conf"
testOutput "${cmd}" "${expected}"

stopEtcd

#### Etcdv3-lib ##########################################################

startEtcd

expected=("Saving  /phonebook/Peter
")

cmd="examples/etcdv3-lib/editor/editor --cfg examples/etcdv3-lib/etcd.conf  put  Peter Company 0907"
testOutput "${cmd}" "${expected}"

stopEtcd

#### Flags-lib ###########################################################

expected=("Registering flags...
Printing flags...
testFlagString:'mystring'
testFlagInt:'1122'
testFlagInt64:'-3344'
testFlagUint:'112'
testFlagUint64:'7788'
testFlagBool:'true'
testFlagDur:'5s'
")

cmd="examples/flags-lib/flags-lib --ep-string mystring --ep-uint 112"
testOutput "${cmd}" "${expected}"

#### Kafka-lib ###########################################################

startKafka

expected=("Kafka connecting
Consuming started
Sync published
Message is stored in topic(test)/partition(0)/offset(1)
")

testOutput examples/kafka-lib/mux/mux "${expected}"

stopKafka

#### Kafka-plugin manual-partitioner #####################################

startCustomizedKafka examples/kafka-plugin/manual-partitioner/server.properties

# Let us test parameters display
expected=("Usage of examples/kafka-plugin/manual-partitioner/manual-partitioner:
  -config-dir=\".\": Location of the configuration files; also set via 'CONFIG_DIR' env variable.
  -kafka-config=\"kafka.conf\": Location of the kafka Client configuration file; also set via 'KAFKA_CONFIG' env variable.
  -logs-config=\"logs.conf\": Location of the configuration files; also set via 'LOGS_CONFIG' env variable.
  -messageCount=\"\": Number of messages which will be send. Set to '0' to just watch
  -microservice-label=\"vpp1\": microservice label; also set via 'MICROSERVICE_LABEL' env variable.
  -offsetMsg=\"\": Use 'latest', 'oldest' or exact number of message offset
")

cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf -h"
testOutput "${cmd}" "${expected}"

# Let us test the running without parameters - in example are generated 10 Kafka Messages to both topics but the consumed is only 5 messages from each topic beginning with offset 5
expected=("Invalid partition -1 - defaulting to 0
offset arg not set, using default value
messageCount arg not set, using default value
Offset: 5, message count: 10
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Sending 10 sync Kafka notifications (protobuf) ...
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key', 
Sending 10 async Kafka notifications (protobuf) ...
Async message successfully delivered
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '9', key: 'async-proto-key', 
")

cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf"
testOutput "${cmd}" "${expected}"

# Let us test - in example are generated next more 10 Kafka Messages to both topics
expected=("Invalid partition -1 - defaulting to 0
offset arg not set, using default value
messageCount arg not set, using default value
Offset: 5, message count: 10
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '9', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key', 
Sending 10 sync Kafka notifications (protobuf) ...
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '10', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '11', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '12', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '13', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '14', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '15', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '16', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '17', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '18', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '19', key: 'proto-key', 
Sending 10 async Kafka notifications (protobuf) ...
Async message successfully delivered
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '10', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '11', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '12', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '13', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '14', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '15', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '16', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '17', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '18', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '19', key: 'async-proto-key', 
")

cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf"
testOutput "${cmd}" "${expected}"

# Let us test - in example no new messages generated - we are consummed all beginning with offset 5 till 19  (which were generated before)
expected=("Invalid partition -1 - defaulting to 0
offset arg not set, using default value
Offset: 5, message count: 0
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '9', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '10', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '11', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '12', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '13', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '14', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '15', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '16', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '17', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '18', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '19', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '10', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '11', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '12', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '13', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '14', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '15', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '16', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '17', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '18', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '19', key: 'proto-key', 
Sending 0 sync Kafka notifications (protobuf) ...
Sending 0 async Kafka notifications (protobuf) ...
")

#this is not working - I do not know why
#cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf -messageCount=\"0\""
#testOutput "${cmd}" "${expected}"

#result is:
#time="2017-09-29 14:23:04.54780" level=info msg="Agent AfterInit took 0s" durationInNs=0 loc="core/agent_core.go(124)" logger=core tag=00000000 
#time="2017-09-29 14:23:04.54782" level=error msg="Error loading core: plugin KafkaExample: Init error ''messageCount' has to be a number, not "0"', took 44.642µs" loc="core/event_loop.go(28)" logger=core tag=00000000 
#Testing examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf -messageCount="0"

#workaround
cmd='examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 0'
testOutput "${cmd}" "${expected}"


# Let us test - in example one new message generated
expected=("Invalid partition -1 - defaulting to 0
offset arg not set, using default value
Offset: 5, message count: 1
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '9', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '10', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '11', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '12', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '13', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '14', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '15', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '16', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '17', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '18', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '19', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '10', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '11', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '12', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '13', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '14', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '15', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '16', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '17', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '18', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '19', key: 'proto-key', 
Sending 1 sync Kafka notifications (protobuf) ...
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '20', key: 'proto-key', 
Sending 1 async Kafka notifications (protobuf) ...
Async message successfully delivered
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '20', key: 'async-proto-key', 
")

# this does not work: cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf -messageCount=\"1\""
cmd='examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 1'
testOutput "${cmd}" "${expected}"


# Let us test - now let us test the offset (it relates to both topics) - till now we generated XX messages
#./manual-partitioner -messageCount="1" -offsetMsg="18"
#./manual-partitioner -messageCount="0" -offsetMsg="20"
#./manual-partitioner -messageCount="0" -offsetMsg="latest"
#./manual-partitioner -messageCount="0" -offsetMsg="oldest"

# Let us test - in example one new message generated - with offset 21 for both topics
expected=("Invalid partition -1 - defaulting to 0
Offset: 18, message count: 1
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '18', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '19', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '20', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '18', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '19', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '20', key: 'proto-key', 
Sending 1 sync Kafka notifications (protobuf) ...
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '21', key: 'proto-key', 
Sending 1 async Kafka notifications (protobuf) ...
Async message successfully delivered
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '21', key: 'async-proto-key', 
")


cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 1 --offsetMsg 18"
testOutput "${cmd}" "${expected}"


# Let us test - in example no new message generated
expected=("Invalid partition -1 - defaulting to 0
Offset: 20, message count: 0
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '20', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '21', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '20', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '21', key: 'proto-key', 
Sending 0 sync Kafka notifications (protobuf) ...
Sending 0 async Kafka notifications (protobuf) ...
")


cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 0 --offsetMsg 20"
testOutput "${cmd}" "${expected}"

# Let us test - in example no new messages generated - we want to list all latest messages
expected=("Invalid partition -1 - defaulting to 0
Offset: -1, message count: 0
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '0', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '1', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '2', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '3', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '4', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '9', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '10', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '11', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '12', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '13', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '14', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '15', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '16', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '17', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '18', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '19', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '20', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '21', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '0', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '1', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '2', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '3', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '4', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '10', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '11', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '12', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '13', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '14', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '15', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '16', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '17', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '18', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '19', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '20', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '21', key: 'proto-key', 
Sending 0 sync Kafka notifications (protobuf) ...
Sending 0 async Kafka notifications (protobuf) ...
")


# this is not working  cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 0 -offsetMsg=\"latest\""
cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 0 -offsetMsg=latest"
testOutput "${cmd}" "${expected}"


# Let us test - in example no new messages generated - we want to list all oldest messages
expected=("Invalid partition -1 - defaulting to 0
Offset: -2, message count: 0
Started Kafka sync event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '0', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '1', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '2', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '3', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '4', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '9', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '10', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '11', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '12', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '13', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '14', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '15', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '16', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '17', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '18', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '19', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '20', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '2', offset '21', key: 'async-proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '0', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '1', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '2', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '3', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '4', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '10', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '11', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '12', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '13', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '14', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '15', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '16', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '17', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '18', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '19', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '20', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '21', key: 'proto-key', 
Sending 0 sync Kafka notifications (protobuf) ...
Sending 0 async Kafka notifications (protobuf) ...
")



cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 0 -offsetMsg=oldest"
testOutput "${cmd}" "${expected}"


# Let us test - in example no new messages generated - wrong value of parameter offsetMsg
expected=("Invalid partition -1 - defaulting to 0
Error loading core: plugin KafkaExample: Init error 'incorrect sync offset value wronginput
")


cmd="examples/kafka-plugin/manual-partitioner/manual-partitioner --kafka-config examples/kafka-plugin/manual-partitioner/kafka.conf --messageCount 0 -offsetMsg=wronginput"
testOutput "${cmd}" "${expected}"


stopKafka

#### Kafka-plugin hash-partitioner #######################################

startKafka

# Let us test parameters display
expected=("Usage of ./hash-partitioner:
-config-dir=\".\": Location of the configuration files; also set via 'CONFIG_DIR' env variable.
-kafka-config=\"kafka.conf\": Location of the kafka Client configuration file; also set via 'KAFKA_CONFIG' env variable.
-logs-config=\"logs.conf\": Location of the configuration files; also set via 'LOGS_CONFIG' env variable.
-messageCount=\"\": Number of messages which will be send. Set to '0' to just watch
-microservice-label=\"vpp1\": microservice label; also set via 'MICROSERVICE_LABEL' env variable.
")

cmd="examples/kafka-plugin/hash-partitioner/hash-partitioner --kafka-config examples/kafka-plugin/hash-partitioner/kafka.conf -h"
testOutput "${cmd}" "${expected}"

# Let us test the running without parameters - in example are generated 10 Kafka Messages to both topics
expected=("messageCount arg not set, using default value
Started Kafka event handler...
Started Kafka async event handler...
Sending 10 sync Kafka notifications (protobuf) ...
Sending 10 async Kafka notifications (protobuf) ...
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '0', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '2', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '1', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '4', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '3', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '6', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '5', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '8', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '7', key: 'async-proto-key', 
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '9', key: 'async-proto-key', 
All plugins initialized successfully
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '0', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '1', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '2', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '3', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '4', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '5', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '6', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '7', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '8', key: 'proto-key', 
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '9', key: 'proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '0', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '1', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '2', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '3', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '4', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '5', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '6', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '7', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '8', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '9', key: 'async-proto-key', 
")

cmd="examples/kafka-plugin/hash-partitioner/hash-partitioner --kafka-config examples/kafka-plugin/hash-partitioner/kafka.conf"
testOutput "${cmd}" "${expected}"

# Let us test - now let us test the messageCount (it relates to both topics)
expected=("Message count:
Started Kafka event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Sending Kafka notification (protobuf)
Sending 0 sync Kafka notifications (protobuf) ...
Sending 0 async Kafka notifications (protobuf) ...
")

cmd='examples/kafka-plugin/hash-partitioner/hash-partitioner --kafka-config examples/kafka-plugin/hash-partitioner/kafka.conf  -messageCount=0'
testOutput "${cmd}" "${expected}"

# Let us test - now let us test the messageCount (it relates to both topics)
expected=("Message count:
Started Kafka event handler...
Started Kafka async event handler...
Starting message handlers for new consumer ...
All plugins initialized successfully
Sending Kafka notification (protobuf)
Sending 1 sync Kafka notifications (protobuf) ...
Received Kafka Message, topic 'example-sync-topic', partition '0', offset '10', key: 'proto-key', 
Sending 1 async Kafka notifications (protobuf) ...
Async message successfully delivered, topic 'example-async-topic', partition '0', offset '10', key: 'async-proto-key', 
Received async Kafka Message, topic 'example-async-topic', partition '0', offset '10', key: 'async-proto-key', 
")

cmd='examples/kafka-plugin/hash-partitioner/hash-partitioner --kafka-config examples/kafka-plugin/hash-partitioner/kafka.conf  -messageCount=1'
testOutput "${cmd}" "${expected}"



stopKafka

#### Kafka-plugin post-init-consumer #######################################

#startKafka
startCustomizedKafka examples/kafka-plugin/manual-partitioner/server.properties

expected=("Usage of examples/kafka-plugin/post-init-consumer/post-init-consumer:
-config-dir=\".\": Location of the configuration files; also set via 'CONFIG_DIR' env variable.
-kafka-config=\"kafka.conf\": Location of the kafka Client configuration file; also set via 'KAFKA_CONFIG' env variable.
-logs-config=\"logs.conf\": Location of the configuration files; also set via 'LOGS_CONFIG' env variable.
-microservice-label=\"vpp1\": microservice label; also set via 'MICROSERVICE_LABEL' env variable.
")

cmd="examples/kafka-plugin/post-init-consumer/post-init-consumer --kafka-config examples/kafka-plugin/post-init-consumer/kafka.conf -h"
testOutput "${cmd}" "${expected}"

expected=("All plugins initialized successfully
Starting 'post-init' manual Consumer
Sending 10 Kafka notifications (protobuf) ...
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '0', key: 'proto-key', 
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '1', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '2', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '3', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '4', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '5', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '6', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '7', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '8', key: 'proto-key',
Received sync Kafka Message, topic 'example-sync-topic', partition '1', offset '9', key: 'proto-key',
")

cmd="examples/kafka-plugin/post-init-consumer/post-init-consumer --kafka-config examples/kafka-plugin/post-init-consumer/kafka.conf"
testOutput "${cmd}" "${expected}"

stopKafka

#### Logs-lib ############################################################

expected=("Started observing beach
A group of walrus emerges from the ocean
The group's number increased tremendously!
Temperature changes
It's over 9000!
The ice breaks!
")
testOutput examples/logs-lib/basic/basic "${expected}"

expected=("DEBUG componentXY
WARN componentXY
ERROR componentXY
")
testOutput examples/logs-lib/custom/custom "${expected}"

#### Logs-plugin #########################################################

expected=("Debug log example
Info log example
Warn log example
Error log example
Stopping agent...
")

testOutput examples/logs-plugin/logs-plugin "${expected}"

#### Simple-agent ########################################################

expected=("etcd config not found  - skip loading this plugin
kafka config not found  - skip loading this plugin
redis config not found  - skip loading this plugin
cassandra client config not found  - skip loading this plugin
All plugins initialized successfully
")

testOutput examples/simple-agent/simple-agent "${expected}" 5

#### Simple-agent with Kafka and ETCD ####################################

startEtcd
startKafka

expected=("Plugin etcdv3: status check probe registered
Plugin kafka: status check probe registered
redis config not found  - skip loading this plugin
cassandra client config not found  - skip loading this plugin
All plugins initialized successfully
")

cmd="examples/simple-agent/simple-agent --etcdv3-config=examples/datasync-plugin/etcd.conf --kafka-config examples/kafka-plugin/hash-partitioner/kafka.conf"
testOutput "${cmd}" "${expected}" 5

stopEtcd
stopKafka

##########################################################################

exit ${exitCode}
