package server

type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// RequestVote 请求投票 用于RPC
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Term = rf.currentTerm
	rf.keepOrFollow(args.Term)
	if args.Term >= rf.currentTerm &&
		(rf.votedFor == -1 || rf.votedFor == args.CandidateId) &&
		(args.LastLogTerm > rf.log[len(rf.log)-1].Term ||
			args.LastLogTerm == rf.log[len(rf.log)-1].Term &&
				args.LastLogIndex >= len(rf.log)-1) { // 日志至少一样新
		rf.votedFor = args.CandidateId
		rf.timer.Reset(randTime()) // 重置选举计时
		reply.VoteGranted = true
	}
	return nil
}

func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply, ch chan bool) {
	err := rf.peers[server].Call("Raft.RequestVote", args, reply)
	if err != nil {
		return
	}
	rf.mu.Lock()
	rf.keepOrFollow(reply.Term)
	// 如果该节点将票投给了me
	if rf.role == CANDIDATE && reply.VoteGranted {
		rf.votes += 1
	}
	rf.mu.Unlock()
	ch <- true
}
