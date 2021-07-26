package structopt

import (
	"testing"
)

func TestWidecharSize(t *testing.T) {
	t.Run("test", testWidecharSize("test", 4))
	t.Run("測試", testWidecharSize("測試", 4))
	t.Run("test 測試 ทดสอบ", testWidecharSize("test測試ทดสอบ", 13))
}

func testWidecharSize(s string, size int) func(*testing.T) {
	return func(t *testing.T) {
		if w_size := WidecharSize(s); w_size != size {
			// not match the expect wide-char size
			t.Errorf("expect %#v size is %v: %v", s, size, w_size)
		}
	}
}
