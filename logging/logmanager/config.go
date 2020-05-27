package logmanager

// Config is a binding that supports to define default log levels for multiple loggers
type Config struct {
	DefaultLevel string                `json:"default-level"`
	Loggers      []LoggerConfig        `json:"loggers"`
	Hooks        map[string]HookConfig `json:"hooks"`
}

// LoggerConfig is configuration of a particular logger.
// Currently we support only logger level.
type LoggerConfig struct {
	Name  string
	Level string // levels: debug, info, warn, error, fatal, panic
}

// HookConfig contains configuration of hook services
type HookConfig struct {
	Protocol string
	Address  string
	Port     int
	Levels   []string
}

// DefaultConfig creates default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultLevel: "info",
		Loggers:      []LoggerConfig{},
		Hooks:        make(map[string]HookConfig),
	}
}
