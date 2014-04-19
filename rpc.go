package main

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"net"
	"net/rpc"
)

type SnowflakeRPC struct {
	idWorkers map[int64]*IdWorker
}

// StartRPC start rpc listen.
func InitRPC() error {
	// TODO check
	idWorkers := make(map[int64]*IdWorker, len(MyConf.WorkerId))
	for _, workerId := range MyConf.WorkerId {
		idWorker, err := NewIdWorker(MyConf.DatacenterId, workerId)
		if err != nil {
			glog.Errorf("NewIdWorker(%d, %d) error(%v)", MyConf.DatacenterId, workerId)
			return err
		}
		idWorkers[workerId] = idWorker
		// TODO register
	}
	s := &SnowflakeRPC{idWorkers: idWorkers}
	rpc.Register(s)
	for _, bind := range MyConf.RPCBind {
		glog.Infof("start listen rpc addr: \"%s\"", bind)
		go rpcListen(bind)
	}
	return nil
}

// rpcListen start rpc listen.
func rpcListen(bind string) {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		glog.Errorf("net.Listen(\"tcp\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		glog.Infof("rpc addr: \"%s\" close", bind)
		if err := l.Close(); err != nil {
			glog.Errorf("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// SnowflakeId generate a id.
func (s *SnowflakeRPC) SnowflakeId(workerId int64, id *int64) error {
	if worker, ok := s.idWorkers[workerId]; !ok {
		glog.Warningf("workerId: %d not register", workerId)
		return errors.New(fmt.Sprintf("snowflake workerId: %d don't register in this service", workerId))
	} else {
		if tid, err := worker.NextId(); err != nil {
			glog.Errorf("worker.NextId() error(%v)", err)
			return err
		} else {
			id = &tid
		}
	}
	return nil
}

// DatacenterId return the services's datacenterId.
func (s *SnowflakeRPC) DatacenterId(ignore int, dataCenterId *int64) error {
	dataCenterId = &MyConf.DatacenterId
	return nil
}
