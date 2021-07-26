package structopt

import (
	"testing"
)

func TestSeparateTags(t *testing.T) {
	t.Run("ignore-1", testSeparateTags(`-`, "-", ""))
	t.Run("ignore-2", testSeparateTags(`- name:"NAME"`, "-", ""))

	t.Run("name-empty", testSeparateTags(`name`, "name", ""))
	t.Run("name-value", testSeparateTags(`name:"NAME"`, "name", "NAME"))
	t.Run("name-value-with-space", testSeparateTags(`name:"N A M E"`, "name", "N A M E"))
	t.Run("name-wide-char", testSeparateTags(`name:"name 名字 ดังนั้น"`, "name", "name 名字 ดังนั้น"))

	t.Run("multi-column-x", testSeparateTags(`x:"1" y:"2" z`, "x", "1"))
	t.Run("multi-column-y", testSeparateTags(`x:"1" y:"2" z`, "y", "2"))
	t.Run("multi-column-z", testSeparateTags(`x:"1" y:"2" z`, "z", ""))
}

func testSeparateTags(tag string, key, value string) func(*testing.T) {
	return func(t *testing.T) {
		tags := sep_tags(tag)
		if v, ok := tags[key]; ok == false || v != value {
			// cannot get the value
			t.Errorf("expect get %v: %v (%v)", value, v, ok)
		}
	}
}
