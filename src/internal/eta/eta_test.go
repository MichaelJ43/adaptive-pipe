package eta

import "testing"

func TestSimpleMovingAverageMs(t *testing.T) {
	avg, ok := SimpleMovingAverageMs([]int64{1000, 2000, 3000})
	if !ok || avg != 2000 {
		t.Fatalf("got %v %v", avg, ok)
	}
	_, ok = SimpleMovingAverageMs(nil)
	if ok {
		t.Fatal("expected empty")
	}
}
