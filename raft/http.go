package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/rpc"
	"strconv"
)

func (r *Raft) getRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if len(req.Form["message"]) > 0 && r.currentLeader != "-1" {
		message := req.Form["message"][0]
		m := new(Message)
		m.MsgId = getRandom()
		m.Msg = message
		//收到消息后,直接转发到领导者
		port := nodeTable[r.currentLeader]
		rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+port)
		if err != nil {
			log.Panic(err)
		}
		b := false
		err = rp.Call("Raft.LeaderReceiveMessage", m, &b)
		if err != nil {
			log.Panic(err)
		}
		fmt.Println("消息是否已发送到领导者: ", b)
		w.Write([]byte("ok!!!"))
	}
}

func (r *Raft) getMsg(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	if len(req.Form["id"]) > 0 {
		id, _ := strconv.Atoi(req.Form["id"][0])
		res := MessageStore[id]
		w.Write([]byte(res))
	}
}

func (r *Raft) HttpListen() {
	http.HandleFunc("/req", r.getRequest)
	http.HandleFunc("/get", r.getMsg)
	fmt.Printf("监听 %v", r.node.HP)
	if err := http.ListenAndServe(r.node.HP, nil); err != nil {
		fmt.Println(err)
		return
	}
}

//返回一个十位数的随机数，作为消息idgit
func getRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := rand.Int(rand.Reader, x)
		if err != nil {
			log.Panic(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}
