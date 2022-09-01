package raft

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"

	"github.com/thoohv5/raft-test/internal/cache"
)

type fsm struct {
	cache cache.ICache
}

func NewFSM(cache cache.ICache) raft.FSM {
	return &fsm{cache: cache}
}

func (f *fsm) Apply(log *raft.Log) interface{} {
	led := &Entity{}
	if err := json.Unmarshal(log.Data, led); err != nil {
		return err
	}

	f.cache.Set(led.Key, led.Value)
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{f.cache}, nil
}

func (f *fsm) Restore(snapshot io.ReadCloser) error {
	return f.cache.UnMarshal(snapshot)
}
