package internal

import (
	"fmt"
	"strings"
	"testing"
)

type testID uint32

// assertPanicsWith fails unless f panics with a message containing want.
func assertPanicsWith(t *testing.T, what, want string, f func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("%s: expected a panic, got none", what)
			return
		}
		if got := fmt.Sprint(r); !strings.Contains(got, want) {
			t.Errorf("%s: panicked with %q, want a message containing %q", what, got, want)
		}
	}()
	f()
}

func TestSlotMapAllocLookup(t *testing.T) {
	var m SlotMap[int, testID]
	a, b := 1, 2

	idA := m.Alloc(&a)
	idB := m.Alloc(&b)

	if idA == idB {
		t.Fatalf("distinct allocations shared a handle: %d", idA)
	}
	if got := m.Lookup(idA); got != &a {
		t.Errorf("Lookup(idA) = %p, want %p", got, &a)
	}
	if got := m.Lookup(idB); got != &b {
		t.Errorf("Lookup(idB) = %p, want %p", got, &b)
	}
}

func TestSlotMapFreeReturnsValueAndReusesSlot(t *testing.T) {
	var m SlotMap[int, testID]
	a, b := 1, 2

	idA := m.Alloc(&a)
	if got := m.Free(idA); got != &a {
		t.Fatalf("Free returned %p, want %p", got, &a)
	}

	idB := m.Alloc(&b)
	idxA, _ := splitID(idA)
	idxB, _ := splitID(idB)
	if idxA != idxB {
		t.Errorf("expected the freed slot (index %d) to be reused, got index %d", idxA, idxB)
	}
	if idA == idB {
		t.Errorf("reused slot handed out an identical handle %d; generation did not advance", idA)
	}
	if got := m.Lookup(idB); got != &b {
		t.Errorf("Lookup(idB) = %p, want %p", got, &b)
	}
	if len(m.slots) != 1 {
		t.Errorf("len(slots) = %d, want 1 (slot should be recycled, not appended)", len(m.slots))
	}
}

func TestSlotMapStaleHandlePanics(t *testing.T) {
	var m SlotMap[int, testID]
	a, b := 1, 2

	idA := m.Alloc(&a)
	m.Free(idA)
	m.Alloc(&b) // recycles the same slot with a bumped generation

	assertPanicsWith(t, "Lookup of a freed handle", "stale handle", func() { m.Lookup(idA) })
	assertPanicsWith(t, "Free of a freed handle", "stale handle", func() { m.Free(idA) })
}

func TestSlotMapDoubleFreePanics(t *testing.T) {
	var m SlotMap[int, testID]
	a := 1

	id := m.Alloc(&a)
	m.Free(id)

	assertPanicsWith(t, "second Free", "stale handle", func() { m.Free(id) })
}

func TestSlotMapOutOfBoundsHandlePanics(t *testing.T) {
	var m SlotMap[int, testID]
	a := 1
	m.Alloc(&a)

	beyond := makeID[testID](99, 1)
	assertPanicsWith(t, "Lookup past the end", "out of bounds", func() { m.Lookup(beyond) })
	assertPanicsWith(t, "Free past the end", "out of bounds", func() { m.Free(beyond) })
}

// The zero handle is reserved (InvalidHandle, re-exported as backend.InvalidBuffer).
// It must report that specific cause rather than being mistaken for an out-of-range
// or stale handle, because the likeliest way to reach it is an uninitialised
// component whose handle field is still the zero value.
func TestSlotMapZeroHandleReportsItsOwnCause(t *testing.T) {
	const want = "reserved zero handle"

	var empty SlotMap[int, testID]
	assertPanicsWith(t, "Lookup(0) on an empty map", want, func() { empty.Lookup(0) })
	assertPanicsWith(t, "Free(0) on an empty map", want, func() { empty.Free(0) })

	var filled SlotMap[int, testID]
	a := 1
	filled.Alloc(&a)
	assertPanicsWith(t, "Lookup(0) on a populated map", want, func() { filled.Lookup(0) })
	assertPanicsWith(t, "Free(0) on a populated map", want, func() { filled.Free(0) })
}

// Alloc must never hand back the reserved value, or a live resource would be
// indistinguishable from an unset handle.
func TestSlotMapAllocNeverReturnsInvalidHandle(t *testing.T) {
	if InvalidHandle != 0 {
		t.Fatalf("InvalidHandle = %d, want 0 (the zero value of every handle type)", InvalidHandle)
	}

	var m SlotMap[int, testID]
	v := 1
	for i := 0; i < 100; i++ {
		if id := m.Alloc(&v); id == InvalidHandle {
			t.Fatalf("allocation %d returned the reserved zero handle", i)
		}
	}
}

// Regression test: the generation field is only genBits wide inside the packed
// handle. If Free lets slot.gen grow past genMask, makeID truncates it and the
// slot starts handing out handles that Lookup immediately rejects.
func TestSlotMapGenerationWrapKeepsHandlesValid(t *testing.T) {
	var m SlotMap[int, testID]
	v := 7

	cycles := (1 << genBits) + 100 // comfortably past one full wrap
	for i := 0; i < cycles; i++ {
		id := m.Alloc(&v)

		if _, gen := splitID(id); gen == 0 {
			t.Fatalf("cycle %d: handed out a handle with generation 0, which is reserved", i)
		}
		if got := m.Lookup(id); got != &v {
			t.Fatalf("cycle %d: Lookup(%d) = %p, want %p", i, id, got, &v)
		}
		if m.Free(id) != &v {
			t.Fatalf("cycle %d: Free returned the wrong value", i)
		}
	}

	if len(m.slots) != 1 {
		t.Errorf("len(slots) = %d, want 1; the slot should have been recycled every cycle", len(m.slots))
	}
}

func TestSlotMapIDRoundTrip(t *testing.T) {
	cases := []struct{ idx, gen uint32 }{
		{0, 1},
		{1, 1},
		{idxMask, 1},
		{0, genMask},
		{idxMask, genMask},
		{12345, 678},
	}
	for _, c := range cases {
		idx, gen := splitID(makeID[testID](c.idx, c.gen))
		if idx != c.idx || gen != c.gen {
			t.Errorf("round trip of (idx=%d, gen=%d) gave (idx=%d, gen=%d)", c.idx, c.gen, idx, gen)
		}
	}
}

// Documents a known limitation: distinct handle TYPES cannot be mixed (that is a
// compile error), but two SlotMaps sharing one handle type share a numbering
// space, and a handle from one silently resolves inside the other. The engine
// keeps a single map per resource type, so this stays unreachable — give each
// resource its own defined ID type and it remains so.
func TestSlotMapHandlesAreNotValidatedAcrossMapsOfTheSameType(t *testing.T) {
	var a, b SlotMap[int, testID]
	x, y := 1, 2

	a.Alloc(&x)
	idB := b.Alloc(&y)

	if got := a.Lookup(idB); got != &x {
		t.Errorf("Lookup = %p, want %p: a cross-map handle resolves against the local slot", got, &x)
	}
}
