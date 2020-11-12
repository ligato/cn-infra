//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package config_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.ligato.io/cn-infra/v2/config"
)

type TestCfg struct {
	Label   string
	Debug   bool
	Port    int
	Rate    float64
	Timeout time.Duration
	List    []string
	Levels  map[string]string
}

func TestGetters(t *testing.T) {
	const cfgData = `
label: MyLabel
debug: true
port: 4444
rate: 1.2
timeout: 3s
list: [a, b, c]
levels:
  api: debug
  server: fatal
`
	conf := config.NewConfig()

	err := conf.ReadFrom(strings.NewReader(cfgData))
	if err != nil {
		t.Fatal(err)
	}
	if len(conf.All()) == 0 {
		t.Fatal("loaded config is empty")
	}
	t.Logf("All: %+v", conf.All())

	t.Run("getters", func(t *testing.T) {
		label := conf.GetString("label")
		if label != "MyLabel" {
			t.Errorf("label expected: %q, got: %q", "MyLabel", label)
		}
		debug := conf.GetBool("debug")
		if debug != true {
			t.Errorf("debug expected: %v, got: %v", true, debug)
		}
		port := conf.GetInt("port")
		if port != 4444 {
			t.Errorf("port expected: %v, got: %v", 4444, port)
		}
		rate := conf.GetFloat64("rate")
		if rate != 1.2 {
			t.Errorf("rate expected: %v, got: %v", 1.2, rate)
		}
		timeout := conf.GetDuration("timeout")
		if timeout != time.Second*3 {
			t.Errorf("timeout expected: %v, got: %v", time.Second*3, timeout)
		}
		list := conf.GetStringSlice("list")
		expectList := []string{"a", "b", "c"}
		if len(list) != len(expectList) {
			t.Errorf("list length expected: %v, got: %v", len(expectList), list)
		}
		if str := fmt.Sprint(list); str != fmt.Sprint(expectList) {
			t.Errorf("list expected: %v, got: %v", expectList, list)
		}
		levels := conf.GetStringMapString("levels")
		expectLevels := map[string]string{"api": "debug", "server": "fatal"}
		if len(levels) != len(expectLevels) {
			t.Errorf("levels length expected: %v, got: %v", len(expectLevels), levels)
		}
		if str := fmt.Sprint(levels); str != fmt.Sprint(expectLevels) {
			t.Errorf("levels expected: %v, got: %v", expectLevels, levels)
		}
	})

	t.Run("unmarshal", func(t *testing.T) {

		var cfg TestCfg

		err = conf.Unmarshal(&cfg)
		if err != nil {
			t.Fatal(err)
		}

		expectedCfg := TestCfg{
			Label:   "MyLabel",
			Debug:   true,
			Port:    4444,
			Rate:    1.2,
			Timeout: time.Second * 3,
			List:    []string{"a", "b", "c"},
			Levels: map[string]string{
				"api":    "debug",
				"server": "fatal",
			},
		}
		if !reflect.DeepEqual(cfg, expectedCfg) {
			t.Fatalf("cfg:\nexpected:\n%+v\ngot:\n%+v", expectedCfg, cfg)
		}
	})
}

func TestUnmarshal(t *testing.T) {
	const cfgData = `
label: MyLabel
debug: true
port: 4444
rate: 1.2
timeout: 3s
list: [a, b, c]
levels:
  api: debug
  server: fatal
#service:
#  endpoint: 10.10.1.2:3333
`
	conf := config.NewConfig()

	err := conf.ReadFrom(strings.NewReader(cfgData))
	if err != nil {
		t.Fatal(err)
	}
	if len(conf.All()) == 0 {
		t.Fatal("loaded config is empty")
	}

	t.Logf("All: %+v", conf.All())

}
