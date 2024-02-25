package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const SAMPLE_HASH = "9c151c3af838278e3ef57c180c7d031c07aefd12f2ccc1e18f2a1e1c7d0ff163"

func TestIncorrectHexLength(t *testing.T) {
	_, err := Hash256FromHexString("AAA")
	assert.ErrorContains(t, err, "incorrect hash length")
}

func TestIncorrectHexFormat(t *testing.T) {
	_, err := Hash256FromHexString("9c151c3af838278e3ef57c180c7d031c07aefd12f2ccc1e18f2a1e1c7d0ff16!")
	assert.ErrorContains(t, err, "incorrect format")
}

func TestCorrectHexFormat(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	assert.ErrorIs(t, err, nil)

	assert.Equal(t, SAMPLE_HASH, hash.String())
}

func TestClone(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	assert.ErrorIs(t, err, nil)

	clone := hash.Clone()
	assert.Equal(t, SAMPLE_HASH, clone.String())

	assert.True(t, hash.Eq(&clone))
}

func TestToString(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	assert.ErrorIs(t, err, nil)

	assert.Equal(t, SAMPLE_HASH, hash.String())
}

func TestBitCount(t *testing.T) {
	assert.Equal(t, 1, BitCount(1))

	// dec(10) = bin(01100100)
	assert.Equal(t, 3, BitCount(100))
}

func TestHammingNorm(t *testing.T) {
	hash := &Hash256{}
	hash.SetAll()

	assert.Equal(t, 256, hash.HammingNorm())

	hash, err := Hash256FromHexString(SAMPLE_HASH)
	assert.ErrorIs(t, err, nil)
	assert.Equal(t, 128, hash.HammingNorm())
}

func TestHammingDistance(t *testing.T) {
	hash1, err := Hash256FromHexString(SAMPLE_HASH)
	assert.ErrorIs(t, err, nil)

	hash2 := Hash256{}
	hash2.ClearAll()

	assert.Equal(t, 128, hash1.HammingDistance(&hash2))

	hash1 = &Hash256{}
	hash1.SetAll()
	hash2 = Hash256{}
	hash2.ClearAll()

	assert.Equal(t, 256, hash1.HammingDistance(&hash2))
	assert.False(t, hash1.HammingDistanceLE(&hash2, 1))
	assert.True(t, hash1.HammingDistanceLE(&hash2, 257))
	assert.True(t, hash1.HammingDistanceLE(hash1, 0))
}

func TestBinaryOperations(t *testing.T) {
	hash, err := Hash256FromHexString(SAMPLE_HASH)
	assert.ErrorIs(t, err, nil)

	result := hash.BitwiseAND(hash)
	hash2 := &result
	assert.True(t, hash2.Eq(hash))

	hashNegative := hash.BitwiseNOT()
	result = hash.BitwiseAND(&hashNegative)
	hash2 = &result
	hash3 := &Hash256{}
	assert.True(t, hash2.Eq(hash3))

	hash_set_all := &Hash256{}
	hash_set_all.SetAll()

	result = hash.BitwiseOR(&hashNegative)
	assert.True(t, result.Eq(hash_set_all))

	result = hash.BitwiseXOR(&hashNegative)
	assert.True(t, result.Eq(hash_set_all))
}
