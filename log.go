package browser

import "github.com/sohaha/zlsgo/zlog"

var Log = zlog.New("")

func init() {
	Log.ResetFlags(zlog.BitLevel | zlog.BitTime)
}

type Logger struct {
	log *zlog.Logger
}

func newLogger() *Logger {
	return &Logger{log: Log}
}

func (l *Logger) Println(i ...interface{}) {
	l.log.Tips(i...)
}
