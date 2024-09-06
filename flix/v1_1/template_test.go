package v1_1

import (
	"bytes"
	"testing"
)

func TestEqualCadenceAstHashes(t *testing.T) {
	hash1, err := cadenceAstHash([]byte(`
		transaction(a: String, b: Int64) {}
	`))
	if err != nil {
		t.Errorf("hash1 error: %s", err)
	}

	hash2, err := cadenceAstHash([]byte(`
		transaction	(
			a: String, 
			b: Int64
		) {
			// This comment should be ignored
		}
	`))
	if err != nil {
		t.Errorf("hash2 error: %s", err)
	}

	if !bytes.Equal(hash1, hash2) {
		t.Errorf("hashes not equal: %s != %s", hash1, hash2)
	}
}

func TestNonEqualCadenceAstHashes(t *testing.T) {
	hash1, err := cadenceAstHash([]byte(`
		transaction(a: String, b: Int64) {}
	`))
	if err != nil {
		t.Errorf("hash1 error: %s", err)
	}

	hash2, err := cadenceAstHash([]byte(`
		transaction(b: Int64, a: String) {}
	`))
	if err != nil {
		t.Errorf("hash2 error: %s", err)
	}

	if bytes.Equal(hash1, hash2) {
		t.Errorf("hashes equal: %s == %s", hash1, hash2)
	}
}
