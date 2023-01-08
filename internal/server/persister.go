package server

import (
	"io/ioutil"
	"os"
	"sync"
)

type Persister struct {
	mu        sync.Mutex
	data_path string
	raftstate []byte
}

func clone(orig []byte) []byte {
	x := make([]byte, len(orig))
	copy(x, orig)
	return x
}

func (ps *Persister) SaveRaftState(state []byte) {
	fp, err := os.OpenFile(ps.data_path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.raftstate = clone(state)
	fp.Write(state)
	fp.Close()
}

func (ps *Persister) ReadRaftState() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	fp, err := os.OpenFile(ps.data_path, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	ps.raftstate, err = ioutil.ReadAll(fp)
	if err != nil {
		panic(err)
	}
	return clone(ps.raftstate)
}

func (ps *Persister) RaftStateSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.raftstate)
}

func NewPersister(data_path string) *Persister {
	ps := &Persister{}
	var err error
	ps.data_path = data_path + "log.data"
	if err != nil {
		panic(err)
	}
	return ps
}
