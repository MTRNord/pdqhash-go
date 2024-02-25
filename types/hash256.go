package types

import (
	"fmt"
	"strconv"
	"strings"
)

// 16 slots of 16 bytes each
// See hashing/pdq/README-MIH.md in upstream repo for why not 8x32 or 32x8, etc.
const HASH256_NUM_SLOTS = 16
const HASH256_HEX_NUM_NYBBLES = 4 * HASH256_NUM_SLOTS

type Hash256 struct {
	W [HASH256_NUM_SLOTS]int
}

func (h *Hash256) GetNumWords() int {
	return HASH256_NUM_SLOTS
}

func (h *Hash256) Clone() Hash256 {
	rv := Hash256{}
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		rv.W[i] = h.W[i]
	}
	return rv
}

func (h *Hash256) String() string {
	var result []string

	for i := HASH256_NUM_SLOTS - 1; i >= 0; i-- {
		result = append(result, fmt.Sprintf("%04x", h.W[i]&0xFFFF))
	}

	return strings.Join(result, "")
}

func (h *Hash256) ToHexString() string {
	return h.String()
}

func Hash256FromHexString(s string) (*Hash256, error) {
	if len(s) != HASH256_HEX_NUM_NYBBLES {
		return nil, fmt.Errorf("incorrect hash length: %s", s)
	}

	rv := &Hash256{}
	i := HASH256_NUM_SLOTS
	for x := 0; x < len(s); x += 4 {
		i -= 1
		val, err := strconv.ParseUint(s[x:x+4], 16, 16)
		if err != nil {
			return nil, fmt.Errorf("incorrect format: %s", s)
		}
		rv.W[i] = int(val)
	}
	return rv, nil
}

func (h *Hash256) HammingNorm16(h2 int) int {
	return BitCount(h2 & 0xFFFF)
}

func BitCount(x int) int {
	x -= (x >> 1) & 0x55555555
	x = ((x >> 2) & 0x33333333) + (x & 0x33333333)
	x = ((x >> 4) + x) & 0x0F0F0F0F
	x += x >> 8
	x += x >> 16
	return x & 0x0000003F
}

func (h *Hash256) ClearAll() {
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		h.W[i] = 0
	}
}

func (h *Hash256) SetAll() {
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		h.W[i] = 0xFFFF
	}
}

func (h *Hash256) HammingNorm() int {
	n := 0

	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		n += h.HammingNorm16(h.W[i])
	}
	return n
}

func (h *Hash256) HammingDistance(that *Hash256) int {
	n := 0
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		n += h.HammingNorm16(h.W[i] ^ that.W[i])
	}
	return n
}

func (h *Hash256) HammingDistanceLE(that *Hash256, d int) bool {
	e := 0
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		e += h.HammingNorm16(h.W[i] ^ that.W[i])
		if e > d {
			return false
		}
	}
	return true
}

func (h *Hash256) SetBit(k int) {
	h.W[(k&255)>>4] |= 1 << (k & 15)
}

func (h *Hash256) FlipBit(k int) {
	h.W[(k&255)>>4] ^= 1 << (k & 15)
}

func (h *Hash256) BitwiseXOR(that *Hash256) Hash256 {
	rv := Hash256{}
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		rv.W[i] = (h.W[i] ^ that.W[i])
	}
	return rv
}

func (h *Hash256) BitwiseAND(that *Hash256) Hash256 {
	rv := Hash256{}
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		rv.W[i] = (h.W[i] & that.W[i])
	}
	return rv
}

func (h *Hash256) BitwiseOR(that *Hash256) Hash256 {
	rv := Hash256{}
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		rv.W[i] = (h.W[i] | that.W[i])
	}
	return rv
}

func (h *Hash256) BitwiseNOT() Hash256 {
	rv := Hash256{}
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		rv.W[i] = (int((^h.W[i])) & 0xFFFF)
	}
	return rv
}

func (h *Hash256) DumpBits() string {
	var str []string

	for i := HASH256_NUM_SLOTS - 1; i >= 0; i-- {
		word := h.W[i] & 0xFFFF
		var bits []string
		for j := 15; j >= 0; j-- {
			if (word & (1 << uint(j))) != 0 {
				bits = append(bits, "1")
			} else {
				bits = append(bits, "0")
			}
		}
		str = append(str, strings.Join(bits, " "))
	}
	return strings.Join(str, "\n")
}

func (h *Hash256) DumpBitsAcross() string {
	var str []string

	for i := HASH256_NUM_SLOTS - 1; i >= 0; i-- {
		word := h.W[i] & 0xFFFF
		for j := 15; j >= 0; j-- {
			if (word & (1 << uint(j))) != 0 {
				str = append(str, "1")
			} else {
				str = append(str, "0")
			}
		}
	}
	return strings.Join(str, " ")
}

func (h *Hash256) DumpWords() string {
	var words []string

	// Iterate over the reversed list of words
	for i := len(h.W) - 1; i >= 0; i-- {
		words = append(words, strconv.Itoa(int(h.W[i])))
	}

	// Join the words with commas
	return strings.Join(words, ",")
}

func (h *Hash256) Eq(other *Hash256) bool {
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		if h.W[i] != other.W[i] {
			return false
		}
	}
	return true
}

func (h *Hash256) Greater(other *Hash256) bool {
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		if h.W[i] > other.W[i] {
			return true
		} else if h.W[i] < other.W[i] {
			return false
		}
	}
	return false
}

func (h *Hash256) Less(other *Hash256) bool {
	for i := 0; i < HASH256_NUM_SLOTS; i++ {
		if h.W[i] < other.W[i] {
			return true
		} else if h.W[i] > other.W[i] {
			return false
		}
	}
	return false
}
