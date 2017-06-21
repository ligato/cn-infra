package logging

// LogLevel represents severity of log record
type LogLevel int

const (
	// DebugLevel is the lowest log level
	DebugLevel LogLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

// Logger provides logging capabilities
type Logger interface {
	LogWithLevel
	// SetLevel modifies the LogLevel
	SetLevel(level LogLevel)
	// GetLevel returns currently set logLevel
	GetLevel() LogLevel
	// WithField creates one structured field
	WithField(key string, value interface{}) LogWithLevel
	// WithFields creates multiple structured fields
	WithFields(fields map[string]interface{}) LogWithLevel
}

// LogWithLevel allows to log with different log levels
type LogWithLevel interface {
	// Debug logs using Debug level
	Debug(args ...interface{})
	// Info logs using Info level
	Info(args ...interface{})
	// Warning logs using Warning level
	Warning(args ...interface{})
	// Error logs using Error level
	Error(args ...interface{})
	// Panic logs using Panic level and panics
	Panic(args ...interface{})
	// Fatal logs using Fatal level and calls os.Exit(1)
	Fatal(args ...interface{})
}

// Registry defines a set of public function for interaction with the logger Registry
type Registry interface {
	// List Loggers returns a map (loggerName => log level)
	ListLoggers() map[string]string
	// SetLevel modifies log level of selected logger in the registry
	SetLevel(logger, level string) error
	// GetLevel returns the currently set log level of the logger from registry
	GetLevel(logger string) (string, error)
	// GetLoggerByName returns a logger instance identified by name from registry
	GetLoggerByName(name string) (*Logger, bool)
	// ClearRegistry removes all loggers except the default one from registry
	ClearRegistry()
}
