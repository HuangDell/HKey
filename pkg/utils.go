package pkg

// 将数组置为0
func Clear(arr []byte) {
	for i := 0; i < len(arr); i++ {
		arr[i] = 0
	}
}
