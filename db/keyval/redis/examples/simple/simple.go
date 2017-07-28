package main

import (
	"os"
	"time"

	"fmt"
	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/utils/config"
	"regexp"
	"strconv"
	"strings"
)

var usage = `usage: %s -n|-c|-s <client.yaml>
	where
	-n		Specifies the use of a node client
	-c		Specifies the use of a cluster client
	-s		Specifies the use of a sentinel client
`

var log = logroot.Logger()

var redisConn *redis.BytesConnectionRedis
var broker keyval.BytesBroker
var watcher keyval.BytesWatcher

var prefix string
var useKeys string
var useRedigo = false

func main() {
	//generateSampleConfigs()

	cfg := loadConfig()
	if cfg == nil {
		return
	}
	fmt.Printf("config: %T:\n%v\n", cfg, cfg)
	fmt.Printf("prefix: %s\n", prefix)
	fmt.Printf("useKeys: %s\n", useKeys)
	fmt.Printf("useRedigo: %t\n", useRedigo)

	if useRedigo {
		redisConn = createConnectionRedigo(cfg)
	} else {
		redisConn = createConnection(cfg)
		if useKeys != "" {
			redisConn.UseKeysCmdForCluster, _ = strconv.ParseBool(useKeys)
		}
	}

	broker = redisConn.NewBroker(prefix)
	watcher = redisConn.NewWatcher(prefix)

	runSimpleExmple()
}

func loadConfig() interface{} {
	numArgs := len(os.Args)
	defer func() {
		// Variety to run the example
		if numArgs > 3 {
			rePrefix := regexp.MustCompile("prefix:.*")
			reKeys := regexp.MustCompile("keys:.*")
			for _, a := range os.Args[3:] {
				switch a {
				case "redigo":
					useRedigo = true
					fmt.Println("Using redigo")
					continue
				case "debug":
					log.SetLevel(logging.DebugLevel)
					continue
				}
				if rePrefix.MatchString(a) {
					prefix = strings.TrimPrefix(a, "prefix:")
				} else if reKeys.MatchString(a) {
					useKeys = strings.TrimPrefix(a, "keys:")
				}
			}
		}
	}()

	if numArgs < 3 {
		fmt.Printf(usage, os.Args[0])
		return nil
	}
	var err error
	t := os.Args[1]
	f := os.Args[2]
	defer func() {
		if err != nil {
			log.Panicf("ParseConfigFromYamlFile(%s) failed: %s", f, err)
		}
	}()

	switch t {
	case "-n":
		var cfg redis.NodeConfig
		err = config.ParseConfigFromYamlFile(f, &cfg)
		return cfg
	case "-c":
		var cfg redis.ClusterConfig
		err = config.ParseConfigFromYamlFile(f, &cfg)
		return cfg
	case "-s":
		var cfg redis.SentinelConfig
		err = config.ParseConfigFromYamlFile(f, &cfg)
		return cfg
	}
	return nil
}

func createConnection(cfg interface{}) *redis.BytesConnectionRedis {
	client, err := redis.CreateClient(cfg)
	if err != nil {
		log.Panicf("CreateNodeClient() failed: %s", err)
	}
	conn, err := redis.NewBytesConnection(client, log)
	if err != nil {
		client.Close()
		log.Panicf("NewBytesConnection() failed: %s", err)
	}
	return conn
}

func createConnectionRedigo(cfg interface{}) *redis.BytesConnectionRedis {
	pool, err := redis.CreateNodeClientConnPool(cfg.(redis.NodeConfig))
	if err != nil {
		log.Panicf("CreateNodeClientConnPool() failed: %s", err)
	}
	conn, err := redis.NewBytesConnectionRedis(pool, log)
	if err != nil {
		pool.Close()
		log.Panicf("NewBytesConnectionRedigo() failed: %s", err)
	}
	return conn
}

func runSimpleExmple() {
	var err error

	keyPrefix := "key"
	keys3 := []string{
		keyPrefix + "1",
		keyPrefix + "2",
		keyPrefix + "3",
	}

	respChan := make(chan keyval.BytesWatchResp, 10)
	err = watcher.Watch(respChan, keyPrefix)
	if err != nil {
		log.Errorf(err.Error())
	}
	go func() {
		for {
			select {
			case r, ok := <-respChan:
				if ok {
					switch r.GetChangeType() {
					case db.Put:
						log.Infof("Watcher received %v: %s=%s", r.GetChangeType(), r.GetKey(), string(r.GetValue()))
					case db.Delete:
						log.Infof("Watcher received %v: %s", r.GetChangeType(), r.GetKey())
					}
				} else {
					log.Error("Something wrong with respChan... bail out")
					return
				}
			default:
				break
			}
		}
	}()
	time.Sleep(2 * time.Second)
	put(keys3[0], "val 1")
	put(keys3[1], "val 2")
	put(keys3[2], "val 3", keyval.WithTTL(time.Second))

	time.Sleep(2 * time.Second)
	get(keys3[0])
	get(keys3[1])
	fmt.Printf("==> NOTE: %s should have expired\n", keys3[2])
	get(keys3[2]) // key3 should've expired
	fmt.Printf("==> NOTE: get(%s) should return false\n", keyPrefix)
	get(keyPrefix) // keyPrefix shouldn't find anything
	listKeys(keyPrefix)
	listVal(keyPrefix)

	del(keyPrefix)

	fmt.Println("==> NOTE: All keys should have been deleted")
	get(keys3[0])
	get(keys3[1])
	listKeys(keyPrefix)
	listVal(keyPrefix)

	txn(keyPrefix)

	log.Info("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)

	// Done watching.  Close the channel.
	log.Infof("Closing connection")
	//close(respChan)
	redisConn.Close()

	fmt.Println("==> NOTE: Call on a closed connection should fail.")
	del(keyPrefix)

	log.Info("Sleep for 10 seconds")
	time.Sleep(30 * time.Second)
}

func put(key, value string, opts ...keyval.PutOption) {
	err := broker.Put(key, []byte(value), opts...)
	if err != nil {
		//log.Panicf(err.Error())
		log.Errorf(err.Error())
	}
}

func get(key string) {
	var val []byte
	var found bool
	var revision int64
	var err error

	val, found, revision, err = broker.GetValue(key)
	if err != nil {
		log.Errorf(err.Error())
	} else if found {
		log.Infof("GetValue(%s) = %t ; val = %s ; revision = %d", key, found, val, revision)
	} else {
		log.Infof("GetValue(%s) = %t", key, found)
	}
}

func listKeys(keyPrefix string) {
	var keys keyval.BytesKeyIterator
	var err error

	keys, err = broker.ListKeys(keyPrefix)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		var count int32
		for {
			key, rev, done := keys.GetNext()
			if done {
				break
			}
			log.Infof("ListKeys(%s):  %s (rev %d)", keyPrefix, key, rev)
			count++
		}
		log.Infof("ListKeys(%s): count = %d", keyPrefix, count)
	}
}

func listVal(keyPrefix string) {
	var keyVals keyval.BytesKeyValIterator
	var err error

	keyVals, err = broker.ListValues(keyPrefix)
	if err != nil {
		log.Errorf(err.Error())
	} else {
		var count int32
		for {
			kv, done := keyVals.GetNext()
			if done {
				break
			}
			log.Infof("ListValues(%s):  %s = %s (rev %d)", keyPrefix, kv.GetKey(), kv.GetValue(), kv.GetRevision())
			count++
		}
		log.Infof("ListValues(%s): count = %d", keyPrefix, count)
	}
}

func del(keyPrefix string) {
	var found bool
	var err error

	found, err = broker.Delete(keyPrefix)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("Delete(%s): found = %t", keyPrefix, found)
}

func txn(keyPrefix string) {
	keys := []string{
		keyPrefix + "101",
		keyPrefix + "102",
		keyPrefix + "103",
		keyPrefix + "104",
	}
	var txn keyval.BytesTxn

	log.Infof("txn(): keys = %v", keys)
	txn = broker.NewTxn()
	for i, k := range keys {
		txn.Put(k, []byte(strconv.Itoa(i+1)))
	}
	txn.Delete(keys[0])
	err := txn.Commit()
	if err != nil {
		log.Errorf("txn(): %s", err)
	}
	listVal(keyPrefix)
}

func generateSampleConfigs() {
	clientConfig := redis.ClientConfig{
		Password:     "",
		DialTimeout:  0,
		ReadTimeout:  0,
		WriteTimeout: 0,
		Pool: redis.PoolConfig{
			PoolSize:           0,
			PoolTimeout:        0,
			IdleTimeout:        0,
			IdleCheckFrequency: 0,
		},
	}
	redis.GenerateConfig(
		&redis.NodeConfig{
			Endpoint: "localhost:6379",
			DB:       0,
			EnableReadQueryOnSlave: false,
			TLS:          redis.TLS{},
			ClientConfig: clientConfig,
		}, "./node-client.yaml")
	redis.GenerateConfig(
		&redis.ClusterConfig{
			Endpoints:              []string{"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003"},
			EnableReadQueryOnSlave: true,
			MaxRedirects:           0,
			RouteByLatency:         true,
			ClientConfig:           clientConfig,
		}, "./cluster-client.yaml")
	redis.GenerateConfig(
		&redis.SentinelConfig{
			Endpoints:    []string{"localhost:26379"},
			MasterName:   "mymaster",
			DB:           0,
			ClientConfig: clientConfig,
		}, "./sentinel-client.yaml")
}
