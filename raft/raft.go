package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Raft struct {
	node              *NodeInfo
	vote              int        //本节点获得的投票数
	lock              sync.Mutex //线程锁
	me                string     //节点编号
	currentTerm       int        //当前任期
	votedFor          string     //为哪个节点投票
	state             int        //当前节点状态 0 follower  1 candidate  2 leader
	lastMessageTime   int64      //发送最后一条消息的时间
	lastHeartBeatTime int64      //最后一条消息的时间
	currentLeader     string     //当前节点的领导
	timeout           int        //心跳超时时间
	voteCh            chan bool  //接受投票成功通道
	heartBeat         chan bool  //心跳信号
}

type NodeInfo struct {
	Id   string
	Port string
	HP   string
}

type Message struct {
	Msg   string
	MsgId int
}

func NewRaft(id, port, hp string) *Raft {
	rf := &Raft{node: &NodeInfo{Id: id, Port: port, HP: hp}}
	//当前节点获得票数
	rf.setVote(0)
	//编号
	rf.me = id
	//给0  1  2三个节点投票，给谁都不投
	rf.setVoteFor("-1")
	//0 follower
	rf.setStatus(0)
	//最后一次心跳检测时间
	rf.lastHeartBeatTime = 0
	rf.timeout = lastHeartBeatTime
	//最初没有领导
	rf.setCurrentLeader("-1")
	//设置任期
	rf.setTerm(0)
	//投票通道
	rf.voteCh = make(chan bool)
	//心跳通道
	rf.heartBeat = make(chan bool)
	return rf
}

//修改节点为候选人状态
func (r *Raft) becomeCandidate() bool {
	fmt.Println("1111111")
	t := randRange(1500, 5000)
	//休眠随机时间后,再开始成为候选人
	time.Sleep(time.Duration(t) * time.Millisecond)
	//如果发现本节点已经投过票,或者已经存在领导者,则不用变身候选状态
	if r.state == 0 && r.currentLeader == "-1" && r.votedFor == "-1" {
		//将节点状态变为 1
		r.setStatus(1)
		//设置为哪个节点投票
		r.setVoteFor(r.me)
		//节点任期加 1
		r.setTerm(r.currentTerm + 1)
		//当前没有领导
		r.setCurrentLeader("-1")
		//为自己投票
		r.voteAdd()
		fmt.Println("本节点已变更为候选人状态")
		fmt.Printf("当前得票数：%d\n", r.vote)
		//开启选举通道
		return true
	} else {
		return false
	}
}

//进行选举
func (r *Raft) election() bool {
	fmt.Println("开始进行领导者选举，向其他节点进行广播")
	go r.broadcast("Raft.Vote", r.node, func(ok bool) {
		r.voteCh <- ok
	})
	for {
		select {
		case <-time.After(time.Second * time.Duration(timeout)):
			fmt.Println("领导者选举超时,节点变更为追随者状态")
			r.reDefault()
			return false
		case ok := <-r.voteCh:
			fmt.Println("3333333")
			if ok {
				r.voteAdd()
				fmt.Println("获得来自其他节点的投票, 当前得票数: ", r.vote)
			}
			if r.vote > raftCount/2 && r.currentLeader == "-1" {
				fmt.Println("获得超过网络节点二分之一的得票数,本节点被选举成为了 leader")
				r.setStatus(2)
				r.setCurrentLeader(r.me)
				fmt.Println("向其他节点进行广播...")
				go r.broadcast("Raft.ConfirmationLeader", r.node, func(ok bool) {
					fmt.Println(ok)
				})
				r.heartBeat <- true
			}
		}
	}
}

//心跳检测方法
func (r *Raft) Heartbeat() {
	//如果收到通道开启的信息,将会向其他节点进行固定频率的心跳检测
	if <-r.heartBeat {
		fmt.Println("5555555")
		for {
			fmt.Printf("本节点 %v 开始发送心跳检测\n", r.node.Id)
			r.broadcast("Raft.HeartbeatRe", r.node, func(ok bool) {
				fmt.Println("收到回复: ", ok)
			})
			r.lastHeartBeatTime = millisecond()
			time.Sleep(time.Second * time.Duration(heartBeatTimes))
		}
	}
}

//设置任期
func (r *Raft) setTerm(term int) {
	r.lock.Lock()
	r.currentTerm = term
	r.lock.Unlock()
}

//设置投票数量
func (r *Raft) setVote(num int) {
	r.lock.Lock()
	r.vote = num
	r.lock.Unlock()
}

//设置为谁投票
func (r *Raft) setVoteFor(id string) {
	r.lock.Lock()
	r.votedFor = id
	r.lock.Unlock()
}

//设置当前身份
func (r *Raft) setStatus(status int) {
	r.lock.Lock()
	r.state = status
	r.lock.Unlock()
}

//设置当前领导
func (r *Raft) setCurrentLeader(leader string) {
	r.lock.Lock()
	r.currentLeader = leader
	r.lock.Unlock()
}

//获取当前时间的毫秒数
func millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//产生随机值
func randRange(min, max int64) int64 {
	//用于心跳信号的时间
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

//投票累加
func (r *Raft) voteAdd() {
	r.lock.Lock()
	r.vote++
	r.lock.Unlock()
}

//恢复默认设置
func (r *Raft) reDefault() {
	r.setVote(0)
	r.setVoteFor("-1")
	r.setStatus(0)
}
