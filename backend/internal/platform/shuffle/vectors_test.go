package shuffle

import (
	"encoding/hex"
	"slices"
	"testing"

	tv "github.com/tirivia/backend/internal/platform/testvectors"
)

func TestSharedShuffleVectors(t *testing.T) {
	var suite tv.ShuffleSuite
	if err := tv.Load(tv.ShuffleFile, &suite); err != nil {
		t.Fatal(err)
	}
	if len(suite.Vectors) < 16 {
		t.Fatalf("only %d shuffle vectors", len(suite.Vectors))
	}
	for _, v := range suite.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			seed, err := hex.DecodeString(v.SeedHex)
			if err != nil {
				t.Fatal(err)
			}
			if got := Order(seed, v.QuestionID, v.N); !slices.Equal(got, v.Order) {
				t.Fatalf("got %v, want %v", got, v.Order)
			}
		})
	}
}
