package server

import "HKey/pkg"

// 日志处理

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogItem
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

// AppendEntries 添加日志条目 供RPC调用
func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.currentTerm
	if args.Term == rf.currentTerm { // 只有Follower才会被调用AppendEntries
		rf.role = FOLLOWER
		rf.timer.Reset(randTime()) // term相同重置计时
	}
	rf.keepOrFollow(args.Term)
	if args.Term >= rf.currentTerm &&
		args.PrevLogIndex < len(rf.log) && // 如果大于则说明该节点至少缺少了一次日志
		args.PrevLogTerm == rf.log[args.PrevLogIndex].Term { // 判断存在日志一致的index
		if args.PrevLogIndex+1 != len(rf.log) || args.Entries != nil {
			// 从日志一致的地方开始向后覆盖
			rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...)
			rf.persist() // 同时将日志写入文件
		}
		if args.LeaderCommit > rf.commitIndex {
			if args.LeaderCommit < len(rf.log) {
				rf.commitIndex = args.LeaderCommit
			} else {
				rf.commitIndex = len(rf.log)
			}
		}
		rf.leader = args.LeaderId
		for _, entry := range args.Entries { // 处理条目
			rf.handleCommand(entry.Command)
		}
		reply.Success = true
	}
	return nil
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply, ch chan bool) {
	if err := rf.peers[server].Call("Raft.AppendEntries", args, reply); err != nil {
		return
	}
	rf.mu.Lock()
	rf.keepOrFollow(reply.Term)
	if rf.role == LEADER && reply.Success {
		rf.nextIndex[server] = args.PrevLogIndex + len(args.Entries) + 1
		rf.matchIndex[server] = args.PrevLogIndex + len(args.Entries)
		// pkg.DPrintf("节点%d接收了日志\n", server)
	} else if rf.role == LEADER && !reply.Success { // 节点拒绝了日志
		pkg.DPrintf("节点%d拒绝了日志，将在下次心跳时重新发送\n", server)
		rf.nextIndex[server]--
		// for rf.nextIndex[server] >= 0 && rf.log[rf.nextIndex[server]].Term == reply.Term {
		// 	rf.nextIndex[server]-- // 寻找日志一致的索引
		// }
	}
	rf.mu.Unlock()
	ch <- true
}
