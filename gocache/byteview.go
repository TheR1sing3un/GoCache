package gocache

//真实存储缓存数据的数据结构(只读,不可被外部修改)
type ByteView struct {
	b []byte
}

//获取缓存数据的内存大小
func (v ByteView) Len() int {
	return len(v.b)
}

//获取一个该缓存数据的拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneByte(v.b)
}

//将数据转化为字符串
func (v ByteView) String() string {
	return string(v.b)
}

//克隆数据
func cloneByte(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
