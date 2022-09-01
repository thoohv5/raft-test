package raft

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type raftEntity struct {
	o *options

	isLeader bool
	r        *raft.Raft
}

type IRaft interface {
	IsLeader() bool
	Set(entity *Entity) (err error)
	Join(peerAddr string)
}

func New(opts ...Option) (IRaft, error) {

	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	o.dataDir = filepath.Join(o.dataDir, strings.Split(o.addr, ":")[1])
	if _, err := os.Stat(o.dataDir); os.IsNotExist(err) {
		err = nil
		if err = os.MkdirAll(o.dataDir, 0777); err != nil {
			return nil, err
		}
	}

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(o.addr)
	raftConfig.Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "raft:",
		Output: os.Stdout,
	})
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(o.dataDir,
		"raft-log.bolt"))
	if nil != err {
		return nil, err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(o.dataDir,
		"raft-stable.bolt"))
	if nil != err {
		return nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(o.dataDir, 1, os.Stdout)
	if nil != err {
		return nil, err
	}

	transport, err := newRaftTransport(o)
	if nil != err {
		return nil, err
	}

	raftNode, err := raft.NewRaft(
		raftConfig,
		NewFSM(o.cache),
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		return nil, err
	}

	re := &raftEntity{
		o: o,
		r: raftNode,
	}

	if o.bootstrap {
		re.isLeader = true
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(configuration)
	} else {
		if err = re.joinRaftCluster(); err != nil {
			return nil, err
		}
	}

	go func() {
		for leaderC := range re.r.LeaderCh() {
			re.isLeader = leaderC
		}
	}()

	return re, nil

}

func (re *raftEntity) joinRaftCluster() error {
	url := fmt.Sprintf("http://%s/join?peerAddress=%s",
		re.o.joinAddr,
		re.o.addr)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(body) != "success" {
		return errors.New(fmt.Sprintf("Error joining cluster: %s", body))
	}
	return nil
}

func (re *raftEntity) Join(peerAddr string) {
	re.r.AddVoter(raft.ServerID(peerAddr), raft.ServerAddress(peerAddr), 0, 0)
}

func (re *raftEntity) Set(entity *Entity) (err error) {
	marshal, err := json.Marshal(entity)
	if err != nil {
		return
	}

	applyRet := re.r.Apply(marshal, 5*time.Second)
	if err = applyRet.Error(); nil != err {
		return
	}
	return
}

func (re *raftEntity) IsLeader() bool {
	return re.isLeader
}
