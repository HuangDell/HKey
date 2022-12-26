package storage

import (
	"HKey/pkg"
	"crypto/md5"
	"fmt"
	"unsafe"
)

var INT_TEST int

const INT_SIZE = unsafe.Sizeof(INT_TEST)
const VALUE_MAX_LEN = 64 // 设定value的最大长度为64字节
const KEY_MAK_LEN = 16   // 设定key得最大长度为16字节
const DELETED_LEN = 1    // 标识字段
const DATA_LEN = INT_SIZE + KEY_MAK_LEN + VALUE_MAX_LEN + 16 + DELETED_LEN

type sbshdr struct {
	len int
	buf []byte
}

type HKey struct {
	deleted []byte
	key     []byte
	hashVal []byte
	value   sbshdr
	store   *handler
}

func NewHKey(data_path string) *HKey {
	hkey := new(HKey)
	hkey.store = newHandler(data_path)
	hkey.deleted = make([]byte, 1)
	hkey.hashVal = make([]byte, 16)
	hkey.key = make([]byte, KEY_MAK_LEN)
	hkey.value.buf = make([]byte, VALUE_MAX_LEN)
	return hkey
}

func (this *HKey) insert(key string, value string) error {
	if len(key) > KEY_MAK_LEN || len(value) > VALUE_MAX_LEN {
		return fmt.Errorf("键或值超过字节")
	}
	temp := md5.Sum([]byte(value))
	copy(this.hashVal, temp[:])
	copy(this.key, key)
	copy(this.value.buf, value)
	this.store.write(this)
	pkg.Clear(this.hashVal) // TODO 也许可以不用clear
	pkg.Clear(this.key)
	pkg.Clear(this.value.buf)
	return nil
}

func (this *HKey) get(key string) (string, error) {
	copy(this.key, key)
	defer pkg.Clear(this.key)
	pos, err := this.store.find(this)
	if err != nil {
		return "", err
	}
	return this.store.read(pos)
}

// update
func (this *HKey) update(key string, value string, pos int) {
	temp := md5.Sum([]byte(value))
	copy(this.hashVal, temp[:])
	copy(this.key, key)
	copy(this.value.buf, value)
	this.store.update(this, pos)
	pkg.Clear(this.hashVal) // TODO 也许可以不用clear
	pkg.Clear(this.key)
	pkg.Clear(this.value.buf)
}

func (this *HKey) find(key string) (int, error) {
	copy(this.key, key)
	defer pkg.Clear(this.key)
	return this.store.find(this)
}

func (this *HKey) delete(key string) (string, error) {
	copy(this.key, key)
	pos, err := this.store.find(this)
	if err != nil {
		return "", err
	}
	if pos == -1 {
		return "(integer) 0", nil
	}
	err = this.store.delete(pos)
	if err != nil {
		return "", err
	}
	return "(integer) 1", nil
}
