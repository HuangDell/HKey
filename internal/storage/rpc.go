package storage

// 定义storage中用于rpc的方法

const SetRPC = "HKey.Set"
const GetRPC = "HKey.Get"
const DelRPC = "HKey.Del"
const ExistsRPC = "HKey.Exists"

type SetArgs struct {
	Key, Value string
}

// Set 面向RPC的方法
func (this *HKey) Set(args SetArgs, ans *string) error {
	*ans = "ok"
	pos, err := this.find(args.Key)
	if err != nil {
		return err
	}
	if pos == -1 {
		err = this.insert(args.Key, args.Value)
		if err != nil {
			return err
		}
	} else {
		this.update(args.Key, args.Value, pos)
	}
	return nil
}

// Get 面向RPC的方法
func (this *HKey) Get(key string, ans *string) error {
	var err error
	*ans, err = this.get(key)
	if err != nil {
		return err
	}
	return nil
}

// Del 面向RPC方法
func (this *HKey) Del(key string, ans *string) error {
	var err error
	*ans, err = this.delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (this *HKey) Exists(key string, ans *string) error {
	var err error
	var pos int
	pos, err = this.find(key)
	if err != nil {
		return err
	}
	if pos == -1 {
		*ans = "(integer) 0"
	} else {
		*ans = "(integer) 1"
	}
	return nil
}

// Show 打印cache中的内容
func (this *HKey) Show(ans *string) {
	*ans = showCache(this.cache)
}
