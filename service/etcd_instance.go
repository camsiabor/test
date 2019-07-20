package service

import (
	"github.com/camsiabor/qservice/impl/gateway/etcd"
	"sync"
)

var etcdMutex sync.Mutex
var etcdConnections = map[string]*etcd.EtcdWatcher{}

func EtcdGet(id string, endpoint string) (instance *etcd.EtcdWatcher, err error) {

	etcdMutex.Lock()
	instance = etcdConnections[id]
	etcdMutex.Unlock()

	if instance == nil {
		etcdMutex.Lock()
		defer etcdMutex.Unlock()
		instance = etcdConnections[id]
		if instance == nil {
			instance = &etcd.EtcdWatcher{}
			instance.Id = id
			etcdConnections[id] = instance
		}
	}

	if len(endpoint) > 0 && !instance.IsConnected() {
		instance.Endpoints = []string{endpoint}
		err = instance.Start(nil)
	}
	return instance, err
}

func EtcdClose(id string) {

	etcdMutex.Lock()
	defer etcdMutex.Unlock()
	var instance = etcdConnections[id]
	if instance == nil {
		return
	}
	_ = instance.Stop(nil)
	delete(etcdConnections, id)

}
