package server

import "HKey/pkg"

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
		rf.persist()               // 同时将日志写入文件
		rf.timer.Reset(randTime()) // 重置选举计时
		reply.VoteGranted = true
		pkg.DPrintf("收到了节点%d投票邀请，并进行了投票\n", args.CandidateId)
	} else {
		pkg.DPrintf("收到了节点%d投票邀请，并拒绝了投票\n", args.CandidateId)
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
		pkg.DPrintf("节点%d将票投给了me，当前票数%d\n", server, rf.votes)
	} else {
		pkg.DPrintf("节点%d没有将票投给me，当前票数%d，当前身份%v\n", server, rf.votes, rf.role)
	}
	rf.mu.Unlock()
	ch <- true
}
