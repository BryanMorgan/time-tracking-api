package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/spf13/viper"
)

var Log *zap.Logger

// Ensure there is a default logger for unit tests
func init() {
	Log = zap.NewNop()
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func DevelopmentTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func InitLogger() {
	var err error
	var config zap.Config
	outputType := viper.GetString("logging.output")
	enableStacktrace := viper.GetBool("logging.enableStacktrace")
	switch outputType {

	case "none":
		Log = zap.NewNop()

	case "console":
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = DevelopmentTimeEncoder
		config.DisableStacktrace = !enableStacktrace
		config.DisableCaller = true
		Log, err = config.Build()

	default:
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.MessageKey = "@message"
		config.EncoderConfig.TimeKey = "@timestamp"
		config.EncoderConfig.LevelKey = "@level"
		config.DisableCaller = true
		config.Sampling = nil
		config.DisableStacktrace = !enableStacktrace
		hostname, _ := os.Hostname()
		config.InitialFields = map[string]interface{}{
			"host": hostname,
		}
		Log, err = config.Build()
	}

	if err != nil {
		log.Panicf("Failed to configure logging with setting: %s error: %v", outputType, err)
	}
}

func RequestLoggingHandler(next http.Handler) http.Handler {
	detailed := viper.GetBool("logging.detailed")

	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		fields := []zapcore.Field{}
		fields = append(fields, zap.String("method", r.Method))
		if detailed {
			fields = append(fields,
				zap.String("@type", "app"),
				zap.String("ip", r.RemoteAddr),
				zap.String("agent", r.UserAgent()))
		}

		// Wrap response writer to get access to the status code
		lw := NewLoggingResponseWriter(w)

		// After request
		defer func() {
			fields = append(fields,
				zap.Duration("duration", time.Since(start)),
				zap.Int("statusCode", lw.statusCode))
			Log.Info(r.RequestURI, fields...)
		}()

		next.ServeHTTP(lw, r)
	}

	return http.HandlerFunc(fn)
}

// Mirror the zap interface to avoid exposing the zap package outside of the logger

func NamedError(key string, err error) zapcore.Field {
	return zap.NamedError(key, err)
}

func Error(err error) zapcore.Field {
	return zap.Error(err)
}

func Skip() zapcore.Field {
	return zap.Skip()
}

func String(key string, val string) zapcore.Field {
	return zap.String(key, val)
}

func Stringer(key string, val fmt.Stringer) zapcore.Field {
	return zap.Stringer(key, val)
}

func ByteString(key string, val []byte) zapcore.Field {
	return zap.ByteString(key, val)
}

func Bool(key string, val bool) zapcore.Field {
	return zap.Bool(key, val)
}

func Int(key string, val int) zapcore.Field {
	return zap.Int(key, val)
}

func Duration(key string, val time.Duration) zapcore.Field {
	return zap.Duration(key, val)
}

func Time(key string, val time.Time) zapcore.Field {
	return zap.Time(key, val)
}
func Object(key string, val zapcore.ObjectMarshaler) zapcore.Field {
	return zap.Object(key, val)
}

func Any(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}
