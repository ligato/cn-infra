// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	"errors"
	"testing"
	"time"

	"strconv"

	"os"

	"github.com/alicebob/miniredis"
	redigo "github.com/garyburd/redigo/redis"
	goredis "github.com/go-redis/redis"
	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/onsi/gomega"
	"github.com/rafaeljusto/redigomock"
)

var miniRedis *miniredis.Miniredis
var mockConn *redigomock.Conn
var mockPool *redigo.Pool
var bytesConn *BytesConnectionRedis
var bytesBrokerWatcher *BytesBrokerWatcherRedis
var iKeys = []interface{}{}
var iVals = []interface{}{}
var iAll = []interface{}{}
var ttl = time.Second
var log logging.Logger

var keyValues = map[string]string{
	"keyWest": "a place",
	"keyMap":  "a map",
}

var useRedigo = false

func TestMain(m *testing.M) {
	log = logroot.Logger()

	var code int
	mockConnectionRedigo()
	code = m.Run()

	var err error
	miniRedis, err = miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer miniRedis.Close()
	createConnectionMiniRedis()
	code = m.Run()

	os.Exit(code)
}

func mockConnectionRedigo() {
	useRedigo = true
	mockConn = redigomock.NewConn()
	for k, v := range keyValues {
		mockConn.Command("SET", k, v).Expect("not used")
		mockConn.Command("SET", k, v, "PX", ttl/time.Millisecond).Expect("not used")
		mockConn.Command("GET", k).Expect(v)
		iKeys = append(iKeys, k)
		iVals = append(iVals, v)
		iAll = append(append(iAll, k), v)
	}
	mockConn.Command("GET", "key").Expect(nil)
	mockConn.Command("GET", "bytes").Expect([]byte("bytes"))
	mockConn.Command("GET", "nil").Expect(nil)

	mockConn.Command("MGET", iKeys...).Expect(iVals)
	mockConn.Command("DEL", iKeys...).Expect(len(keyValues))

	mockConn.Command("MSET", []interface{}{"keyWest", keyValues["keyWest"], "keyMap", keyValues["keyMap"]}...).Expect(nil)
	mockConn.Command("DEL", []interface{}{"keyWest"}...).Expect(1).Expect(nil)

	// for negative tests
	manufacturedError := errors.New("manufactured error")
	mockConn.Command("SET", "error", "error").ExpectError(manufacturedError)
	mockConn.Command("GET", "error").ExpectError(manufacturedError)
	mockConn.Command("GET", "redisError").Expect(redigo.Error("Blah"))
	mockConn.Command("GET", "unknown").Expect(struct{}{})
	mockPool = &redigo.Pool{
		Dial: func() (redigo.Conn, error) { return mockConn, nil },
	}
	bytesConn, _ = NewBytesConnectionRedis(mockPool, logroot.Logger())
	bytesBrokerWatcher = bytesConn.NewBrokerWatcher("")
}

func createConnectionMiniRedis() {
	useRedigo = false
	for k, v := range keyValues {
		miniRedis.Set(k, v)
		iKeys = append(iKeys, k)
		iVals = append(iVals, v)
		iAll = append(append(iAll, k), v)
	}
	miniRedis.Set("bytes", "bytes")

	clientConfig := ClientConfig{
		Password:     "",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
		Pool: PoolConfig{
			PoolSize:           0,
			PoolTimeout:        0,
			IdleTimeout:        0,
			IdleCheckFrequency: 0,
		},
	}
	nodeConfig := NodeConfig{
		Endpoint: miniRedis.Addr(),
		DB:       0,
		AllowReadQueryToSlave: false,
		TLS:          TLS{},
		ClientConfig: clientConfig,
	}
	var client Client
	client = goredis.NewClient(&goredis.Options{
		Network: "tcp",
		Addr:    nodeConfig.Endpoint,

		// Database to be selected after connecting to the server
		DB: nodeConfig.DB,

		// Enables read only queries on slave nodes.
		ReadOnly: nodeConfig.AllowReadQueryToSlave,

		// TLS Config to use. When set TLS will be negotiated.
		TLSConfig: nil,

		// Optional password. Must match the password specified in the requirepass server configuration option.
		Password: nodeConfig.Password,

		// Dial timeout for establishing new connections. Default is 5 seconds.
		DialTimeout: nodeConfig.DialTimeout * time.Second,
		// Timeout for socket reads. If reached, commands will fail with a timeout instead of blocking. Default is 3 seconds.
		ReadTimeout: nodeConfig.ReadTimeout * time.Second,
		// Timeout for socket writes. If reached, commands will fail with a timeout instead of blocking. Default is ReadTimeout.
		WriteTimeout: nodeConfig.WriteTimeout * time.Second,

		// Maximum number of socket connections. Default is 10 connections per every CPU as reported by runtime.NumCPU.
		PoolSize: nodeConfig.Pool.PoolSize,
		// Amount of time client waits for connection if all connections are busy before returning an error. Default is ReadTimeout + 1 second.
		PoolTimeout: nodeConfig.Pool.PoolTimeout * time.Second,
		// Amount of time after which client closes idle connections. Should be less than server's timeout. Default is 5 minutes.
		IdleTimeout: nodeConfig.Pool.IdleTimeout * time.Second,
		// Frequency of idle checks. Default is 1 minute. When minus value is set, then idle check is disabled.
		IdleCheckFrequency: nodeConfig.Pool.IdleCheckFrequency,

		// Dialer creates new network connection and has priority over Network and Addr options.
		// Dialer func() (net.Conn, error)
		// Hook that is called when new connection is established
		// OnConnect func(*Conn) error

		// Maximum number of retries before giving up. Default is to not retry failed commands.
		MaxRetries: 0,
		// Minimum backoff between each retry. Default is 8 milliseconds; -1 disables backoff.
		MinRetryBackoff: 0,
		// Maximum backoff between each retry. Default is 512 milliseconds; -1 disables backoff.
		MaxRetryBackoff: 0,
	})
	// client = &MockGoredisClient{}
	bytesConn, _ = NewBytesConnection(client, logroot.Logger())
	bytesBrokerWatcher = bytesConn.NewBrokerWatcher("")
}

///////////////////////////////////////////////////////////////////////////////
// Redigo - https://github.com/garyburd/redigo/redis

func TestConnection(t *testing.T) {
	gomega.RegisterTestingT(t)

	client, err := CreateClient(NodeConfig{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(client).ShouldNot(gomega.BeNil())

	client, err = CreateClient(ClusterConfig{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(client).ShouldNot(gomega.BeNil())

	client, err = CreateClient(SentinelConfig{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(client).ShouldNot(gomega.BeNil())

	var cfg *NodeConfig
	client, err = CreateClient(cfg)
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(client).Should(gomega.BeNil())
	client, err = CreateClient(nil)
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(client).Should(gomega.BeNil())
}

func TestBrokerWatcher(t *testing.T) {
	gomega.RegisterTestingT(t)
	prefix := bytesBrokerWatcher.GetPrefix()
	gomega.Expect(prefix).ShouldNot(gomega.BeNil())

	broker := bytesConn.NewBroker("")
	gomega.Expect(broker).Should(gomega.BeAssignableToTypeOf(bytesBrokerWatcher))

	watcher := bytesConn.NewWatcher("")
	gomega.Expect(watcher).Should(gomega.BeAssignableToTypeOf(bytesBrokerWatcher))
}

func TestPut(t *testing.T) {
	gomega.RegisterTestingT(t)
	err := bytesBrokerWatcher.Put("keyWest", []byte(keyValues["keyWest"]))
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	err = bytesBrokerWatcher.Put("keyWest", []byte(keyValues["keyWest"]), keyval.WithTTL(ttl))
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestGet(t *testing.T) {
	gomega.RegisterTestingT(t)
	val, found, _, err := bytesBrokerWatcher.GetValue("keyWest")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeTrue())
	gomega.Expect(val).Should(gomega.Equal([]byte(keyValues["keyWest"])))

	val, found, _, err = bytesBrokerWatcher.GetValue("bytes")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeTrue())
	gomega.Expect(val).ShouldNot(gomega.BeNil())

	val, found, _, err = bytesBrokerWatcher.GetValue("nil")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())
}

func TestListValues(t *testing.T) {
	gomega.RegisterTestingT(t)

	if useRedigo {
		// Implicitly call SCAN (or previous, "KEYS")
		mockConn.Command("SCAN", "0", "MATCH", "key*").Expect([]interface{}{[]byte("0"), iKeys})
		//mockConn.Command("KEYS", "key*").Expect(iKeys)
	}
	keyVals, err := bytesConn.ListValues("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	for {
		kv, last := keyVals.GetNext()
		if last {
			break
		}
		gomega.Expect(kv.GetKey()).Should(gomega.SatisfyAny(gomega.BeEquivalentTo("keyWest"), gomega.BeEquivalentTo("keyMap")))
		gomega.Expect(kv.GetValue()).Should(gomega.SatisfyAny(gomega.BeEquivalentTo(keyValues["keyWest"]), gomega.BeEquivalentTo(keyValues["keyMap"])))
		gomega.Expect(kv.GetRevision()).ShouldNot(gomega.BeNil())
	}

	if useRedigo {
		// Implicitly call SCAN (or previous, "KEYS")
		mockConn.Command("SCAN", "0", "MATCH", "key*").Expect([]interface{}{[]byte("0"), iKeys})
		//mockConn.Command("KEYS", "key*").Expect(iKeys)
	}
	keyVals, err = bytesBrokerWatcher.ListValues("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	for {
		kv, last := keyVals.GetNext()
		if last {
			break
		}
		gomega.Expect(kv.GetKey()).Should(gomega.SatisfyAny(gomega.BeEquivalentTo("keyWest"), gomega.BeEquivalentTo("keyMap")))
		gomega.Expect(kv.GetValue()).Should(gomega.SatisfyAny(gomega.BeEquivalentTo(keyValues["keyWest"]), gomega.BeEquivalentTo(keyValues["keyMap"])))
		gomega.Expect(kv.GetRevision()).ShouldNot(gomega.BeNil())
	}
}

func TestListKeys(t *testing.T) {
	if useRedigo {
		// Each SCAN (or previous, "KEYS") is set on demand in individual test that calls it.
		mockConn.Command("SCAN", "0", "MATCH", "key*").Expect([]interface{}{[]byte("0"), iKeys})
		//mockConn.Command("KEYS", "key*").Expect(iKeys)
	}

	gomega.RegisterTestingT(t)
	keys, err := bytesBrokerWatcher.ListKeys("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	for {
		k, _, last := keys.GetNext()
		if last {
			break
		}
		gomega.Expect(k).Should(gomega.SatisfyAny(gomega.BeEquivalentTo("keyWest"), gomega.BeEquivalentTo("keyMap")))
	}
}

func TestDel(t *testing.T) {
	if useRedigo {
		// Implicitly call SCAN (or previous, "KEYS")
		mockConn.Command("SCAN", "0", "MATCH", "key*").Expect([]interface{}{[]byte("0"), iKeys})
		//mockConn.Command("KEYS", "key*").Expect(iKeys)
	}

	gomega.RegisterTestingT(t)
	found, err := bytesBrokerWatcher.Delete("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeTrue())
}

func TestTxn(t *testing.T) {
	gomega.RegisterTestingT(t)
	txn := bytesBrokerWatcher.NewTxn()
	txn.Put("keyWest", []byte(keyValues["keyWest"])).Put("keyMap", []byte(keyValues["keyMap"]))
	txn.Delete("keyWest")
	err := txn.Commit()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	if !useRedigo {
		val, found, _, err := bytesBrokerWatcher.GetValue("keyWest")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(found).Should(gomega.BeFalse())
		gomega.Expect(val).Should(gomega.BeNil())

		val, found, _, err = bytesBrokerWatcher.GetValue("keyMap")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(found).Should(gomega.BeTrue())
		gomega.Expect(val).Should(gomega.Equal([]byte(keyValues["keyMap"])))
	}
	txn = bytesBrokerWatcher.NewTxn()
	txn.Put("{hashTag}key", []byte{}).Delete("keyWest")
	checkCrossSlot(txn.(*Txn))
}

func TestWatchRedigo(t *testing.T) {
	if !useRedigo {
		return
	}

	gomega.RegisterTestingT(t)

	respChan := make(chan keyval.BytesWatchResp)
	var count int

	key := iKeys[0]
	mockConn.Command("PSUBSCRIBE", []interface{}{keySpaceEventPrefix + "key*"}...).Expect(newSubscriptionResponse("psubscribe", keySpaceEventPrefix+"key*", 1))
	mockConn.Command("GET", key).Expect(keyValues[key.(string)])
	count = 0
	mockConn.AddSubscriptionMessage(newPMessage(keySpaceEventPrefix+"key*", keySpaceEventPrefix+key.(string), "set"))
	count++
	mockConn.AddSubscriptionMessage(newPMessage(keySpaceEventPrefix+"key*", keySpaceEventPrefix+key.(string), "del"))
	count++
	bytesConn.Watch(respChan, "key")
	consumeEvent(respChan, count)

	key = iKeys[1]
	mockConn.Command("PSUBSCRIBE", []interface{}{keySpaceEventPrefix + "key*"}...).Expect(newSubscriptionResponse("psubscribe", keySpaceEventPrefix+"key*", 1))
	mockConn.Command("GET", key).Expect(keyValues[key.(string)])
	count = 0
	mockConn.AddSubscriptionMessage(newPMessage(keySpaceEventPrefix+"key*", keySpaceEventPrefix+key.(string), "set"))
	count++
	mockConn.AddSubscriptionMessage(newPMessage(keySpaceEventPrefix+"key*", keySpaceEventPrefix+key.(string), "del"))
	count++
	bytesBrokerWatcher.Watch(respChan, "key")
	consumeEvent(respChan, count)
}

func consumeEvent(respChan chan keyval.BytesWatchResp, eventCount int) {
	for {
		r, ok := <-respChan
		if ok {
			switch r.GetChangeType() {
			case db.Put:
				log.Debugf("Watcher received %v: %s=%s (rev %d)",
					r.GetChangeType(), r.GetKey(), string(r.GetValue()), r.GetRevision())
			case db.Delete:
				log.Debugf("Watcher received %v: %s (rev %d)",
					r.GetChangeType(), r.GetKey(), r.GetRevision())
				r.GetValue()
			}
		} else {
			log.Error("Something wrong with Watch channel... bail out")
			break
		}
		eventCount--
		if eventCount == 0 {
			return
		}
	}
}

func newSubscriptionResponse(kind string, chanName string, count int) []interface{} {
	values := []interface{}{}
	values = append(values, interface{}([]byte(kind)))
	values = append(values, interface{}([]byte(chanName)))
	values = append(values, interface{}([]byte(strconv.Itoa(count))))
	return values
}

func newPMessage(pattern string, chanName string, data string) []interface{} {
	values := []interface{}{}
	values = append(values, interface{}([]byte("pmessage")))
	values = append(values, interface{}([]byte(pattern)))
	values = append(values, interface{}([]byte(chanName)))
	values = append(values, interface{}([]byte(data)))
	return values
}

func TestGetShouldNotApplyWildcard(t *testing.T) {
	gomega.RegisterTestingT(t)
	val, found, _, err := bytesBrokerWatcher.GetValue("key")
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())
}

func TestPutError(t *testing.T) {
	if !useRedigo {
		return
	}
	gomega.RegisterTestingT(t)
	err := bytesBrokerWatcher.Put("error", []byte("error"))
	gomega.Expect(err).Should(gomega.HaveOccurred())
}

func TestGetError(t *testing.T) {
	if !useRedigo {
		return
	}
	gomega.RegisterTestingT(t)
	val, found, _, err := bytesBrokerWatcher.GetValue("error")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())

	val, found, _, err = bytesBrokerWatcher.GetValue("redisError")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())

	val, found, _, err = bytesBrokerWatcher.GetValue("unknown")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	gomega.Expect(found).Should(gomega.BeFalse())
	gomega.Expect(val).Should(gomega.BeNil())
}

func TestBrokerClosed(t *testing.T) {
	gomega.RegisterTestingT(t)

	txn := bytesConn.NewTxn()
	txn2 := bytesBrokerWatcher.NewTxn()
	err := bytesConn.Close()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	respChan := make(chan keyval.BytesWatchResp)

	// byteConn
	err = bytesConn.Put("any", []byte("any"))
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, _, _, err = bytesConn.GetValue("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesConn.ListValues("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesConn.ListKeys("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesConn.Delete("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())

	txn.Put("keyWest", []byte(keyValues["keyWest"])).Put("keyMap", []byte(keyValues["keyMap"]))
	txn.Delete("keyWest")
	err = txn.Commit()
	gomega.Expect(err).Should(gomega.HaveOccurred())

	txn = bytesConn.NewTxn()
	gomega.Expect(txn).Should(gomega.BeNil())

	bytesConn.Watch(respChan, "key")

	// bytesBrokerWatcher
	err = bytesBrokerWatcher.Put("any", []byte("any"))
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, _, _, err = bytesBrokerWatcher.GetValue("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesBrokerWatcher.ListValues("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesBrokerWatcher.ListKeys("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())
	_, err = bytesBrokerWatcher.Delete("any")
	gomega.Expect(err).Should(gomega.HaveOccurred())

	txn2.Put("keyWest", []byte(keyValues["keyWest"])).Put("keyMap", []byte(keyValues["keyMap"]))
	txn2.Delete("keyWest")
	err = txn2.Commit()
	gomega.Expect(err).Should(gomega.HaveOccurred())

	txn2 = bytesBrokerWatcher.NewTxn()
	gomega.Expect(txn2).Should(gomega.BeNil())

	bytesBrokerWatcher.Watch(respChan, "key")

	err = bytesConn.Close()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
}

///////////////////////////////////////////////////////////////////////////////
// go-redis https://github.com/go-redis/redis

type MockGoredisClient struct {
}

func (c *MockGoredisClient) Close() error {
	return nil
}
func (c *MockGoredisClient) Del(keys ...string) *goredis.IntCmd {
	args := stringsToInterfaces(append([]string{"del"}, keys...)...)
	cmd := goredis.NewIntCmd(args...)
	//TODO: Manipulate command result here...
	return cmd
}
func (c *MockGoredisClient) Get(key string) *goredis.StringCmd {
	cmd := goredis.NewStringCmd("get", key)
	//TODO: Manipulate command result here...
	return cmd
}
func (c *MockGoredisClient) MGet(keys ...string) *goredis.SliceCmd {
	args := stringsToInterfaces(append([]string{"mget"}, keys...)...)
	cmd := goredis.NewSliceCmd(args...)
	//TODO: Manipulate command result here...
	return cmd
}
func (c *MockGoredisClient) Scan(cursor uint64, match string, count int64) *goredis.ScanCmd {
	args := []interface{}{"scan", cursor}
	if match != "" {
		args = append(args, "match", match)
	}
	if count > 0 {
		args = append(args, "count", count)
	}
	cmd := goredis.NewScanCmd(func(cmd goredis.Cmder) error {
		//TODO: Manipulate command result here...
		return nil
	}, args...)
	return cmd
}
func (c *MockGoredisClient) Set(key string, value interface{}, expiration time.Duration) *goredis.StatusCmd {
	args := make([]interface{}, 3, 4)
	args[0] = "set"
	args[1] = key
	args[2] = value
	if expiration > 0 {
		if expiration < time.Second || expiration%time.Second != 0 {
			args = append(args, "px", expiration/time.Millisecond)
		} else {
			args = append(args, "ex", expiration/time.Second)
		}
	}
	cmd := goredis.NewStatusCmd(args...)
	//TODO: Manipulate command result here...
	return cmd
}
func (c *MockGoredisClient) TxPipeline() goredis.Pipeliner {
	//TODO: Manipulate the pipeliner...
	return &goredis.Pipeline{}
}
func (c *MockGoredisClient) PSubscribe(channels ...string) *goredis.PubSub {
	pubSub := &goredis.PubSub{}
	//TODO: PubSub is a struct with all internal fields. Is it possible to manipulate?
	return pubSub
}

func stringsToInterfaces(ss ...string) []interface{} {
	args := make([]interface{}, len(ss))
	for i, s := range ss {
		args[i] = s
	}
	return args
}
