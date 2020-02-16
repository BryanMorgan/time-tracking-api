package logger

import (
	"bufio"
	"bytes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"regexp"
	"testing"
)

func captureOutput(funcToRun func()) string {
	var buffer bytes.Buffer

	oldLogger := Log

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buffer)

	Log = zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))

	funcToRun()
	writer.Flush()

	Log = oldLogger

	return buffer.String()
}

/**
 * Log a message and verify the captured output contains the message
 */
func TestLogging(t *testing.T) {
	InitLogger()

	const testLogMessage = "Test logger output"
	output := captureOutput(func() {
		Log.Info(testLogMessage)
	})

	var messageRegEx = regexp.MustCompile(`.*\sINFO\s(.*)\s`)
	subMatches := messageRegEx.FindStringSubmatch(output)

	if subMatches == nil || len(subMatches) < 2 {
		t.Errorf("Could not find message in logging output. Log output: %s. Test message: %s", output, testLogMessage)
	} else if subMatches[1] != testLogMessage {
		t.Errorf("Logging output does not contain test string. Expected: %s. Got: %v", testLogMessage, subMatches)
	}
}
