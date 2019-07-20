package service

import (
	"github.com/camsiabor/qservice/impl/gateway/zookeeper"
	"sync"
)

var zkwMutex sync.Mutex
var zkwInstances = map[string]*zookeeper.ZooWatcher{}

func ZooWatcherGet(id string, endpoint string) (instance *zookeeper.ZooWatcher, err error) {

	zkwMutex.Lock()
	instance = zkwInstances[id]
	zkwMutex.Unlock()

	if instance == nil {
		zkwMutex.Lock()
		defer zkwMutex.Unlock()
		instance = zkwInstances[id]
		if instance == nil {
			instance = &zookeeper.ZooWatcher{}
			instance.Id = id
			zkwInstances[id] = instance
		}
	}

	if len(endpoint) > 0 && !instance.IsConnected() {
		instance.Endpoints = []string{endpoint}
		err = instance.Start(nil)
	}
	return instance, err
}

func ZooWatcherClose(id string) {

	zkwMutex.Lock()
	defer zkwMutex.Unlock()
	var instance = zkwInstances[id]
	if instance == nil {
		return
	}
	_ = instance.Stop(nil)
	delete(zkwInstances, id)

}
