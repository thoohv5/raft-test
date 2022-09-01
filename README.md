# raft-test
this is a raft test

# 编译
go build  -o build cmd/raft/main.go

# node1
./build/main -httpAddr="127.0.0.1:6000" -raftAddr="127.0.0.1:7000" -dataDir="./test-data" -bootstrap=true

# node2
./build/main -httpAddr="127.0.0.1:6001" -raftAddr="127.0.0.1:7001" -joinAddr="127.0.0.1:6000" -dataDir="./test-data"

# node3
./build/main -httpAddr="127.0.0.1:6002" -raftAddr="127.0.0.1:7002" -joinAddr="127.0.0.1:6000" -dataDir="./test-data"