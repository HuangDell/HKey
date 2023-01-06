package server

import (
	"os"
	"sync"
)

type Persister struct {
	mu        sync.Mutex
	raftstate []byte
	fp        *os.File
}

func clone(orig []byte) []byte {
	x := make([]byte, len(orig))
	copy(x, orig)
	return x
}

// func (ps *Persister) Copy() *Persister {
// 	ps.mu.Lock()
// 	defer ps.mu.Unlock()
// 	np := NewPersister()
// 	np.raftstate = ps.raftstate
// 	return np
// }

func (ps *Persister) SaveRaftState(state []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.raftstate = clone(state)
	ps.fp.Write(state)
}

func (ps *Persister) ReadRaftState() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.fp.Read(ps.raftstate)
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
	ps.fp, err = os.OpenFile(data_path+"log.data", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return ps
}
