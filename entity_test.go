package engine

import (
	"testing"
)

type TestComponent struct {
	testInt int
	testStr string
}

func TestComponentArrayAdd(t *testing.T) {
	var e0 EntityID = 0
	ca := NewComponentArray[TestComponent]()
	cpnt0 := TestComponent{testInt: 0, testStr: "zero"}
	ca.Upsert(e0, cpnt0)
	ca.Upsert(e0, cpnt0)
}

func TestComponentArrayGet(t *testing.T) {
	var e0 EntityID = 0
	ca := NewComponentArray[TestComponent]()
	cpnt, err := ca.Get(e0)
	if cpnt != nil || err == nil {
		t.Error("Component array should return an error when getting from empty array")
	}
}

func TestComponentArrayDelete(t *testing.T) {
	var e0 EntityID = 0
	var e1 EntityID = 1
	var e2 EntityID = 2
	ca := NewComponentArray[TestComponent]()
	cpnt0 := TestComponent{testInt: 0, testStr: "zero"}
	cpnt1 := TestComponent{testInt: 1, testStr: "one"}
	cpnt2 := TestComponent{testInt: 2, testStr: "two"}
	ca.Upsert(e0, cpnt0)
	ca.Upsert(e1, cpnt1)
	ca.Upsert(e2, cpnt2)
	ca.Remove(e0)
	if len(ca.component2EntityMap) != 2 {
		t.Error("component2EntityMap has bad length")
	}
	if len(ca.entity2ComponentMap) != 2 {
		t.Error("entity2ComponentMap has bad length")
	}
	retCpnt2, err := ca.Get(e2)
	if err != nil {
		t.Error("Got error instead of component 2")
	}
	if cpnt2.testInt != retCpnt2.testInt || cpnt2.testStr != retCpnt2.testStr {
		t.Errorf("Got wrong component2 %v", retCpnt2)
	}
}
