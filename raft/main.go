package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	raftCount         = 3 //定义节点数量
	timeout           = 3 //选举超时时间 单位 秒
	lastHeartBeatTime = 7 //心跳检测超时时间
	heartBeatTimes    = 3 //心跳检测频率

	nodeTable    = make(map[string]string) //节点池
	httpTable    = make(map[string]string) //http接口
	MessageStore = make(map[int]string)    //用于储存消息
)

func main() {
	nodeTable = map[string]string{
		"A": ":9000",
		"B": ":9001",
		"C": ":9002",
		"D": ":9003",
		"E": ":9004",
	}
	httpTable = map[string]string{
		"A": ":8000",
		"B": ":8001",
		"C": ":8002",
		"D": ":8003",
		"E": ":8004",
	}
	time.Sleep(10 * time.Second)
	if len(os.Args) < 2 {
		if os.Args[1] == "A" || os.Args[1] == "B" || os.Args[1] == "C" {
			log.Fatal("程序参数不正确")
		}
	}
	id := os.Args[1]

	raft := NewRaft(id, nodeTable[id], httpTable[id])
	//启用 RPC,注册 raft
	go RpcRegister(raft)
	//开启心跳检测
	go raft.Heartbeat()
	//开启一个 http 监听

	go raft.HttpListen()

	time.Sleep(1 * time.Second)
Circle:
	//开始选举
	go func() {
		for {
			//成为候选人
			if raft.becomeCandidate() {
				//成为候选人节点后,向其他节点要选票来进行选举
				if raft.election() {
					break
				} else {
					continue
				}
			} else {
				break
			}
		}
	}()

	//进行心跳检测
	for {
		time.Sleep(time.Millisecond * 5000)
		//TODO
		if raft.lastHeartBeatTime != 0 && (millisecond()-raft.lastHeartBeatTime) > int64(raft.timeout*1000) {
			fmt.Printf("心跳检测超时，已超过%d秒\n", raft.timeout)
			fmt.Println("即将重新开启选举")
			raft.reDefault()
			raft.setCurrentLeader("-1")
			raft.lastHeartBeatTime = 0
			goto Circle
		}
	}
}
