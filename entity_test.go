package engine

import "testing"

type TestComponent struct {
	testInt int
	testStr string
}

func TestComponentArrayDelete(t *testing.T) {
	var e1 EntityID = 0
	var e2 EntityID = 1
	var e3 EntityID = 2
	ca := NewComponentArray[TestComponent]()
	ca.Set(e1, TestComponent{testInt: 1, testStr: "one"})
	ca.Set(e2, TestComponent{testInt: 2, testStr: "two"})
	ca.Set(e3, TestComponent{testInt: 3, testStr: "three"})
	t.Error(423)
	for i, cpnt := range ca.arr {
		t.Logf("arridx %d / %d %s", i, cpnt.testInt, cpnt.testStr)
	}
}
