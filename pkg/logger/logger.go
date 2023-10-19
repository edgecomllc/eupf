package logger

type Logger interface {
	SetLoggerLevel(level string) error
	Fatal(msg string)
	Fatalf(format string, v ...interface{})
	Fatale(err error)
	Error(err error)
	Info(msg string)
	Infof(format string, v ...interface{})
	Panicf(format string, v ...interface{})
	Printf(format string, v ...interface{})
}

var impl Logger

func SetLogger(repository Logger) {
	impl = repository
}

func SetLoggerLevel(level string) error {
	return impl.SetLoggerLevel(level)
}

func Fatal(msg string) {
	impl.Fatal(msg)
}

func Fatalf(format string, v ...interface{}) {
	impl.Fatalf(format, v...)
}

func Fatale(err error) {
	impl.Fatale(err)
}

func Error(err error) {
	impl.Error(err)
}

func Info(msg string) {
	impl.Info(msg)
}

func Infof(format string, v ...interface{}) {
	impl.Infof(format, v...)
}

func Panicf(format string, v ...interface{}) {
	impl.Panicf(format, v...)
}

func Printf(format string, v ...interface{}) {
	impl.Printf(format, v...)
}
