package server

import (
	"HKey/internal/storage"
	"fmt"
	"math/rand"
	"net/http"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

const (
	FOLLOWER  = 0
	CANDIDATE = 1
	LEADER    = 2
)

// 定义各种间隔时间
const (
	electionTimeoutMin = 300 * time.Millisecond
	electionTimeoutMax = 600 * time.Millisecond
	heartbeatTimeout   = 25 * time.Millisecond
	checkTimeout       = 5 * time.Millisecond
)

func randTime() time.Duration {
	diff := (electionTimeoutMax - electionTimeoutMin).Milliseconds()
	return electionTimeoutMin + time.Duration(rand.Intn(int(diff)))*time.Millisecond
}

func wait(n int, ch chan bool) {
	for i := 1; i < n; i++ {
		select {
		case <-ch:
		case <-time.After(checkTimeout):
			return
		}
	}
}

// Raft 定义Raft的主体
type Raft struct {
	mu          sync.Mutex
	store       *storage.HKey
	peers       []*rpc.Client // 与其它节点的连接
	leader      int           // leader 地址
	me          int
	commitIndex int // 日志提交index
	lastApplied int
	currentTerm int       // 当前Term
	votedFor    int       // 投票目标
	log         []LogItem // 日志
	role        int       // 身份
	votes       int       // 得票数
	nextIndex   []int     // 下一个发送给follower的条目索引
	matchIndex  []int
	timer       *time.Timer
}

func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term := rf.currentTerm
	isLeader := rf.role == LEADER
	return term, isLeader
}

// 修改当前节点状态
func (rf *Raft) keepOrFollow(term int) {
	if term > rf.currentTerm { // 新的leader已经出现
		rf.currentTerm = term
		rf.votedFor = -1 // 重置投票
		rf.role = FOLLOWER
	}
}

func NewRaft(data_path string, address string, me int) *Raft {
	rf := &Raft{}
	rf.store = storage.NewHKey(data_path)
	rf.me = me
	rf.votedFor = -1
	rf.log = make([]LogItem, 1)
	rf.role = FOLLOWER
	rf.timer = time.NewTimer(randTime())

	err := rpc.Register(rf) // 注册Raft
	if err != nil {
		panic(err)
	}
	rpc.HandleHTTP()
	go http.ListenAndServe(address, nil)

	// go func() {
	// 	for rf.me != -1 {
	// 		time.Sleep(checkTimeout)
	// 		rf.mu.Lock()
	// 		for rf.me != -1 && rf.commitIndex > rf.lastApplied {
	// 			rf.lastApplied++
	// 			applyCh <- ApplyMsg{Index: rf.lastApplied, Command: rf.log[rf.lastApplied].Command}
	// 		}
	// 		rf.mu.Unlock()
	// 	}
	// }()

	return rf
}

// Start 由引导程序启动集群 RPC调用
func (rf *Raft) Start(nodesInfo []InfoArgs, ans *string) error {
	// 首先是与各节点建立连接
	rf.peers = make([]*rpc.Client, len(nodesInfo))
	rf.nextIndex = make([]int, len(nodesInfo))
	rf.matchIndex = make([]int, len(nodesInfo))
	for idx, arg := range nodesInfo {
		if idx == rf.me {
			continue
		}
		conn, err := rpc.DialHTTP(arg.Protocol, arg.Address)
		if err != nil {
			return err
		}
		rf.peers[idx] = conn
	}
	// 然后开始工作
	go rf.work()
	fmt.Printf("Raft网络 节点：%d构建完成\n", rf.me)
	return nil
}

// raft集群下的工作
func (rf *Raft) work() {
	for rf.me != -1 {
		<-rf.timer.C
		switch rf.role {
		case FOLLOWER: // 继续工作
			rf.mu.Lock()
			rf.role = CANDIDATE
			rf.timer.Reset(0)
			rf.mu.Unlock()

		case CANDIDATE: // 当变为candidate时要开始vote
			rf.mu.Lock()
			rf.currentTerm++
			rf.votedFor = rf.me
			rf.votes = 1
			rf.timer.Reset(randTime())
			m := len(rf.log)
			rf.mu.Unlock()
			args := RequestVoteArgs{rf.currentTerm, rf.me,
				m - 1, rf.log[m-1].Term}
			reply := new(RequestVoteReply)
			ch := make(chan bool)
			// 对每个节点都进行发送
			for _, i := range rand.Perm(len(rf.peers)) {
				rf.mu.Lock()
				if rf.me != -1 && rf.role == CANDIDATE && i != rf.me {
					go rf.sendRequestVote(i, args, reply, ch)
				}
				rf.mu.Unlock()
			}

			wait(len(rf.peers), ch) // 等待所有发收成功或超时

			rf.mu.Lock()
			if rf.me != -1 && rf.role == CANDIDATE && 2*rf.votes > len(rf.peers) { // 成功竞选
				rf.role = LEADER
				rf.leader = rf.me
				rf.timer.Reset(0)
				for i := 0; i < len(rf.peers); i++ {
					rf.nextIndex[i] = m
				}
				fmt.Printf("节点%d目前成为Leader\n", rf.me)
			}
			rf.mu.Unlock()

		case LEADER: // leader处理日志
			rf.mu.Lock()
			rf.timer.Reset(heartbeatTimeout)
			m := len(rf.log)
			rf.mu.Unlock()
			ch := make(chan bool)
			for _, i := range rand.Perm(len(rf.peers)) {
				rf.mu.Lock()
				if rf.me != -1 && rf.role == LEADER && i != rf.me {
					args := AppendEntriesArgs{rf.currentTerm, rf.me,
						rf.nextIndex[i] - 1, rf.log[rf.nextIndex[i]-1].Term,
						nil, rf.commitIndex}
					if rf.nextIndex[i] < m {
						args.Entries = make([]LogItem, m-rf.nextIndex[i])
						copy(args.Entries, rf.log[rf.nextIndex[i]:m])
					}
					reply := AppendEntriesReply{}
					go rf.sendAppendEntries(i, args, &reply, ch)
				}
				rf.mu.Unlock()
			}

			wait(len(rf.peers), ch) // 等待所有发收成功或超时

			rf.mu.Lock()
			N := m - 1
			if rf.me != -1 && rf.role == LEADER && N > rf.commitIndex && rf.log[N].Term == rf.currentTerm {
				count := 1
				for i := 0; i < len(rf.peers); i++ {
					if i != rf.me && rf.matchIndex[i] >= N {
						count++
					}
				}
				if 2*count > len(rf.peers) { // 确认提交
					rf.commitIndex = N
				}
			}
			rf.mu.Unlock()
		}
	}

}

// Command 用于处理用户命令  注意只有leader才能处理！
func (rf *Raft) Command(command string, ans *string) error {
	fmt.Println("---------------------------------------------")
	fmt.Printf("Command:%s\nDate:%v\n", command, time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("---------------------------------------------")
	if rf.role != LEADER {
		fmt.Printf("非Leader转交command\n")
		err := rf.peers[rf.leader].Call("Raft.Command", command, ans)
		if err != nil {
			return err
		}
	} else {
		*ans = rf.handleCommand(command)
		rf.log = append(rf.log, LogItem{Command: command, Term: rf.currentTerm}) // 加入日志，在下次心跳时分发给节点
	}
	return nil
}

func (rf *Raft) handleCommand(command string) string {
	var ans string
	words := strings.Split(command, " ")
	action := words[0]
	switch action {
	case "set":
		ans = rf.set(words)
	case "get":
		ans = rf.get(words)
	case "del":
		ans = rf.del(words)
	case "exists":
		ans = rf.exists(words)
	default:
		fmt.Println("Unknown command")
	}
	return ans
}

// 处理set命令的解析
func (rf *Raft) set(words []string) string {
	if len(words) != 3 {
		return "error arguments"
	}
	var ans string
	var args = storage.SetArgs{
		Key: words[1], Value: words[2],
	}
	err := rf.store.Set(args, &ans)
	if err != nil {
		panic(err)
	}
	return ans
}

func (rf *Raft) get(words []string) string {
	if len(words) != 2 {
		return "error arguments"
	}
	var ans string
	err := rf.store.Get(words[1], &ans)
	if err != nil {
		fmt.Println(err)
	}
	if ans == "nil" {
		fmt.Println("(nil)")
		ans = "(nil)"
	} else {
		fmt.Printf("\"%s\"\n", ans)
		ans = "\"" + ans + "\""
	}
	return ans
}

func (rf *Raft) del(words []string) string {
	if len(words) != 2 {
		return "error arguments"
	}
	var ans string
	err := rf.store.Del(words[1], &ans)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ans)
	return ans
}

func (rf *Raft) exists(words []string) string {
	if len(words) != 2 {
		return "error arguments"
	}
	var ans string
	err := rf.store.Exists(words[1], &ans)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ans)
	return ans
}
