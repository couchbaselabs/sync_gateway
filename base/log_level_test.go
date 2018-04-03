package base

import (
	"testing"
	"time"

	assert "github.com/couchbaselabs/go.assert"
)

func TestLogLevel(t *testing.T) {
	logLevel := LEVEL_NONE
	assert.False(t, logLevel.Enabled(LEVEL_DEBUG))
	assert.False(t, logLevel.Enabled(LEVEL_INFO))
	assert.False(t, logLevel.Enabled(LEVEL_WARN))
	assert.False(t, logLevel.Enabled(LEVEL_ERROR))

	logLevel.Set(LEVEL_INFO)
	assert.False(t, logLevel.Enabled(LEVEL_DEBUG))
	assert.True(t, logLevel.Enabled(LEVEL_INFO))
	assert.True(t, logLevel.Enabled(LEVEL_WARN))
	assert.True(t, logLevel.Enabled(LEVEL_ERROR))

	logLevel.Set(LEVEL_WARN)
	assert.False(t, logLevel.Enabled(LEVEL_DEBUG))
	assert.False(t, logLevel.Enabled(LEVEL_INFO))
	assert.True(t, logLevel.Enabled(LEVEL_WARN))
	assert.True(t, logLevel.Enabled(LEVEL_ERROR))
}

func TestLogLevelNames(t *testing.T) {
	assert.Equals(t, LogLevelName(LEVEL_NONE), "none")
	assert.Equals(t, LogLevelName(LEVEL_DEBUG), "debug")
	assert.Equals(t, LogLevelName(LEVEL_INFO), "info")
	assert.Equals(t, LogLevelName(LEVEL_WARN), "warn")
	assert.Equals(t, LogLevelName(LEVEL_ERROR), "error")
}

func TestLogLevelText(t *testing.T) {
	var logLevel LogLevel
	text, err := logLevel.MarshalText()
	assert.Equals(t, err, nil)
	assert.Equals(t, string(text), "none")

	logLevel.Set(LEVEL_DEBUG)
	text, err = logLevel.MarshalText()
	assert.Equals(t, err, nil)
	assert.Equals(t, string(text), "debug")

	logLevel.Set(LEVEL_INFO)
	text, err = logLevel.MarshalText()
	assert.Equals(t, err, nil)
	assert.Equals(t, string(text), "info")

	logLevel.Set(LEVEL_WARN)
	text, err = logLevel.MarshalText()
	assert.Equals(t, err, nil)
	assert.Equals(t, string(text), "warn")

	logLevel.Set(LEVEL_ERROR)
	text, err = logLevel.MarshalText()
	assert.Equals(t, err, nil)
	assert.Equals(t, string(text), "error")
}

func TestLogLevelConcurrency(t *testing.T) {
	logLevel := LEVEL_WARN
	stop := make(chan struct{})

	go func() {
		for {
			select {
			default:
				logLevel.Set(LEVEL_ERROR)
			case <-stop:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			default:
				logLevel.Set(LEVEL_DEBUG)
			case <-stop:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			default:
				_ = logLevel.Enabled(LEVEL_WARN)
			case <-stop:
				return
			}
		}
	}()

	time.Sleep(time.Millisecond * 100)
	stop <- struct{}{}
}

func BenchmarkLogLevelName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = LogLevelName(LEVEL_WARN)
	}
}

func BenchmarkLogLevelEnabled(b *testing.B) {
	logLevel := LEVEL_INFO
	benchmarkLogLevelEnabled(b, "Hit", LEVEL_ERROR, logLevel)
	benchmarkLogLevelEnabled(b, "Miss", LEVEL_DEBUG, logLevel)
}

func benchmarkLogLevelEnabled(b *testing.B, name string, l LogLevel, logLevel LogLevel) {
	b.Run(name, func(bn *testing.B) {
		for i := 0; i < bn.N; i++ {
			logLevel.Enabled(l)
		}
	})
}
