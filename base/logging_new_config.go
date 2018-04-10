package base

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/pkg/errors"
)

const (
	errorMinAge = 180
	warnMinAge  = 90
	infoMinAge  = 3
	debugMinAge = 1
)

type LoggingConfig struct {
	LogFilePath    string           `json:",omitempty"` // Absolute or relative path on the filesystem to the log file directory. A relative path is from the directory that contains the Sync Gateway executable file.
	RedactionLevel RedactionLevel   `json:",omitempty"`
	Console        LogConsoleConfig `json:",omitempty"` // Console Output
	Error          LogFileConfig    `json:",omitempty"` // Error log file output
	Warn           LogFileConfig    `json:",omitempty"` // Warn log file output
	Info           LogFileConfig    `json:",omitempty"` // Info log file output
	Debug          LogFileConfig    `json:",omitempty"` // Debug log file output

	DeprecatedDefaultLog *LogAppenderConfig `json:"default,omitempty"` // Deprecated "default" logging option.
}

type LogConsoleConfig struct {
	LogLevel     *LogLevel `json:",omitempty"` // Log Level for the console output
	LogKeys      []string  `json:",omitempty"` // Log Keys for the console output
	ColorEnabled bool      `json:",omitempty"` // Log with color for the console output

	logKey *LogKey
	output io.Writer
	logger *log.Logger
}

type LogFileConfig struct {
	Enabled  *bool             `json:",omitempty"` // Toggle for this log output
	Rotation logRotationConfig `json:",omitempty"` // Log rotation settings

	output io.Writer
	logger *log.Logger
}

type logRotationConfig struct {
	MaxSize   *int `json:",omitempty"` // The maximum size in MB of the log file before it gets rotated.
	MaxAge    *int `json:",omitempty"` // The maximum number of days to retain old log files.
	LocalTime bool `json:",omitempty"` // If true, it uses the computer's local time to format the backup timestamp.
}

func (c *LoggingConfig) Init() error {
	if c == nil {
		return errors.New("invalid LoggingConfig")
	}

	if err := validateLogFilePath(&c.LogFilePath); err != nil {
		return err
	}

	if err := c.Error.init(LevelError, c.LogFilePath, errorMinAge); err != nil {
		return err
	}
	errorLogger = c.Error

	if err := c.Warn.init(LevelWarn, c.LogFilePath, warnMinAge); err != nil {
		return err
	}
	warnLogger = c.Warn

	if err := c.Info.init(LevelInfo, c.LogFilePath, infoMinAge); err != nil {
		return err
	}
	infoLogger = c.Info

	if err := c.Debug.init(LevelDebug, c.LogFilePath, debugMinAge); err != nil {
		return err
	}
	debugLogger = c.Debug

	if err := c.Console.init(); err != nil {
		return err
	}
	consoleLogger = c.Console

	return nil
}

func (lcc *LogConsoleConfig) init() error {
	if lcc == nil {
		return errors.New("invalid LogConsoleConfig")
	}

	// Default to os.Stderr if alternative output is not set
	if lcc.output == nil {
		lcc.output = os.Stderr
	}
	lcc.logger = log.New(lcc.output, "", 0)

	// Default log level
	if lcc.LogLevel == nil {
		newLevel := LevelInfo
		lcc.LogLevel = &newLevel
	}

	// HTTP log key is always on
	newKeys := KeyHTTP
	if len(lcc.LogKeys) > 0 {
		newKeys.Enable(ToLogKey(lcc.LogKeys))
	}
	lcc.logKey = &newKeys

	return nil
}

func (lfc *LogFileConfig) init(level LogLevel, logFilePath string, minAge int) error {
	if lfc == nil {
		return errors.New("invalid LogFileConfig")
	}

	// set default enabled based on level
	if lfc.Enabled == nil {
		if level == LevelDebug {
			// Unset debug defaults to disabled
			lfc.Enabled = BoolPtr(false)
			lfc.logger = nil
			return nil
		}
		lfc.Enabled = BoolPtr(true)
	} else if !*lfc.Enabled {
		// remove logger if it's been disabled
		lfc.logger = nil
		return nil
	}

	if lfc.Rotation.MaxSize == nil {
		defaultMaxSize := 200
		lfc.Rotation.MaxSize = &defaultMaxSize
	}

	if lfc.Rotation.MaxAge == nil {
		// Default to double the minAge
		defaultMaxAge := minAge * 2
		lfc.Rotation.MaxAge = &defaultMaxAge
	} else if *lfc.Rotation.MaxAge < minAge {
		return fmt.Errorf("MaxAge for %s was set to %d which is below the minimum of %d", LogLevelName(level), *lfc.Rotation.MaxAge, minAge)
	}

	lfc.output = newLumberjackOutput(
		filepath.Join(filepath.FromSlash(logFilePath), "sg_"+LogLevelName(level)+".log"),
		*lfc.Rotation.MaxSize,
		*lfc.Rotation.MaxAge,
	)
	lfc.logger = log.New(lfc.output, "", 0)

	return nil
}

func validateLogFilePath(logFilePath *string) error {
	if logFilePath == nil || *logFilePath == "" {
		*logFilePath = defaultLogFilePath
	}

	err := os.MkdirAll(*logFilePath, 0700)
	if err != nil {
		return errors.Wrap(err, ErrInvalidLogFilePath.Error())
	}

	// Ensure LogFilePath is a directory. Lumberjack will check permissions when it opens the logfile.
	if f, err := os.Stat(*logFilePath); err != nil {
		return errors.Wrap(err, ErrInvalidLogFilePath.Error())
	} else if !f.IsDir() {
		return errors.Wrap(ErrInvalidLogFilePath, "not a directory")
	}

	return nil
}

func newLumberjackOutput(filename string, maxSize, maxAge int) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename: filename,
		MaxSize:  maxSize,
		MaxAge:   maxAge,
		Compress: true,
	}
}

// shouldLog returns true if the given logLevel and logKey should get logged.
func (lcc *LogConsoleConfig) shouldLog(logLevel LogLevel, logKey LogKey) bool {
	return lcc != nil &&
		lcc.logger != nil &&
		lcc.LogLevel.Enabled(logLevel) &&
		// if logging at KEY_ALL, allow it unless KEY_NONE is set
		((logKey == KeyAll && !lcc.logKey.Enabled(KeyNone)) ||
			// Otherwise check the given log key is enabled
			lcc.logKey.Enabled(logKey))
}

// shouldLog returns true if we can log.
func (lfc *LogFileConfig) shouldLog() bool {
	return lfc != nil && lfc.logger != nil &&
		// Check the log file is enabled
		(lfc.Enabled != nil && *lfc.Enabled)
}
