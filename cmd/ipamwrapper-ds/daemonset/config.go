// Copyright 2022 Authors of Inspur
// SPDX-License-Identifier: Apache-2.0

package daemonset

import (
	"github.com/Inspur-Data/ipamwrapper/pkg/logging"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const (
	HttpPort       = "5555"
	UnixSocketPath = "/var/run/inspur/ipamwrapper.sock"
	ConfigPath     = "/tmp/inspur/config-map"
)

type envConf struct {
	envName      string
	defaultValue string
	required     bool
}

// EnvInfo collects the env and relevant agentContext properties.
var envInfo = []envConf{
	{"CONFIG_PATH", ConfigPath, false},
	{"HTTP_PORT", HttpPort, false},
	{"UNIX_SOCKET_PATH", UnixSocketPath, false},
}

type Config struct {
	CommitVersion string
	CommitTime    string
	AppVersion    string
	GoMaxProcs    int

	// flags
	ConfigPath string

	// env
	LogLevel      string
	EnabledMetric bool

	HttpPort                 string
	MetricHttpPort           string
	GopsListenPort           string
	PyroscopeAddress         string
	ClusterDefaultIPv4IPPool []string `yaml:"clusterDefaultIPv4IPPool"`
	ClusterDefaultIPv6IPPool []string `yaml:"clusterDefaultIPv6IPPool"`
	IPPoolMaxAllocatedIPs    int
	WaitSubnetPoolTime       int
	WaitSubnetPoolMaxRetries int

	// configmap
	IpamUnixSocketPath string `yaml:"ipamUnixSocketPath"`
	NetworkMode        string `yaml:"networkMode"`
	EnableIPv4         bool   `yaml:"enableIPv4"`
	EnableIPv6         bool   `yaml:"enableIPv6"`
}

var ConfigInstance = Config{
	HttpPort:           HttpPort,
	IpamUnixSocketPath: UnixSocketPath,
	ConfigPath:         ConfigPath,
}

// ParseConfiguration set the env to AgentConfiguration
func ParseConfiguration() error {
	var result string

	for i := range envInfo {
		env, ok := os.LookupEnv(envInfo[i].envName)
		if ok {
			result = strings.TrimSpace(env)
		} else {
			// if no env and required, set it to default value.
			result = envInfo[i].defaultValue
		}
		if len(result) == 0 {
			if envInfo[i].required {
				logging.Panicf("empty value of %s,it is required", envInfo[i].envName)
			} else {
				// if no env and none-required, just use the empty value.
				continue
			}
		}
	}

	return nil
}

// LoadConfigmap reads configmap data from cli flag config-path
func LoadConfigmap() error {
	configmapBytes, err := os.ReadFile(ConfigInstance.ConfigPath)
	if nil != err {
		//return logging.Errorf("failed to read configmap file, error: %v", err)
	}

	err = yaml.Unmarshal(configmapBytes, &ConfigInstance)
	if nil != err {
		//return logging.Errorf("failed to parse configmap, error: %v", err)
	}

	if ConfigInstance.IpamUnixSocketPath == "" {
		ConfigInstance.IpamUnixSocketPath = UnixSocketPath
	}

	return nil
}
