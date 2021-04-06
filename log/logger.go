package log

import "errors"

var _logger Logger

// Level represents a log level.
type Level int32

const (
	// DEBUG information for programmer lowlevel analysis.
	DEBUG Level = 1
	// INFO information about steady state operations.
	INFO Level = 2
	// WARN is for logging messages about possible issues.
	WARN Level = 3
	// ERROR is for logging errors.
	ERROR Level = 4
	// FATAL is for logging fatal messages. The sytem shutsdown after logging the message.
	FATAL Level = 5

	PANIC Level = 6
)

const (
	// InstanceZapLogger is zap logger instance.
	InstanceZapLogger int = iota
	// InstanceLogrusLogger is logrus logger instance.
	InstanceLogrusLogger
)

var (
	errInvalidLoggerInstance = errors.New("invalid logger instance")
)

// Fields Type to pass when we want to call WithFields for structured logging.
type Fields map[string]interface{}

// Logger ...
type Logger interface {
	Debugf(format string, args ...interface{})

	Infof(format string, args ...interface{})

	Warnf(format string, args ...interface{})

	Errorf(format string, args ...interface{})

	Fatalf(format string, args ...interface{})

	Panicf(format string, args ...interface{})

	WithFields(fields Fields) Logger
}

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "debug"
	case INFO:
		return "info"
	case WARN:
		return "warn"
	case ERROR:
		return "error"
	case FATAL:
		return "fatal"
	case PANIC:
		return "panic"
	default:
		return "unknown"
	}
}

func Str2Level(level string) (logLevel Level) {
	switch level {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warning", "warn":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	case "fatal":
		logLevel = FATAL
	case "panic":
		logLevel = PANIC
	default:
		logLevel = INFO
	}
	return
}

// Configuration stores the config for the logger
// For some loggers there can only be one level across writers, for such the level of Console is picked by default.
type Configuration struct {
	EnableConsole     bool   `json:"enable_console" mapstructure:"enable_console"`
	ConsoleJSONFormat bool   `json:"console_json_format" mapstructure:"console_json_format"`
	ConsoleLevel      Level  `json:"console_level" mapstructure:"console_level"`
	EnableFile        bool   `json:"enable_file" mapstructure:"enable_file"`
	FileJSONFormat    bool   `json:"file_json_format" mapstructure:"file_json_format"`
	FileLevel         Level  `json:"file_level" mapstructure:"file_level"`
	FileLocation      string `json:"file_location" mapstructure:"file_location"`
	FileMaxSize       int    `json:"file_max_size" mapstructure:"file_max_size"`
	FileMaxBackups    int    `json:"file_max_backups" mapstructure:"file_max_backups"`
	FileMaxAge        int    `json:"file_max_age" mapstructure:"file_max_age"`
	FileCompress      bool   `json:"file_compress" mapstructure:"file_compress"`
}

// New returns an instance of logger
func New(config Configuration, loggerInstance int) {
	switch loggerInstance {
	case InstanceZapLogger:
		logger, err := newZapLogger(config)
		panicError(err)
		_logger = logger

	case InstanceLogrusLogger:
		logger, err := newLogrusLogger(config)
		panicError(err)
		_logger = logger

	default:
		panic(errInvalidLoggerInstance)
	}
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	_logger.Debugf(format, args...)
}

// Infof ...
func Infof(format string, args ...interface{}) {
	_logger.Infof(format, args...)
}

// Warnf ...
func Warnf(format string, args ...interface{}) {
	_logger.Warnf(format, args...)
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	_logger.Errorf(format, args...)
}

// Fatalf ...
func Fatalf(format string, args ...interface{}) {
	_logger.Fatalf(format, args...)
}

// Panicf ...
func Panicf(format string, args ...interface{}) {
	_logger.Panicf(format, args...)
}

// WithFields ...
func WithFields(fields Fields) Logger {
	return _logger.WithFields(fields)
}
