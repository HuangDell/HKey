package pkg

// import "HKey/internal/server"

// Clear 将数组置为0
func Clear(arr []byte) {
	for i := 0; i < len(arr); i++ {
		arr[i] = 0
	}
}

//
// type Queue struct {
// 	ele []server.LogItem
// }
//
// func (q *Queue) Push(e server.LogItem) {
// 	q.ele = append(q.ele, e)
// }
//
// func (q *Queue) Top() server.LogItem {
// 	return q.ele[0]
// }
//
// func (q *Queue) Pop() {
// 	q.ele = q.ele[1:]
// }
//
// func (q *Queue) Size() int {
// 	return len(q.ele)
// }
