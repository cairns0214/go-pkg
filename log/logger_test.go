package log

import (
	"testing"
)

func TestNew(t *testing.T) {
	config := Configuration{
		EnableConsole:     true,
		ConsoleLevel:      DEBUG,
		ConsoleJSONFormat: false,
		EnableFile:        true,
		FileLevel:         INFO,
		FileJSONFormat:    true,
		FileLocation:      `../tmp/logger_test.log`,
	}
	New(config, InstanceZapLogger)

	Debugf("debug...")
	Infof("info...")
	Warnf("warn...")
	Errorf("error...")

	contextFieldsLogger := WithFields(Fields{"key1": "value1"})
	contextFieldsLogger.Warnf("content warn...")
	contextFieldsLogger.Fatalf("content fatal...")
}
