package logging

import "sync/atomic"

type Level int32

const (
	Quiet Level = iota
	Normal
	Verbose
)

var level atomic.Int32

func Set(l Level) {
	level.Store(int32(l))
}

func Get() Level {
	return Level(level.Load())
}

func Enabled(at Level) bool {
	return Get() >= at
}
