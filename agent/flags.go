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

package agent

import (
	flag "github.com/spf13/pflag"

	"go.ligato.io/cn-infra/v2/config"
)

func init() {
	config.SetDefault("log-level", "info")
	config.BindEnv("log-level", "LOG_LEVEL")
	config.SetDefault("config", "")
	config.BindEnv("config", "CONFIG_FILE")
}

func AddFlagsTo(set *flag.FlagSet) {
	set.StringP("config", "", "", "Config file location.")
	set.StringP("log-level", "", "", "Set the logging level (debug|info|warn|error|fatal).")
	set.BoolP("debug", "D", false, "Enable debug mode.")
	set.BoolP("version", "V", false, "Print version and exit.")
}

/*
var (
	FlagSet = flag.NewFlagSet("agent", flag.ExitOnError)

	logLevelVal string
	configVal   string
	versionVal  bool
	debugVal    bool
)

func init() {
	FlagSet.StringVarP(&logLevelVal, "log-level", "", "info", "Set the logging level (debug|info|warn|error|fatal).")
	FlagSet.StringVarP(&configVal, "config", "", "", "Config file location.")
	FlagSet.BoolVarP(&versionVal, "version", "V", false, "Print version and exit.")
	FlagSet.BoolVarP(&debugVal, "debug", "D", false, "Enable debug mode.")
}*/
