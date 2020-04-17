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
	DefaultConf Config = newConf(defaultViper)

	defaultViper = viper.GetViper()
)

func init() {
	viper.SupportedExts = append(viper.SupportedExts, "conf")
	defaultViper.SetConfigName("config")
	defaultViper.SetConfigType("yaml")
	defaultViper.SetDefault(DirFlag, DirDefault)
	defaultViper.BindEnv(DirFlag, "CONFIG_DIR")
	defaultViper.AutomaticEnv()
}

func Load() error                                     { return DefaultConf.Load() }
func LoadFrom(r io.Reader) error                      { return DefaultConf.LoadFrom(r) }
func MergeFrom(r io.Reader) error                     { return DefaultConf.MergeFrom(r) }
func MergeMap(m map[string]interface{}) error         { return DefaultConf.MergeMap(m) }
func WriteAs(filename string) error                   { return DefaultConf.WriteAs(filename) }
func Unmarshal(cfg interface{}) error                 { return DefaultConf.Unmarshal(cfg) }
func UnmarshalKey(key string, cfg interface{}) error  { return DefaultConf.UnmarshalKey(key, cfg) }
func BindEnv(key string, env string) error            { return DefaultConf.BindEnv(key, env) }
func BindFlag(key string, flag *Flag) error           { return DefaultConf.BindFlag(key, flag) }
func Sub(key string) Config                           { return DefaultConf.Sub(key) }
func All() map[string]interface{}                     { return DefaultConf.All() }
func Set(key string, val interface{})                 { DefaultConf.Set(key, val) }
func SetDefault(key string, val interface{})          { DefaultConf.SetDefault(key, val) }
func Get(key string) interface{}                      { return DefaultConf.Get(key) }
func GetString(key string) string                     { return DefaultConf.GetString(key) }
func GetBool(key string) bool                         { return DefaultConf.GetBool(key) }
func GetInt(key string) int                           { return DefaultConf.GetInt(key) }
func GetFloat64(key string) float64                   { return DefaultConf.GetFloat64(key) }
func GetDuration(key string) time.Duration            { return DefaultConf.GetDuration(key) }
func GetStringSlice(key string) []string              { return DefaultConf.GetStringSlice(key) }
func GetStringMap(key string) map[string]interface{}  { return DefaultConf.GetStringMap(key) }
func GetStringMapString(key string) map[string]string { return DefaultConf.GetStringMapString(key) }
