package pdq

import (
	"log"
	"testing"

	"github.com/MTRNord/pdqhash-go/types"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

type Pair struct {
	First  string
	Second string
}

var DATA_ARRAY = [19]Pair{
	{
		"./test-images/misc-images/b.jpg",
		"d8f8f0cce0f4a84f0e370a22028f67f0b36e2ed596623e1d33e6b39c4e9c9b22",
	},
	{
		"./test-images/misc-images/c.png",
		"e64cc9d91e623842f8d1f1d9a398e78c9f199a3bd87924f2b7e11e0bf061b064",
	},
	{
		"./test-images/misc-images/small.jpg",
		"0007001f003f003f007f00ff00ff00ff01ff01ff01ff03ff03ff03ff03ff03ff",
	},
	{
		"./test-images/misc-images/wee.jpg",
		"6227401f601ff4ccafcc9fad4b0d95d371a2eb7265a3285234d228ca94deeb2d",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q0003.jpg",
		"54a977c221d14c1c43ba5e6e21d4a13989a3553f1462611cbb85fda7be83b677",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q0004.jpg",
		"992d44af36d69e6ca6b812585928bac11def254ef5398c6d07466c9abcc65b92",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q0122.jpg",
		"cfb2009ddd21c6dab0046a7745b5984757a8a4535b3377aea2591d32b33ff940",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q0291.jpg",
		"a0fe94f1e5cc1cc8dd855948498dc9243f7ca27336f036d7f212b74bc103c9a7",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q0746.jpg",
		"1049d96239e24d4dca2c55512b8bdb77425f4dbcf575a0a95555aaab5554aaaa",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q1050.jpg",
		"489db672e9190276d452aeab41eba20f02375fe4092d88defdf491a5c55c5f70",
	},
	{
		"./test-images/reg-test-input/labelme-subset/q2821.jpg",
		"b150231ffae4710ffcf4f18bb574b109a576f14bb8543189f8743289f174b109",
	},
	{
		"./test-images/reg-test-input/dih/bridge-1-original.jpg",
		"d8f8f0cce0f4a84f0e370a22028f67f0b36e2ed596623e1d33e6b39c4e9c9b22",
	},
	{
		"./test-images/reg-test-input/dih/bridge-2-rotate-90.jpg",
		"38a50efd71c83f429013d68d0ffffc52e34e0e15ada952a9d29684214aa9e5af",
	},
	{
		"./test-images/reg-test-input/dih/bridge-3-rotate-180.jpg",
		"2dadda64b5a142e5d362209057da895ae63b8c7fc277b4b766b319361f893188",
	},
	{
		"./test-images/reg-test-input/dih/bridge-4-rotate-270.jpg",
		"a5f0a457248995e8c9065c275aaa54d8b61ba4bdf8fcfc0387c32f8b0bfc4f05",
	},
	{
		"./test-images/reg-test-input/dih/bridge-5-flipx.jpg",
		"d8f80f31e0f417b00e37f5dd028f980fb36ed12a9662c1e233e64c634e9c64dd",
	},
	{
		"./test-images/reg-test-input/dih/bridge-6-flipy.jpg",
		"0dad259bb1a1bd18d362576556da32a1e63b7380c2374b4866b3c6c91b89ce77",
	},
	{
		"./test-images/reg-test-input/dih/bridge-7-flip-plus-1.jpg",
		"f0a5e10271dcc0bd9c5309720fff018de34ef1e8ada9a956d2967ade1ea91a50",
	},
	{
		"./test-images/reg-test-input/dih/bridge-8-flip-minus-1.jpg",
		"69f05aa8a4996a17c146a2da5aaaab07b61b5b60f8fc07fc83c3d0740bfcb0fa",
	},
}

func TestPDQHasher(t *testing.T) {
	vips.LoggingSettings(nil, vips.LogLevelMessage)
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 0,
		MaxCacheFiles:    5,
		MaxCacheMem:      50 * 1024 * 1024,
		MaxCacheSize:     100,
		ReportLeaks:      false,
		CacheTrace:       false,
		CollectStats:     false,
	})
	defer vips.Shutdown()
	pdqHasher := NewPDQHasher()
	hammingTolerance := 16

	for _, pair := range DATA_ARRAY {
		imagePath := pair.First
		expectedHash := pair.Second

		hash := pdqHasher.FromFile(imagePath)

		expectedHashObj, err := types.Hash256FromHexString(expectedHash)
		assert.ErrorIs(t, err, nil)

		computedHash := hash.Hash

		hammingDistance := computedHash.HammingDistance(expectedHashObj)
		log.Printf("Image: %s Hamming distance: %d\n", imagePath, hammingDistance)
		assert.True(t, hammingDistance <= hammingTolerance)
	}
}