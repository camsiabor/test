package main

import (
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

var zkconns = map[string]*zk.Conn{}

func GetZookeeper(id string, endpoint string) (*zk.Conn, error) {
	var err error
	var conn = zkconns[id]
	if conn == nil {
		conn, _, err = zk.Connect([]string{endpoint}, time.Duration(15)*time.Second)
		if err != nil {
			return nil, err
		}
		zkconns[id] = conn
	}
	return conn, nil
}
