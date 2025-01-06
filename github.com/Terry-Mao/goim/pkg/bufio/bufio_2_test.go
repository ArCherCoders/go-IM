package bufio

import (
	"fmt"
	"strings"
	"testing"
)

func Test_bufio(t *testing.T) {
	data := "hello world"
	reader := strings.NewReader(data)
	b := NewReader(reader)
	peek, err := b.Peek(12)
	fmt.Printf("peek:%v\n,err:%v\n", peek, err)
}
