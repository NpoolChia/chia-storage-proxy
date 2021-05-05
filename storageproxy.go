package main

import (
	"encoding/json"
	log "github.com/EntropyPool/entropy-logger"
	"github.com/NpoolChia/chia-storage-proxy/types"
	"github.com/NpoolRD/http-daemon"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type StorageProxyConfig struct {
	Port         int      `json:"port"`
	StorageHosts []string `json:"storage_hosts"`
}

type StorageProxy struct {
	config       StorageProxyConfig
	mutex        sync.Mutex
	postingHosts map[string]int
}

func NewStorageProxy(cfgFile string) *StorageProxy {
	proxy := &StorageProxy{
		postingHosts: map[string]int{},
	}

	buf, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot read config file %v: %v", cfgFile, err)
		return nil
	}

	err = json.Unmarshal(buf, &proxy.config)
	if err != nil {
		log.Errorf(log.Fields{}, "cannot parse config file %v: %v", cfgFile, err)
		return nil
	}

	return proxy
}

func (p *StorageProxy) Run() error {
	httpdaemon.RegisterRouter(httpdaemon.HttpRouter{
		Location: types.NewPlotAPI,
		Handler:  p.NewPlotRequest,
		Method:   "POST",
	})

	httpdaemon.Run(p.config.Port)

	return nil
}

func (p *StorageProxy) NewPlotRequest(w http.ResponseWriter, req *http.Request) (interface{}, string, int) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err.Error(), -1
	}

	input := types.NewPlotInput{}
	err = json.Unmarshal(b, &input)
	if err != nil {
		return nil, err.Error(), -2
	}

	_, err = os.Stat(input.PlotDir)
	if err != nil {
		return nil, err.Error(), -3
	}

	err = filepath.Walk(input.PlotDir, func(path string, info os.FileInfo, err error) error {
		b, _ := json.Marshal(info)
		log.Infof(log.Fields{}, "walk to %v in %v [%v]", input.PlotDir, path, string(b))
		return nil
	})
	if err != nil {
		return nil, err.Error(), -4
	}

	return nil, "", 0
}