package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/thoohv5/raft-test/internal/cache"
	"github.com/thoohv5/raft-test/internal/raft"
)

var (
	httpAddr  string
	raftAddr  string
	joinAddr  string
	dataDir   string
	bootstrap bool
)

func init() {
	flag.StringVar(&httpAddr, "httpAddr", "", "httpAddr")
	flag.StringVar(&raftAddr, "raftAddr", "", "raftAddr")
	flag.StringVar(&joinAddr, "joinAddr", "", "joinAddr")
	flag.StringVar(&dataDir, "dataDir", "", "dataDir")
	flag.BoolVar(&bootstrap, "bootstrap", false, "bootstrap")
}

func main() {
	flag.Parse()

	fmt.Println(httpAddr, raftAddr, joinAddr, dataDir, bootstrap)

	router := gin.Default()

	cacheManage := cache.New()
	raftNode, err := raft.New(
		raft.WithAddr(raftAddr),
		raft.WithJoinAddr(joinAddr),
		raft.WithDataDir(dataDir),
		raft.WithBootstrap(bootstrap),
		raft.WithCache(cacheManage))
	if nil != err {
		panic(err)
	}

	router.GET("/set", func(c *gin.Context) {

		if !raftNode.IsLeader() {
			c.String(http.StatusOK, "no leader")
			return
		}

		key := c.Query("key")
		value := c.Query("value")
		err := raftNode.Set(&raft.Entity{
			Key:   key,
			Value: value,
		})
		if err != nil {
			c.JSON(http.StatusOK, err)
			return
		}
		c.String(http.StatusOK, "success")
	})

	router.GET("/get", func(c *gin.Context) {
		key := c.Query("key")
		value := cacheManage.Get(key)
		c.JSON(http.StatusOK, value)
	})

	router.GET("/join", func(c *gin.Context) {
		raftNode.Join(c.Query("peerAddress"))
		c.String(http.StatusOK, "success")
	})

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		err := srv.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
		log.Println(err)
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
