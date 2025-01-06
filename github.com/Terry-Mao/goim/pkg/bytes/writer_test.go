package bytes

import (
	"fmt"
	"testing"
)

func TestWriter(t *testing.T) {
	w := NewWriterSize(64)
	if w.Len() != 0 && w.Size() != 64 {
		t.FailNow()
	}
	b := []byte("hello")
	w.Write(b)
	//if !reflect.DeepEqual(b, w.Buffer()) {
	//	t.FailNow()
	//}
	fmt.Println(w.Len())
	fmt.Println(string(w.Buffer()))
	//w.Peek(len(b))
	//w.Reset()
	//for i := 0; i < 1024; i++ {
	//	w.Write(b)
	//}
	//w.Reset()
	//if w.Len() != 0 {
	//	t.FailNow()
	//}
}
