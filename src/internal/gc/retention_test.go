package gc

import (
	"reflect"
	"testing"
)

func TestBuildNumbersToExpire(t *testing.T) {
	desc := []int64{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5}
	got := BuildNumbersToExpire(desc, 10)
	want := []int64{5} // only the 11th oldest by number
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if BuildNumbersToExpire([]int64{1, 2}, 10) != nil {
		t.Fatal("expected nil")
	}
}
