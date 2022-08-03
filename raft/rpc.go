package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"time"
)

//rpc 服务注册
func RpcRegister(raft *Raft) {
	//注册一个服务器
	err := rpc.Register(raft)
	if err != nil {
		log.Panic(err)
	}
	port := raft.node.Port
	//把服务绑定到http协议上
	rpc.HandleHTTP()
	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("注册 rpc 服务失败", err)
	}
}

func (r *Raft) broadcast(method string, args interface{}, fun func(ok bool)) {
	//设置不要自己给自己广播
	for nodeId, port := range nodeTable {
		if nodeId == r.node.Id {
			continue
		}
		rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+port)
		if err != nil {
			fun(false)
			continue
		}
		var bo = false
		err = rp.Call(method, args, &bo)
		if err != nil {
			fun(false)
			continue
		}
		fun(bo)
	}
}

//投票
func (r *Raft) Vote(node NodeInfo, b *bool) error {
	fmt.Println("2222222")
	if r.votedFor != "-1" || r.currentLeader != "-1" {
		*b = false
	} else {
		r.setVoteFor(node.Id)
		fmt.Println("投票成功,已投 ", node.Id, " 节点")
		*b = true
	}
	return nil
}

//确认领导者
func (r *Raft) ConfirmationLeader(node NodeInfo, b *bool) error {
	fmt.Println("44444444")
	r.setCurrentLeader(node.Id)
	*b = true
	fmt.Println("已发现网络中的领导节点,", node.Id, "成为了领导者！")
	r.reDefault()
	return nil
}

//心跳检测回复
func (r *Raft) HeartbeatRe(node NodeInfo, b *bool) error {
	r.setCurrentLeader(node.Id)
	r.lastHeartBeatTime = millisecond()
	fmt.Printf("接收到来自领导节点 %v 的心跳检测\n", node.Id)
	fmt.Printf("当前时间为:%d\n", millisecond())
	*b = true
	return nil
}

//领导者接收到,追随者节点转发过来的消息
func (r *Raft) LeaderReceiveMessage(message Message, b *bool) error {
	fmt.Println("领导者节点接收到转发过来的消息 ", message.MsgId, "->", message.Msg)
	MessageStore[message.MsgId] = message.Msg
	*b = true
	fmt.Println("准备将消息进行广播...")
	num := 0
	go r.broadcast("Raft.ReceiveMessage", message, func(ok bool) {
		if ok {
			num++
		}
	})
	for {
		//TODO 如果没有过半呢
		if num > raftCount/2-1 {
			fmt.Println("全网已超过半数节点接收到消息", message.MsgId, "->", MessageStore[message.MsgId], " raft验证通过")
			r.lastMessageTime = millisecond()
			fmt.Println("准备将消息提交信息发送至客户端...")
			go r.broadcast("Raft.ConfirmationMessage", message, func(ok bool) {
			})
			break
		} else {
			//休息会儿
			time.Sleep(time.Millisecond * 100)
		}
	}
	return nil
}

func (r *Raft) ReceiveMessage(message Message, b *bool) error {
	fmt.Println("接受到领导者节点发来的信息 ", message.MsgId, "->", message.Msg)
	MessageStore[message.MsgId] = message.Msg
	*b = true
	fmt.Println("已回复接收到消息，待领导者确认后打印")
	return nil
}

func (r *Raft) ConfirmationMessage(message Message, b *bool) error {
	go func() {
		for {
			if v, ok := MessageStore[message.MsgId]; ok {
				fmt.Println("raft验证通过,可以打印消息 ", message.MsgId, "->", v)
				r.lastMessageTime = millisecond()
				break
			} else {
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()
	*b = true
	return nil
}
