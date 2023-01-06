package server

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
	if args.Term == rf.currentTerm {
		rf.role = FOLLOWER
		rf.timer.Reset(randTime()) // term相同重置计时
	}
	rf.keepOrFollow(args.Term)
	if args.Term >= rf.currentTerm &&
		args.PrevLogIndex < len(rf.log) &&
		args.PrevLogTerm == rf.log[args.PrevLogIndex].Term { // term和log要匹配
		if args.PrevLogIndex+1 != len(rf.log) || args.Entries != nil {
			rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...) // 删除不匹配并添加未持有的日志
			rf.persist()
		}
		if args.LeaderCommit > rf.commitIndex {
			if args.LeaderCommit < len(rf.log) {
				rf.commitIndex = args.LeaderCommit
			} else {
				rf.commitIndex = len(rf.log)
			}
		}
		rf.leader = args.LeaderId
		for _, entry := range args.Entries {
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
	} else if rf.role == LEADER && !reply.Success {
		rf.nextIndex[server]--
		for rf.nextIndex[server] >= 0 && rf.log[rf.nextIndex[server]].Term == reply.Term {
			rf.nextIndex[server]-- // 跳过同一term的log
		}
	}
	rf.mu.Unlock()
	ch <- true
}
