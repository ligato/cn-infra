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
	"io"
	"time"

	"github.com/spf13/viper"
)

var (
	DefaultConfig Config = defaultConf

	defaultConf  = newConf(defaultViper)
	defaultViper = viper.GetViper()
)

func init() {
	viper.SupportedExts = append(viper.SupportedExts, "conf")
	defaultViper.SetConfigName("config")
	defaultViper.SetConfigType("yaml")
	defaultViper.SetDefault("config-dir", ".")
	defaultViper.BindEnv("config-dir", "CONFIG_DIR")
	defaultViper.AutomaticEnv()
}

func Read() error                                     { return DefaultConfig.Read() }
func ReadFrom(r io.Reader) error                      { return DefaultConfig.ReadFrom(r) }
func MergeFrom(r io.Reader) error                     { return DefaultConfig.MergeFrom(r) }
func MergeMap(m map[string]interface{}) error         { return DefaultConfig.MergeMap(m) }
func WriteAs(filename string) error                   { return DefaultConfig.WriteAs(filename) }
func Unmarshal(cfg interface{}) error                 { return DefaultConfig.Unmarshal(cfg) }
func UnmarshalKey(key string, cfg interface{}) error  { return DefaultConfig.UnmarshalKey(key, cfg) }
func Set(key string, val interface{})                 { DefaultConfig.Set(key, val) }
func BindEnv(key string, env string) error            { return DefaultConfig.BindEnv(key, env) }
func BindFlag(key string, flag *Flag) error           { return DefaultConfig.BindFlag(key, flag) }
func SetDefault(key string, val interface{})          { DefaultConfig.SetDefault(key, val) }
func Sub(key string) Config                           { return DefaultConfig.Sub(key) }
func All() map[string]interface{}                     { return DefaultConfig.All() }
func Get(key string) interface{}                      { return DefaultConfig.Get(key) }
func GetString(key string) string                     { return DefaultConfig.GetString(key) }
func GetBool(key string) bool                         { return DefaultConfig.GetBool(key) }
func GetInt(key string) int                           { return DefaultConfig.GetInt(key) }
func GetFloat64(key string) float64                   { return DefaultConfig.GetFloat64(key) }
func GetDuration(key string) time.Duration            { return DefaultConfig.GetDuration(key) }
func GetStringSlice(key string) []string              { return DefaultConfig.GetStringSlice(key) }
func GetStringMap(key string) map[string]interface{}  { return DefaultConfig.GetStringMap(key) }
func GetStringMapString(key string) map[string]string { return DefaultConfig.GetStringMapString(key) }
