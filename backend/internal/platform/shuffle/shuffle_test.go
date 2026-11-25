package shuffle

import (
	"fmt"
	"slices"
	"testing"
)

func TestOrderIsDeterministicPermutation(t *testing.T) {
	seed := []byte("player-seed-0001")
	a := Order(seed, "q-1", 3)
	b := Order(seed, "q-1", 3)
	if !slices.Equal(a, b) {
		t.Fatalf("not deterministic: %v vs %v", a, b)
	}
	sorted := slices.Clone(a)
	slices.Sort(sorted)
	if !slices.Equal(sorted, []int{0, 1, 2}) {
		t.Fatalf("not a permutation: %v", a)
	}
}

// All 6 orders must appear with roughly equal frequency across questions, otherwise
// "everyone tap the top button" coordination would gain an edge.
func TestOrderDistribution(t *testing.T) {
	const trials = 60_000
	counts := map[string]int{}
	seed := []byte("seed")
	for i := range trials {
		counts[fmt.Sprint(Order(seed, fmt.Sprintf("question-%d", i), 3))]++
	}
	if len(counts) != 6 {
		t.Fatalf("expected 6 distinct orders, got %d: %v", len(counts), counts)
	}
	for order, c := range counts {
		// Expected 10,000 each; modulo bias for n=3 is < 1%. Allow +-6%.
		if c < 9_400 || c > 10_600 {
			t.Errorf("order %s appeared %d times (expected ~10000)", order, c)
		}
	}
}
