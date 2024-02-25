package types

import (
	"strings"
	"testing"
)

const SAMPLE_HASH = "9c151c3af838278e3ef57c180c7d031c07aefd12f2ccc1e18f2a1e1c7d0ff163"

func TestIncorrectHexLength(t *testing.T) {
	_, err := Hash256FromHexString("AAA")
	if !strings.HasPrefix(err.Error(), "incorrect hash length") {
		t.Errorf("Incorrect error message: %s", err)
	}
}

func TestIncorrectHexFormat(t *testing.T) {
	_, err := Hash256FromHexString("9c151c3af838278e3ef57c180c7d031c07aefd12f2ccc1e18f2a1e1c7d0ff16!")
	if !strings.HasPrefix(err.Error(), "incorrect format") {
		t.Errorf("Incorrect error message: %s", err)
	}
}

func TestCorrectHexFormat(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if hash.String() != SAMPLE_HASH {
		t.Errorf("Incorrect hash: %s", hash.String())
	}
}

func TestClone(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	clone := hash.Clone()
	if clone.String() != SAMPLE_HASH {
		t.Errorf("Incorrect hash: %s", clone.String())
	}
	if !hash.Eq(&clone) {
		t.Errorf("Incorrect equality")
	}
}

func TestToString(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if hash.String() != SAMPLE_HASH {
		t.Errorf("Incorrect hash: %s", hash.String())
	}
}

func TestBitCount(t *testing.T) {
	if BitCount(1) != 1 {
		t.Errorf("Incorrect bit count")
	}

	// dec(10) = bin(01100100)
	if BitCount(100) != 3 {
		t.Errorf("Incorrect bit count")
	}
}

func TestHammingNorm(t *testing.T) {
	hash := &Hash256{}
	hash.SetAll()
	if hash.HammingNorm() != 256 {
		t.Errorf("Incorrect hamming norm")
	}

	hash, err := Hash256FromHexString(SAMPLE_HASH)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if hash.HammingNorm() != 128 {
		t.Errorf("Incorrect hamming norm")
	}
}

func TestHammingDistance(t *testing.T) {
	hash1, err := Hash256FromHexString(SAMPLE_HASH)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	hash2 := Hash256{}
	hash2.ClearAll()

	if hash1.HammingDistance(&hash2) != 128 {
		t.Errorf("Incorrect hamming distance")
	}

	hash1 = &Hash256{}
	hash1.SetAll()
	hash2 = Hash256{}
	hash2.ClearAll()

	if hash1.HammingDistance(&hash2) != 256 {
		t.Errorf("Incorrect hamming distance")
	}
	if hash1.HammingDistanceLE(&hash2, 1) {
		t.Errorf("Incorrect hamming distance")
	}
	if !hash1.HammingDistanceLE(&hash2, 257) {
		t.Errorf("Incorrect hamming distance")
	}
	if !hash1.HammingDistanceLE(hash1, 0) {
		t.Errorf("Incorrect hamming distance")
	}
}

func TestBinaryOperations(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	result := hash.BitwiseAND(hash)
	hash2 := &result
	if !hash2.Eq(hash) {
		t.Errorf("Incorrect AND")
	}

	hashNegative := hash.BitwiseNOT()
	result = hash.BitwiseAND(&hashNegative)
	hash2 = &result
	hash3 := &Hash256{}
	if !hash2.Eq(hash3) {
		t.Errorf("Incorrect NOT")
	}

	hash_set_all := &Hash256{}
	hash_set_all.SetAll()

	result = hash.BitwiseOR(&hashNegative)
	if !result.Eq(hash_set_all) {
		t.Errorf("Incorrect OR with SET ALL")
	}

	result = hash.BitwiseXOR(&hashNegative)
	if !result.Eq(hash_set_all) {
		t.Errorf("Incorrect XOR with SET ALL")
	}
}
