package storage

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// 设定数据内容存储路径
const data_path = "build/server/data/"

// handler 用于数据的持久性保存
type handler struct {
	deletedPos int
	fp         *os.File
}

func newHandler() *handler {
	var err error
	h := new(handler)
	h.fp, err = os.OpenFile(data_path+"1.data", os.O_CREATE|os.O_RDWR, os.ModePerm)
	h.deletedPos = -1
	if err != nil {
		panic(err)
	}
	return h
}

// 用于关闭handler
func (this *handler) close() {
	this.fp.Close()
}

func (this *handler) write(data *HKey) {
	var len = make([]byte, INT_SIZE)
	if this.deletedPos != -1 {
		var offset int
		offset = int(DATA_LEN) * this.deletedPos
		this.fp.Seek(int64(offset), io.SeekStart)
		this.deletedPos = -1
	} else {
		this.fp.Seek(0, io.SeekEnd)
	}
	binary.LittleEndian.PutUint32(len, uint32(data.value.len))
	this.fp.Write(data.key)
	this.fp.Write(data.deleted)
	this.fp.Write(data.hashVal)
	this.fp.Write(len)
	this.fp.Write(data.value.buf)
}

// 从文件中通过Key读取Value
func (this *handler) read(pos int) (string, error) {
	if pos == -1 {
		return "nil", nil
	}
	block := make([]byte, DATA_LEN)
	_, err := this.fp.ReadAt(block, int64(pos*int(DATA_LEN)))
	if err != nil {
		return "", fmt.Errorf("读取数据异常%v", err)
	}
	return string(block[DATA_LEN-VALUE_MAX_LEN:]), nil
}

// 查找key
func (this *handler) find(data *HKey) (int, error) {
	var pos int
	this.fp.Seek(0, io.SeekStart)
	block := make([]byte, DATA_LEN)
	for {
		_, err := this.fp.Read(block)
		if err == io.EOF {
			return -1, nil
		}
		if err != nil {
			return -1, fmt.Errorf("查找数据异常%v", err)
		}
		if string(data.key) == string(block[:KEY_MAK_LEN]) && block[KEY_MAK_LEN] == 0 {
			// fmt.Printf("Find %s at %d\n", string(data.key), pos)
			return pos, nil
		}
		if block[KEY_MAK_LEN] == 1 {
			this.deletedPos = pos
		}
		pos++
	}
}

func (this *handler) update(data *HKey, pos int) {
	offset := pos*int(DATA_LEN) + KEY_MAK_LEN + DELETED_LEN
	var len = make([]byte, INT_SIZE)
	binary.LittleEndian.PutUint32(len, uint32(data.value.len))
	this.fp.Seek(int64(offset), io.SeekStart) // TODO 这些error如何处理
	this.fp.Write(data.hashVal)
	this.fp.Write(len)
	this.fp.Write(data.value.buf)
}

func (this *handler) delete(pos int) error {
	d := []byte{1}
	offset := pos*int(DATA_LEN) + KEY_MAK_LEN
	_, err := this.fp.WriteAt(d, int64(offset))
	if err != nil {
		return fmt.Errorf("删除数据错误%v", err)
	}
	this.deletedPos = pos
	return nil
}
