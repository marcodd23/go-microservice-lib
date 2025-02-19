package configx_test

import (
	"os"
	"testing"

	"github.com/marcodd23/go-micro-core/pkg/configx"
	"github.com/stretchr/testify/assert"
)

// Shared configuration content
var configContent = `
name: "TestApp"
environment: "development"
version: "latest"
logging:
  level: "debug"
gcp:
  projectNumber: 620222630834
  project: test-project
  location: europe-west4
image:
  repository: europe-west4-docker.pkg.dev/ingka-opoa-nexus-dev/nexus-docker-repo
  name: test-app
  pullPolicy: Always
  tag: latest
server:
  port: "8080"
  concurrency: 10
  disableStartupMsg: false
  bodyLimit: 1048576
`

// Shared configuration content
var configContentNoLimit = `
name: "TestApp"
environment: "development"
version: "latest"
logging:
  level: "debug"
gcp:
  projectNumber: 620222630834
  project: test-project
  location: europe-west4
image:
  repository: europe-west4-docker.pkg.dev/ingka-opoa-nexus-dev/nexus-docker-repo
  name: test-app
  pullPolicy: Always
  tag: latest
server:
  port: "8080"
  concurrency: 10
  disableStartupMsg: false
`

type TestConfiguration struct {
	configx.BaseConfig `mapstructure:",squash"`
}

func createTestConfigFile(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}

	return file.Name()
}

func TestLoadConfigFromFile(t *testing.T) {
	configFilePath := createTestConfigFile(t, configContent)
	defer os.Remove(configFilePath)

	var cfg TestConfiguration
	err := configx.ReadConfiguration(configFilePath, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "TestApp", cfg.GetServiceName())
	assert.Equal(t, "development", cfg.GetEnvironment())
	assert.NotNil(t, cfg.Logging)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.NotNil(t, cfg.Server)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, 10, cfg.Server.Concurrency)
	assert.NotNil(t, cfg.Gcp)
	assert.Equal(t, "test-project", cfg.Gcp.ProjectId)
	assert.Equal(t, "620222630834", cfg.Gcp.ProjectNumber)
	assert.Equal(t, "europe-west4", cfg.Gcp.Location)
	assert.Equal(t, false, cfg.Server.DisableStartupMessage)
}

func TestEnvVariableOverridesConfig(t *testing.T) {
	configFilePath := createTestConfigFile(t, configContent)
	defer os.Remove(configFilePath)

	// Set environment variable to override server port
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	var cfg TestConfiguration
	err := configx.ReadConfiguration(configFilePath, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "TestApp", cfg.GetServiceName())
	assert.Equal(t, "development", cfg.GetEnvironment())
	assert.NotNil(t, cfg.Logging)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.NotNil(t, cfg.Server)
	assert.Equal(t, "9090", cfg.Server.Port) // Expecting overridden value
	assert.Equal(t, 10, cfg.Server.Concurrency)
	assert.Equal(t, false, cfg.Server.DisableStartupMessage)
	assert.Equal(t, 1048576, cfg.Server.GetBodyLimit())
}

func TestMissingBodySizeConfig(t *testing.T) {
	configFilePath := createTestConfigFile(t, configContentNoLimit)
	defer os.Remove(configFilePath)

	var cfg TestConfiguration
	err := configx.ReadConfiguration(configFilePath, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "TestApp", cfg.GetServiceName())
	assert.Equal(t, "development", cfg.GetEnvironment())
	assert.NotNil(t, cfg.Logging)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.NotNil(t, cfg.Server)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, 10, cfg.Server.Concurrency)
	assert.NotNil(t, cfg.Gcp)
	assert.Equal(t, "test-project", cfg.Gcp.ProjectId)
	assert.Equal(t, "620222630834", cfg.Gcp.ProjectNumber)
	assert.Equal(t, "europe-west4", cfg.Gcp.Location)
	assert.Equal(t, false, cfg.Server.DisableStartupMessage)
	assert.Equal(t, 4194304, cfg.Server.GetBodyLimit()) // Expecting default value
}
