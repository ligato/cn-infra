package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	flag "github.com/spf13/pflag"
	"go.ligato.io/cn-infra/v2/logging"
)

func init() {
	flag.String("config-dir", ".", "Location of config directory")
}

const (
	// FlagSuffix is added to plugin name while loading plugins configuration.
	FlagSuffix = "-config"

	// EnvSuffix is added to plugin name while loading plugins configuration from ENV variable.
	EnvSuffix = "_CONFIG"

	// FileExtension is used as a default extension for config files in flags.
	FileExtension = ".conf"
)

const pluginConfigsPrefix = "plugin-configs."

// FlagName returns config flag name for the name, usually plugin.
func FlagName(name string) string {
	return strings.ToLower(name) + FlagSuffix
}

// Filename returns config filename for the name, usually plugin.
func Filename(name string) string {
	return name + FileExtension
}

// EnvVar returns config env variable for the name, usually plugin.
func EnvVar(name string) string {
	return strings.ReplaceAll(strings.ToUpper(name), "-", "_") + EnvSuffix
}

const (
	// DirFlag as flag name (see implementation in declareFlags())
	// is used to define default directory where config files reside.
	// This flag name is derived from the name of the plugin.
	DirFlag = "config-dir"

	// DirDefault holds a default value "." for flag, which represents current working directory.
	DirDefault = "."

	// DirUsage used as a flag (see implementation in declareFlags()).
	DirUsage = "Location of the config files; can also be set via 'CONFIG_DIR' env variable."
)

// DefineDirFlag defines flag for configuration directory.
/*func DefineDirFlag() {
	if flag.CommandLine.Lookup(DirFlag) == nil {
		flag.CommandLine.String(DirFlag, DirDefault, DirUsage)
	}
}*/

// PluginConfig is API for plugins to access configuration.
//
// Aim of this API is to let a particular plugin to bind it's configuration
// without knowing a particular key name. The key name is injected into Plugin.
type PluginConfig interface {
	// LoadValue parses configuration for a plugin and stores the results in data.
	// The argument data is a pointer to an instance of a go structure.
	LoadValue(data interface{}) (found bool, err error)

	// GetConfigName returns config name derived from plugin name:
	// flag = PluginName + FlagSuffix (evaluated most often as absolute path to a config file)
	GetConfigName() string
}

type (
	FlagSet = flag.FlagSet
	Flag    = flag.Flag
)

// pluginFlags is used for storing flags for Plugins before agent starts.
var pluginFlags = make(map[string]*FlagSet)

func GetFlagSetFor(name string) *FlagSet {
	return pluginFlags[name]
}

// DefineFlagsFor registers defined flags for plugin with given name.
/*func DefineFlagsFor(name string) {
	if plugSet, ok := pluginFlags[name]; ok {
		flag.CommandLine.AddFlagSet(plugSet)
		//plugSet.VisitAll(func(f *flag.Flag) {
		//	flag.CommandLine.Var(f.Value, f.Name, f.Usage)
		//})
	}
}*/

type options struct {
	Conf Config

	FlagName    string
	FlagDefault string
	FlagUsage   string

	flagSet *FlagSet
}

// Option is an option used in ForPlugin
type Option func(*options)

// WithCustomizedFlag is an option to customize config flag for plugin in ForPlugin.
// The parameters are used to replace defaults in this order: flag name, default, usage.
func WithCustomizedFlag(s ...string) Option {
	return func(o *options) {
		if len(s) > 0 {
			o.FlagName = s[0]
		}
		if len(s) > 1 {
			o.FlagDefault = s[1]
		}
		if len(s) > 2 {
			o.FlagUsage = s[2]
		}
	}
}

// WithExtraFlags is an option to define additional flags for plugin in ForPlugin.
func WithExtraFlags(f func(flags *FlagSet)) Option {
	return func(o *options) {
		f(o.flagSet)
	}
}

// WithConfig is an option to set custom Config in ForPlugin.
func WithConfig(conf Config) Option {
	return func(o *options) {
		o.Conf = conf
	}
}

// ForPlugin returns API that is injectable to a particular Plugin
// and is used to read it's configuration.
//
// By default it tries to lookup `<plugin-name> + "-config"`in flags and declare
// the flag if it's not defined yet. There are options that can be used
// to customize the config flag for plugin and/or define additional flags for the plugin.
func ForPlugin(name string, opts ...Option) PluginConfig {
	opt := options{
		FlagName:    FlagName(name),
		FlagDefault: Filename(name),
		FlagUsage:   fmt.Sprintf("Location of the %q plugin config file; can set also via %q.", name, EnvVar(name)),
		flagSet:     flag.NewFlagSet(name, flag.ContinueOnError),
	}
	for _, o := range opts {
		o(&opt)
	}

	if opt.Conf == nil {
		opt.Conf = DefaultConf
	}

	if opt.FlagName != "" && opt.flagSet.Lookup(opt.FlagName) == nil {
		opt.flagSet.String(opt.FlagName, opt.FlagDefault, opt.FlagUsage)

		f := opt.flagSet.Lookup(opt.FlagName)
		f.Deprecated = "use single config file" //"Plugin flags XXX-config have been deprecated!"
		//f.Hidden = true
		opt.Conf.BindEnv(pluginConfigsPrefix+opt.FlagName, EnvVar(name))
		opt.Conf.BindFlag(pluginConfigsPrefix+opt.FlagName, f)
	}

	opt.flagSet.VisitAll(func(f *flag.Flag) {
		if f.Annotations == nil {
			f.Annotations = make(map[string][]string)
		}
		f.Annotations["plugin"] = []string{name}
	})

	pluginFlags[name] = opt.flagSet

	return &pluginConfig{
		name:       name,
		conf:       opt.Conf,
		configFlag: opt.FlagName,
	}
}

// Dir returns config directory by evaluating the flag DirFlag. It interprets "." as current working directory.
func Dir() (dir string, err error) {
	return GetString("config-dir"), nil
	/*if flg := flag.CommandLine.Lookup(DirFlag); flg != nil {
		val := flg.Value.String()
		logrus.DefaultLogger().Debugf("dir flag value: %q", val)
		if strings.HasPrefix(val, ".") {
			cwd, err := os.Getwd()
			if err != nil {
				logrus.DefaultLogger().Errorf("getcwd: %v", err)
				return cwd, err
			}
			return filepath.Join(cwd, val), nil
		}
		return val, nil
	}
	return "", nil*/
}

type pluginConfig struct {
	name string
	conf Config

	configFlag string
	access     sync.Mutex
	configName string
}

// LoadValue binds the configuration to config method argument.
func (p *pluginConfig) LoadValue(cfg interface{}) (found bool, err error) {
	log := logger.WithField("name", p.name)

	conf := p.conf

	if logger.GetLevel() >= logging.TraceLevel {
		conf.(*Conf).Debug()
	}

	//var def map[string]interface{}
	def := cfgToStringMap(cfg)
	log.Tracef("setting default to: %+v", def)
	conf.SetDefault(p.name, def)

	cfgName := p.GetConfigName()
	/*if cfgName == "" {
		return false, nil
	}*/
	if cfgName != "" {
		log.Tracef("pluginConfig.cfgName: %q", cfgName)
		data, err := parseYamlFileForMerge(cfgName)
		if err != nil {
			return false, err
		}
		log.Tracef("data from plugin config file: %+v", data)
		err = conf.MergeMap(map[string]interface{}{
			p.name: data,
		})
		if err != nil {
			log.Debugf("ERROR merging from plugin config file: %v", err)
			return false, err
		}
		log.Debugf("merged %q OK: %+v", p.name, conf.(*Conf).AllKeys())
		if logger.GetLevel() >= logging.TraceLevel {
			conf.(*Conf).Debug()
		}
	}

	pluginCfg := conf.Get(p.name)
	if pluginCfg == nil || !conf.(*Conf).InConfig(p.name) {
		return false, nil
	}
	log.Tracef("pluginCfg: %+v", pluginCfg)

	/*if conf.GetBool(p.name + ".disabled") {
		return false, nil
	}*/

	if err := conf.UnmarshalKey(p.name, cfg); err != nil {
		return true, err
	}

	return true, nil
}

func cfgToStringMap(cfg interface{}) map[string]interface{} {
	b, err := json.Marshal(cfg)
	if err != nil {
		return nil
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

// GetConfigName looks up flag value and uses it to:
// 1. Find config in flag value location.
// 2. Alternatively, it tries to find it in config dir
// (see also Dir() comments).
func (p *pluginConfig) GetConfigName() string {
	//return p.conf.GetString(p.configFlag)
	p.access.Lock()
	defer p.access.Unlock()
	if p.configName == "" {
		p.configName = p.getConfigName()
	}
	logger.Debugf("GetConfigName %v: %q", p.configFlag, p.configName)
	return p.configName
}

func (p *pluginConfig) getConfigName() string {
	//if flg := flag.CommandLine.Lookup(p.configFlag); flg != nil {
	//logger.Tracef("found config flag %v: %v", flg.Name, flg.Value)
	//if val := flg.Value.String(); val != "" {
	if val := p.conf.GetString(pluginConfigsPrefix + p.configFlag); val != "" {
		logger.Tracef("plugin config file value: %q", val)
		// if the file exists (value from flag)
		if _, err := os.Stat(val); !os.IsNotExist(err) {
			return val
		}
		cfgDir, err := Dir()
		if err != nil {
			logger.Error(err)
			return ""
		}
		// if the file exists (flag value in config dir)
		dirVal := filepath.Join(cfgDir, val)
		logger.Tracef("checking dirVal: %q", dirVal)
		if _, err := os.Stat(dirVal); !os.IsNotExist(err) {
			return dirVal
		} else {
			logger.Debugf("ERROR os.Stat %q: %v", dirVal, err)
		}
	}
	//}
	return ""
}
