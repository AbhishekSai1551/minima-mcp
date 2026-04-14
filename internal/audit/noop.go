package audit

type noopLogger struct{}

func NewNoopLogger() *Logger {
	return &Logger{}
}

func (l *Logger) noop() bool {
	return l.logDir == "" || l.disabled
}
