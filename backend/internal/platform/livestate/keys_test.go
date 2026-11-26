package livestate

import (
	"strconv"
	"strings"
	"testing"
)

// hashTag returns what Valkey hashes: the text between the first "{" and the next "}" (cluster-spec "Hash tags").
func hashTag(key string) string {
	open := strings.IndexByte(key, '{')
	if open < 0 {
		return key
	}
	end := strings.IndexByte(key[open+1:], '}')
	if end <= 0 {
		return key
	}
	return key[open+1 : open+1+end]
}

func TestScriptKeysShareOneSlot(t *testing.T) {
	for _, user := range []string{"alice", "user-1", "çelës", "a}b{c"} {
		b := Bucket(user)
		tag := hashTag(PlayerKey("show7", user))
		for _, k := range []string{TallyKey("show7", 3, b), CountedKey("show7", 3, b)} {
			if hashTag(k) != tag {
				t.Errorf("user %q: %s and %s hash differently", user, PlayerKey("show7", user), k)
			}
		}
	}
}

func TestBucketsSpread(t *testing.T) {
	seen := map[int]bool{}
	for i := range 10_000 {
		seen[Bucket("user-"+strconv.Itoa(i))] = true
	}
	if len(seen) != TallyShards {
		t.Fatalf("only %d of %d buckets used", len(seen), TallyShards)
	}
}
