package utils

import (
	"bytes"
	"math/rand"
	"sync"
	"time"
)

var Rand = newRand()

type randUtil struct {
	rand *rand.Rand
	mx   sync.Mutex
}

func newRand() *randUtil {
	return &randUtil{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		mx:   sync.Mutex{},
	}
}

// 随机 [0, max-1] 之间的随机数
func (u *randUtil) Rand(max int64) int64 {
	return u.RandStart(0, max)
}

// 随机返回 [start, end-1] 之间的随机数
func (u *randUtil) RandStart(start, end int64) int64 {
	if end <= start {
		return 0
	}

	u.mx.Lock()
	v := u.rand.Int63n(end - start)
	u.mx.Unlock()
	return v + start
}

// 随机 [0, max-1] count 次并把执行索引和随机的数值传给 fn 函数执行
func (u *randUtil) RandF(max, count int64, fn func(i, v int64)) {
	u.RandStartF(0, max, count, fn)
}

// 随机 [start, end-1] count 次并把执行索引和随机的数值传给 fn 函数执行
func (u *randUtil) RandStartF(start, end, count int64, fn func(i, v int64)) {
	if end <= start || count == 0 {
		return
	}

	max := end - start
	for i := int64(0); i < count; i++ {
		u.mx.Lock()
		v := u.rand.Int63n(max)
		u.mx.Unlock()
		fn(i, v+start)
	}
}

// 随机指定长度的文本, 随机字符串候选词来自base
func (u *randUtil) RandText(count int, base string) string {
	if count <= 0 {
		return ""
	}

	tr := []rune(base)
	l := len(tr)

	var buf = &bytes.Buffer{}

	u.mx.Lock()
	for i := 0; i < count; i++ {
		buf.WriteRune(tr[u.rand.Intn(l)])
	}
	u.mx.Unlock()
	return buf.String()
}
