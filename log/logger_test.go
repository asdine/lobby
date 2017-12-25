package log_test

import (
	"bytes"
	"io/ioutil"
	stdlog "log"
	"testing"

	"github.com/asdine/lobby/log"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	var testCases = []func(l *log.Logger){
		func(l *log.Logger) { l.Println("message") },
		func(l *log.Logger) { l.Printf("message\n") },
	}

	for _, test := range testCases {
		t.Run("WithPrefix", func(t *testing.T) {
			var buff bytes.Buffer
			stdlog.SetFlags(0)
			logger := log.New(log.Output(&buff), log.Prefix("prefix"))
			test(logger)
			require.Equal(t, "i | prefix message\n", buff.String())
		})

		t.Run("WithoutPrefix", func(t *testing.T) {
			var buff bytes.Buffer
			stdlog.SetFlags(0)
			logger := log.New(log.Output(&buff))
			test(logger)
			require.Equal(t, "i | message\n", buff.String())
		})
	}
}

func TestLoggerDebug(t *testing.T) {
	var testCases = []func(l *log.Logger){
		func(l *log.Logger) { l.Debug("message") },
		func(l *log.Logger) { l.Debugf("message\n") },
	}

	for _, test := range testCases {
		t.Run("WithoutDebug", func(t *testing.T) {
			var buff bytes.Buffer
			stdlog.SetFlags(0)
			logger := log.New(log.Output(&buff), log.Prefix("prefix"))
			test(logger)
			require.Equal(t, "", buff.String())
		})

		t.Run("WithDebugAndPrefix", func(t *testing.T) {
			var buff bytes.Buffer
			stdlog.SetFlags(0)
			logger := log.New(
				log.Output(&buff),
				log.Prefix("prefix"),
				log.Debug(true),
			)
			test(logger)
			require.Equal(t, "d | prefix message\n", buff.String())
		})

		t.Run("WithDebugNoPrefix", func(t *testing.T) {
			var buff bytes.Buffer
			stdlog.SetFlags(0)
			logger := log.New(
				log.Output(&buff),
				log.Debug(true),
			)
			test(logger)
			require.Equal(t, "d | message\n", buff.String())
		})
	}
}

func BenchmarkLog(b *testing.B) {
	logger := log.New(log.Output(ioutil.Discard), log.Prefix("prefix"))
	vs := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		vs[i] = "foo"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Println(vs...)
	}
}
