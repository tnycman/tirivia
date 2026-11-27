package seal

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
)

func key(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, KeyBytes)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	return k
}

func TestRoundTripAndRotation(t *testing.T) {
	k1, k2 := key(t), key(t)
	old, err := New(1, map[byte][]byte{1: k1})
	if err != nil {
		t.Fatal(err)
	}
	sealed := old.Seal([]byte("+355691234567"), []byte("row-1"))
	rotated, err := New(2, map[byte][]byte{1: k1, 2: k2})
	if err != nil {
		t.Fatal(err)
	}
	got, err := rotated.Open(sealed, []byte("row-1"))
	if err != nil || !bytes.Equal(got, []byte("+355691234567")) {
		t.Fatalf("open after rotation: %q %v", got, err)
	}
	if rotated.Seal([]byte("x"), nil)[0] != 2 {
		t.Fatal("new ciphertexts must use the current key")
	}
}

func TestRejects(t *testing.T) {
	s, err := New(1, map[byte][]byte{1: key(t)})
	if err != nil {
		t.Fatal(err)
	}
	sealed := s.Seal([]byte("secret"), []byte("row-1"))
	tampered := append([]byte{}, sealed...)
	tampered[len(tampered)-1] ^= 1
	cases := map[string][]byte{"wrong aad": sealed, "tampered": tampered, "unknown key": append([]byte{9}, sealed[1:]...), "empty": nil}
	for name, c := range cases {
		aad := []byte("row-1")
		if name == "wrong aad" {
			aad = []byte("row-2")
		}
		if _, err := s.Open(c, aad); !errors.Is(err, ErrSealed) {
			t.Errorf("%s: got %v", name, err)
		}
	}
	if _, err := New(1, map[byte][]byte{1: make([]byte, 16)}); !errors.Is(err, ErrKey) {
		t.Error("short key accepted")
	}
	if _, err := New(2, map[byte][]byte{1: key(t)}); !errors.Is(err, ErrKey) {
		t.Error("missing current key accepted")
	}
}
