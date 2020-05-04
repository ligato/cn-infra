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

package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"

	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"
)

// Config defines an API for handling configuration.
type Config interface {
	Read() error
	ReadFrom(io.Reader) error
	MergeFrom(io.Reader) error
	MergeMap(map[string]interface{}) error
	WriteAs(filename string) error

	SetDefault(key string, val interface{})
	BindEnv(key string, env string) error
	BindFlag(key string, flag *Flag) error
	Set(key string, val interface{})

	Sub(key string) Config
	All() map[string]interface{}
	Unmarshal(cfg interface{}) error
	UnmarshalKey(key string, cfg interface{}) error

	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
}

// NewConfig returns fresh new Config instance.
func NewConfig() Config {
	v := viper.New()
	return newConf(v)
}

// NewConfigFromFile returns new Config instance that reads from given file.
func NewConfigFromFile(file string) Config {
	v := viper.New()
	v.SetConfigFile(file)
	return newConf(v)
}

type config struct {
	*viper.Viper
}

func newConf(v *viper.Viper) *config {
	return &config{Viper: v}
}

func (c *config) Read() error {
	configDir := c.Viper.GetString(DirFlag)

	logger.Debugf("adding config path: %q", configDir)
	c.Viper.AddConfigPath(configDir)

	if err := c.Viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			logger.Debugf("ReadInConfig() error: %T %v", err, err)
			return err
		}
	}
	logger.Debugf("loaded config from: %q", c.Viper.ConfigFileUsed())

	if logger.GetLevel() >= logging.TraceLevel {
		c.Debug()
	}

	return nil
}

func (c *config) ReadFrom(r io.Reader) error {
	return c.Viper.ReadConfig(r)
}

func (c *config) MergeFrom(r io.Reader) error {
	return c.Viper.MergeConfig(r)
}

func (c *config) MergeMap(m map[string]interface{}) error {
	return c.Viper.MergeConfigMap(m)
}

func (c *config) WriteAs(filename string) error {
	logger.Tracef("config.WriteAs %q", filename)
	return c.Viper.WriteConfigAs(filename)
}

func (c *config) BindEnv(key string, env string) error {
	logger.Tracef("bind config key %q to ENV %q", key, env)
	err := c.Viper.BindEnv(key, env)
	if err != nil {
		logger.Debugf("ERROR binding to env: %v", err)
	}
	return err
}

func (c *config) BindFlag(key string, flag *Flag) error {
	logger.Tracef("bind config key %q to FLAG %q", key, flag.Name)
	err := c.Viper.BindPFlag(key, flag)
	if err != nil {
		logger.Debugf("ERROR binding to flag: %v", err)
	}
	return err
}

func (c *config) Sub(key string) Config {
	logger.Tracef("config.Sub %q", key)
	sub := c.Viper.Sub(key)
	if sub == nil {
		return nil
	}
	return newConf(sub)
}

func (c *config) All() map[string]interface{} {
	return c.Viper.AllSettings()
}

func (c *config) Unmarshal(cfg interface{}) error {
	logger.Tracef("config.Unmarshal")
	return c.Viper.Unmarshal(cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "json" })
}

func (c *config) UnmarshalKey(key string, cfg interface{}) error {
	logger.Tracef("config.UnmarshalKey %q", key)
	return c.Viper.UnmarshalKey(key, cfg, func(c *mapstructure.DecoderConfig) { c.TagName = "json" })
}
func (c *config) Dump() {
	fmt.Println(" ===  CONFIG DUMP  ===")
	pretty.Printf(" VIPER: %# v\n", c.All())
	fmt.Println(" --- ")
}

func (c *config) Debug() {
	/*
		fmt.Printf("Aliases:\n%#v\n", v.aliases)
		fmt.Printf("Override:\n%#v\n", v.override)
		fmt.Printf("PFlags:\n%#v\n", v.pflags)
		fmt.Printf("Env:\n%#v\n", v.env)
		fmt.Printf("Key/Value Store:\n%#v\n", v.kvstore)
		fmt.Printf("Config:\n%#v\n", v.config)
		fmt.Printf("Defaults:\n%#v\n", v.defaults)
	*/
	fmt.Println(" ===  CONFIG DUMP  ===")
	//pp.Printf(" VIPER: %#v\n", c.Viper)
	c.Viper.Debug()
	fmt.Println(" ---  +  +  +  +  ---")
	pretty.Printf(" VIPER: %# v\n", c.All())
	fmt.Println(" ---  ---  ---")
}

var logger = logrus.NewLogger("config")

func init() {
	if debug := os.Getenv("DEBUG"); strings.Contains(debug, "config") {
		logger.SetLevel(logging.TraceLevel)
	}
}
