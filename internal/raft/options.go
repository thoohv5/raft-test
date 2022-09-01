package raft

import "github.com/thoohv5/raft-test/internal/cache"

type options struct {
	addr      string
	joinAddr  string
	dataDir   string
	bootstrap bool

	cache cache.ICache
}

type Option func(o *options)

func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

func WithJoinAddr(joinAddr string) Option {
	return func(o *options) {
		o.joinAddr = joinAddr
	}
}

func WithDataDir(dataDir string) Option {
	return func(o *options) {
		o.dataDir = dataDir
	}
}

func WithBootstrap(bootstrap bool) Option {
	return func(o *options) {
		o.bootstrap = bootstrap
	}
}

func WithCache(cache cache.ICache) Option {
	return func(o *options) {
		o.cache = cache
	}
}
