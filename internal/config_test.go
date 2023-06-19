package internal_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/kkereziev/notifier/internal"
)

const _testDotEnvFileName = ".env.dist"

func TestConfigValidation(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error("cannot get current working directory: ", err)
	}

	testEnvFilePath := fmt.Sprintf("%s/../../%s", dir, _testDotEnvFileName)
	if err := godotenv.Load(testEnvFilePath); err != nil {
		t.Error("failed loading env vars: ", err)
	}

	if _, err := internal.NewConfig(); err != nil {
		t.Error("error creating new config: ", err)
	}
}
