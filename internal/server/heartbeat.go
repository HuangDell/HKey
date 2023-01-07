package server

import "HKey/pkg"

// HeartbeatArgs 心跳过程中传递的参数
type HeartbeatArgs struct {
	Term     int
	LeaderId int
}

// Heartbeat 心跳 供RPC调用
func (rf *Raft) Heartbeat(args HeartbeatArgs, reply *bool) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.Term == rf.currentTerm {
		rf.role = FOLLOWER
		rf.timer.Reset(randTime()) // 收到心跳，重置选举时间
	}
	rf.keepOrFollow(args.Term)
	return nil
}

func (rf *Raft) sendHeartbeat(server int, args HeartbeatArgs) {
	if err := rf.peers[server].Call("Raft.Heartbeat", args, nil); err != nil {
		pkg.DPrintf("节点%d未响应心跳,疑似已经掉线", server)
		return
	}
}
