package leader

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"

	"github.com/hashicorp/raft"
)

type Leader struct {
	Mtx        sync.RWMutex
	LeaderName string
}

func (l *Leader) Apply(d *raft.Log) interface{} {
	l.Mtx.Lock()
	defer l.Mtx.Unlock()
	w := string(d.Data)
	l.LeaderName = w
	log.Printf("Updated leader is %s", w)
	return nil
}

func (l *Leader) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{l.LeaderName}, nil
}

func (l *Leader) Restore(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	leaderName := string(b)
	l.LeaderName = leaderName
	return nil
}

type snapshot struct {
	leaderName string
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write([]byte(s.leaderName))
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("sink.Write(): %v", err)
	}
	return sink.Close()
}

func (s *snapshot) Release() {
}
