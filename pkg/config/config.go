package config

import (
	"encoding/json"
	logging "github.com/Inspur-Data/k8-ipam/pkg/logging"
)

func ParseConfig(args []byte) (*CNIConf, error) {
	cniConfig := &CNIConf{}
	err := json.Unmarshal(args, cniConfig)
	if err != nil {
		logging.Debugf("json unmarshal failed: %v", err)
		return nil, logging.Errorf("json unmarshal failed: %v", err)
	}
	if cniConfig.IPAM.LogFile != "" {
		logging.SetLogFile(cniConfig.IPAM.LogFile)
	}
	if cniConfig.IPAM.LogLevel != "" {
		logging.SetLogLevel(cniConfig.IPAM.LogLevel)
	}

	if cniConfig.IPAM == nil {
		return nil, logging.Errorf("IPAM config is nil")
	}
	for _, version := range SupportCniVersion {
		if cniConfig.CNIVersion == version {
			return cniConfig, nil
		}
	}
	return nil, logging.Errorf("unsupported cni version: %v,the supported cni version %v", cniConfig.CNIVersion, SupportCniVersion)
}
