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
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/logging/logrus"
)

// Config defines API for config management.
type Config interface {
	Load() error
	LoadFrom(io.Reader) error
	MergeFrom(io.Reader) error
	MergeMap(map[string]interface{}) error
	WriteAs(filename string) error

	BindEnv(key string, env string) error
	BindFlag(key string, flag *Flag) error
	SetDefault(key string, val interface{})
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
	v := viper.NewWithOptions()
	return newConf(v)
}

var logger = logrus.NewLogger("config")

func init() {
	if debug := os.Getenv("DEBUG"); strings.Contains(debug, "config") {
		logger.SetLevel(logging.TraceLevel)
	}
}

type Conf struct {
	*viper.Viper
}

func newConf(v *viper.Viper) *Conf {
	if v.Get(DirFlag) == nil {
		v.SetDefault(DirFlag, ".")
		v.BindEnv(DirFlag, "CONFIG_DIR")
	}
	return &Conf{Viper: v}
}

func (c *Conf) Load() error {
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
		c.Viper.Debug()
	}

	return nil
}

func (c *Conf) LoadFrom(r io.Reader) error {
	return c.Viper.ReadConfig(r)
}

func (c *Conf) MergeFrom(r io.Reader) error {
	return c.Viper.MergeConfig(r)
}

func (c *Conf) MergeMap(m map[string]interface{}) error {
	return c.Viper.MergeConfigMap(m)
}

func (c *Conf) WriteAs(filename string) error {
	logger.Tracef("Conf.WriteAs %q", filename)
	return c.Viper.WriteConfigAs(filename)
}

func (c *Conf) BindEnv(key string, env string) error {
	logger.Tracef("bind config key %q to ENV %q", key, env)
	err := c.Viper.BindEnv(key, env)
	if err != nil {
		logger.Debugf("ERROR binding to env: %v", err)
	}
	return err
}

func (c *Conf) BindFlag(key string, flag *Flag) error {
	logger.Tracef("bind config key %q to FLAG %q", key, flag.Name)
	err := c.Viper.BindPFlag(key, flag)
	if err != nil {
		logger.Debugf("ERROR binding to flag: %v", err)
	}
	return err
}

func (c *Conf) Sub(key string) Config {
	logger.Tracef("Conf.Sub %q", key)
	sub := c.Viper.Sub(key)
	if sub == nil {
		return nil
	}
	return newConf(sub)
}

func (c *Conf) All() map[string]interface{} {
	logger.Debugf("Conf.All: %+v", c.All())
	return c.Viper.AllSettings()
}

func (c *Conf) Unmarshal(cfg interface{}) error {
	logger.Tracef("Conf.Unmarshal")
	return c.Viper.Unmarshal(cfg)
}

func (c *Conf) UnmarshalKey(key string, cfg interface{}) error {
	logger.Tracef("Conf.UnmarshalKey %q", key)
	return c.Viper.UnmarshalKey(key, cfg)
}
