package structopt

import (
	"testing"
)

func TestAtoX(t *testing.T) {
	t.Run("0", testInt("0", 0))
	t.Run("1", testInt("1", 1))
	t.Run("255", testInt("255", 255))
	t.Run("65535", testInt("65535", 65535))
	t.Run("4294967295", testInt("4294967295", 4294967295))
	t.Run("9223372036854775807", testInt("9223372036854775807", 9223372036854775807))
	t.Run("-9223372036854775808", testInt("-9223372036854775808", -9223372036854775808))

	t.Run("0", testUint("0", 0))
	t.Run("1", testUint("1", 1))
	t.Run("255", testUint("255", 255))
	t.Run("65535", testUint("65535", 65535))
	t.Run("4294967295", testUint("4294967295", 4294967295))
	t.Run("9223372036854775807", testUint("9223372036854775807", 9223372036854775807))
	t.Run("18446744073709551615", testUint("18446744073709551615", 18446744073709551615))
}

func testInt(s string, ans int64) func(*testing.T) {
	return func(t *testing.T) {
		val, err := AtoI(s)
		switch {
		case err != nil:
			t.Fatalf("cannot run AtoI(%v): %v", s, err)
		case ans != val:
			t.Errorf("AtoI(%v) = %v: %v", s, val, ans)
		}
	}
}

func testUint(s string, ans uint64) func(*testing.T) {
	return func(t *testing.T) {
		val, err := AtoU(s)
		switch {
		case err != nil:
			t.Fatalf("cannot run AtoI(%v): %v", s, err)
		case ans != val:
			t.Errorf("AtoI(%v) = %v: %v", s, val, ans)
		}
	}
}

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
