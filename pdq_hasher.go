package pdq

//lint:file-ignore U1000 Ignore all unused code, it's pending tests

import (
	"log"

	"math"

	"github.com/MTRNord/pdqhash-go/helpers"
	"github.com/MTRNord/pdqhash-go/types"
	"github.com/davidbyttow/govips/v2/vips"

	_ "image/jpeg"
)

// From Wikipedia: standard RGB to luminance (the 'Y' in 'YUV').
const LUMA_FROM_R_COEFF = 0.299
const LUMA_FROM_G_COEFF = 0.587
const LUMA_FROM_B_COEFF = 0.114

func DCT_MATRIX_SCALE_FACTOR() float64 {
	return math.Sqrt(2.0 / 64.0)
}

// Wojciech Jarosz 'Fast Image Convolutions' ACM SIGGRAPH 2001:
// X,Y,X,Y passes of 1-D box filters produces a 2D tent filter.
const PDQ_NUM_JAROSZ_XY_PASSES = 2

/*
Since PDQ uses 64x64 blocks, 1/64th of the image height/width
respectively is a full block. But since we use two passes, we want half
that window size per pass. Example: 1024x1024 full-resolution input. PDQ
downsamples to 64x64. Each 16x16 block of the input produces a single
downsample pixel.  X,Y passes with window size 8 (= 1024/128) average
pixels with 8x8 neighbors. The second X,Y pair of 1D box-filter passes
accumulate data from all 16x16.
*/
const PDQ_JAROSZ_WINDOW_SIZE_DIVISOR = 128

// Flags for which dihedral-transforms are desired to be produced.
const PDQ_DO_DIH_ORIGINAL = 0x01
const PDQ_DO_DIH_ROTATE_90 = 0x02
const PDQ_DO_DIH_ROTATE_180 = 0x04
const PDQ_DO_DIH_ROTATE_270 = 0x08
const PDQ_DO_DIH_FLIPX = 0x10
const PDQ_DO_DIH_FLIPY = 0x20
const PDQ_DO_DIH_FLIP_PLUS1 = 0x40
const PDQ_DO_DIH_FLIP_MINUS1 = 0x80
const PDQ_DO_DIH_ALL = 0xFF

/**
 * The only class state is the DCT matrix, so this class may either be
 * instantiated once per image, or instantiated once and used for all images;
 * the latter will be slightly faster as the DCT matrix will not need to be
 * recomputed once per image.
 */
type PDQHasher struct {
	DCT_matrix [][]float64
}

/**
 * Container for multiple-value object: the hash is a 64-character hex
 * string and the quality is an integer in the range 0..100.
 */
type HashAndQuality struct {
	Hash    *types.Hash256
	Quality int
}

type HashesAndQuality struct {
	hash           *types.Hash256
	hashRotate90   *types.Hash256
	hashRotate180  *types.Hash256
	hashRotate270  *types.Hash256
	hashFlipX      *types.Hash256
	hashFlipY      *types.Hash256
	hashFlipPlus1  *types.Hash256
	hashFlipMinus1 *types.Hash256
	quality        int
}

func NewPDQHasher() *PDQHasher {
	return &PDQHasher{
		DCT_matrix: ComputeDCTMatrix(),
	}
}

func ComputeDCTMatrix() [][]float64 {
	d := make([][]float64, 16)
	for i := 0; i < len(d); i++ {
		di := make([]float64, 64)
		for j := 0; j < len(di); j++ {
			di[j] = DCT_MATRIX_SCALE_FACTOR() * math.Cos((math.Pi/2.0/64.0)*(float64(i)+1.0)*(2.0*float64(j)+1.0))
		}
		d[i] = di
	}

	return d
}

func allocateMatrix(numRows, numCols int) [][]float64 {
	// Create a slice of slices to represent the matrix
	matrix := make([][]float64, numRows)

	// Allocate memory for each row
	for i := range matrix {
		matrix[i] = make([]float64, numCols)
	}

	return matrix
}

func (p *PDQHasher) FromFile(filename string) HashAndQuality {
	params := vips.NewImportParams()
	params.AutoRotate.Set(false)

	image, err := vips.LoadImageFromFile(filename, params)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	// resizing the image proportionally to max 512px width and max 512px height
	err = image.ThumbnailWithSize(512, 512, vips.InterestingNone, vips.SizeDown)
	if err != nil {
		log.Fatalf("Error resizing image: %v", err)
	}
	numCols := image.Width()
	numRows := image.Height()

	buffer1 := make([]float64, numCols*numRows)
	buffer2 := make([]float64, numCols*numRows)
	buffer64x64 := allocateMatrix(64, 64)
	buffer16x64 := allocateMatrix(16, 64)
	buffer16x16 := allocateMatrix(16, 16)

	return p.FromImage(image, buffer1, buffer2, buffer64x64, buffer16x64, buffer16x16)
}

func (p *PDQHasher) FromImage(image *vips.ImageRef, buffer1, buffer2 []float64, buffer64x64, buffer16x64, buffer16x16 [][]float64) HashAndQuality {
	numCols := image.Width()
	numRows := image.Height()

	p.fillFloatLumaFromBufferImage(image, &buffer1)

	return p.pdqHash256FromFloatLuma(buffer1, buffer2, numRows, numCols, buffer64x64, buffer16x64, buffer16x16)
}

func (p *PDQHasher) fillFloatLumaFromBufferImage(image *vips.ImageRef, luma *[]float64) {
	numCols := image.Width()
	numRows := image.Height()

	err := image.ToColorSpace(vips.InterpretationSRGB)
	if err != nil {
		log.Fatalf("Error converting to RGB: %v", err)
	}

	for i := 0; i < numRows; i++ {
		for j := 0; j < numCols; j++ {
			colorArray, err := image.GetPoint(j, i)
			if err != nil {
				log.Fatalf("Error getting pixel: %v", err)
			}
			r := colorArray[0]
			g := colorArray[1]
			b := colorArray[2]
			(*luma)[i*numCols+j] = LUMA_FROM_R_COEFF*float64(r) + LUMA_FROM_G_COEFF*float64(g) + LUMA_FROM_B_COEFF*float64(b)
		}
	}
}

func (p *PDQHasher) pdqHash256FromFloatLuma(fullBuffer1, fullBuffer2 []float64, numRows, numCols int, buffer64x64, buffer16x64, buffer16x16 [][]float64) HashAndQuality {
	windowSizeAlongRows := p.computeJaroszWindowSize(numCols)
	windowSizeAlongCols := p.computeJaroszWindowSize(numRows)
	p.jaroszFilterFloat(&fullBuffer1, &fullBuffer2, numRows, numCols, windowSizeAlongRows, windowSizeAlongCols, PDQ_NUM_JAROSZ_XY_PASSES)

	p.decimateFloat(&fullBuffer1, numRows, numCols, &buffer64x64)
	quality := p.computePDQImageDomainQualityMetric(buffer64x64)
	p.dct64To16(&buffer64x64, &buffer16x64, &buffer16x16)
	hash := p.pdqBuffer16x16ToBits(buffer16x16)

	return HashAndQuality{hash, quality}
}

func (p *PDQHasher) DihedralFromFile(filename string, dihedralFlags int) HashesAndQuality {
	image, err := vips.NewImageFromFile(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	numRows := image.Height()
	numCols := image.Width()

	buffer1 := make([]float64, numCols*numRows)
	buffer2 := make([]float64, numCols*numRows)

	buffer64x64 := allocateMatrix(64, 64)
	buffer16x64 := allocateMatrix(16, 64)
	buffer16x16 := allocateMatrix(16, 16)
	buffer16x16Aux := allocateMatrix(16, 16)

	return p.dihedralFromBufferedImage(image, buffer1, buffer2, buffer64x64, buffer16x64, buffer16x16, buffer16x16Aux, dihedralFlags)
}

func (p *PDQHasher) dihedralFromBufferedImage(image *vips.ImageRef, buffer1, buffer2 []float64, buffer64x64, buffer16x64, buffer16x16, buffer16x16Aux [][]float64, dihedralFlags int) HashesAndQuality {
	numRows := image.Height()
	numCols := image.Width()

	p.fillFloatLumaFromBufferImage(image, &buffer1)

	return p.pdqHash256esFromFloatLuma(buffer1, buffer2, numRows, numCols, buffer64x64, buffer16x64, buffer16x16, buffer16x16Aux, dihedralFlags)
}

func (p *PDQHasher) pdqHash256esFromFloatLuma(fullBuffer1, fullBuffer2 []float64, numRows, numCols int, buffer64x64, buffer16x64, buffer16x16, buffer16x16Aux [][]float64, dihedralFlags int) HashesAndQuality {
	windowSizeAlongRows := p.computeJaroszWindowSize(numCols)
	windowSizeAlongCols := p.computeJaroszWindowSize(numRows)
	p.jaroszFilterFloat(&fullBuffer1, &fullBuffer2, numRows, numCols, windowSizeAlongRows, windowSizeAlongCols, PDQ_NUM_JAROSZ_XY_PASSES)

	p.decimateFloat(&fullBuffer1, numRows, numCols, &buffer64x64)
	quality := p.computePDQImageDomainQualityMetric(buffer64x64)
	p.dct64To16(&buffer64x64, &buffer16x64, &buffer16x16)

	var hash *types.Hash256
	var hashRotate90 *types.Hash256
	var hashRotate180 *types.Hash256
	var hashRotate270 *types.Hash256
	var hashFlipX *types.Hash256
	var hashFlipY *types.Hash256
	var hashFlipPlus1 *types.Hash256
	var hashFlipMinus1 *types.Hash256

	if dihedralFlags&PDQ_DO_DIH_ORIGINAL != 0 {
		hash = p.pdqBuffer16x16ToBits(buffer16x16)
	}

	if dihedralFlags&PDQ_DO_DIH_ROTATE_90 != 0 {
		p.dct16OriginalToRotate90(&buffer16x16, &buffer16x16Aux)
		hashRotate90 = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	if dihedralFlags&PDQ_DO_DIH_ROTATE_180 != 0 {
		p.dct16OriginalToRotate180(&buffer16x16, &buffer16x16Aux)
		hashRotate180 = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	if dihedralFlags&PDQ_DO_DIH_ROTATE_270 != 0 {
		p.dct16OriginalToRotate270(&buffer16x16, &buffer16x16Aux)
		hashRotate270 = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	if dihedralFlags&PDQ_DO_DIH_FLIPX != 0 {
		p.dct16OriginalToFlipX(&buffer16x16, &buffer16x16Aux)
		hashFlipX = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	if dihedralFlags&PDQ_DO_DIH_FLIPY != 0 {
		p.dct16OriginalToFlipY(&buffer16x16, &buffer16x16Aux)
		hashFlipY = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	if dihedralFlags&PDQ_DO_DIH_FLIP_PLUS1 != 0 {
		p.dct16OriginalToFlipPlus1(&buffer16x16, &buffer16x16Aux)
		hashFlipPlus1 = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	if dihedralFlags&PDQ_DO_DIH_FLIP_MINUS1 != 0 {
		p.dct16OriginalToFlipMinus1(&buffer16x16, &buffer16x16Aux)
		hashFlipMinus1 = p.pdqBuffer16x16ToBits(buffer16x16Aux)
	}

	return HashesAndQuality{hash, hashRotate90, hashRotate180, hashRotate270, hashFlipX, hashFlipY, hashFlipPlus1, hashFlipMinus1, quality}
}

// numRows x numCols in row-major order
func (p *PDQHasher) decimateFloat(in *[]float64, inNumRows, inNumCols int, out *[][]float64) {
	for i := 0; i < 64; i++ {
		ini := int(((float64(i) + 0.5) * float64(inNumRows)) / 64.0)
		for j := 0; j < 64; j++ {
			inj := int(((float64(j) + 0.5) * float64(inNumCols)) / 64.0)
			(*out)[i][j] = (*in)[ini*inNumCols+inj]
		}
	}
}

/**
 * This is all heuristic (see the PDQ hashing doc). Quantization
 * matters since we want to count *significant* gradients, not just the
 * some of many small ones. The constants are all manually selected, and
 * tuned as described in the document.
 */
func (p *PDQHasher) computePDQImageDomainQualityMetric(buffer64x64 [][]float64) int {
	gradientSum := 0
	for i := 0; i < 63; i++ {
		for j := 0; j < 64; j++ {
			u := buffer64x64[i][j]
			v := buffer64x64[i+1][j]
			d := int(((u - v) * 100.0) / 255.0)
			gradientSum += int(helpers.Abs(d))
		}
	}
	for i := 0; i < 64; i++ {
		for j := 0; j < 63; j++ {
			u := buffer64x64[i][j]
			v := buffer64x64[i][j+1]
			d := int(((u - v) * 100.0) / 255.0)
			gradientSum += int(helpers.Abs(d))
		}
	}
	quality := float64(gradientSum) / 90.0
	if quality > 100 {
		quality = 100
	}
	return int(quality)
}

/**
 * Full 64x64 to 64x64 can be optimized e.g. the Lee algorithm.
 *    But here we only want slots (1-16)x(1-16) of the full 64x64 output.
 *    Careful experiments showed that using Lee along all 64 slots in one
 *    dimension, then Lee along 16 slots in the second, followed by
 *    extracting slots 1-16 of the output, was actually slower than the
 *    current implementation which is completely non-clever/non-Lee but
 *    computes only what is needed.
 */
func (p *PDQHasher) dct64To16(A, T, B *([][]float64)) {
	D := p.DCT_matrix

	*T = make([][]float64, 16)
	for i := 0; i < 16; i++ {
		ti := make([]float64, 64)

		for j := 0; j < 64; j++ {
			tij := float64(0.0)
			for k := 0; k < 64; k++ {
				tij += D[i][k] * (*A)[k][j]
			}
			ti[j] = tij
		}
		(*T)[i] = ti
	}

	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			sumk := float64(0.0)
			for k := 0; k < 64; k++ {
				sumk += (*T)[i][k] * D[j][k]
			}
			(*B)[i][j] = sumk
		}
	}
}

/*
   -------------------------------------
   orig      rot90     rot180    rot270
   noxpose   xpose     noxpose   xpose
   + + + +   - + - +   + - + -   - - - -
   + + + +   - + - +   - + - +   + + + +
   + + + +   - + - +   + - + -   - - - -
   + + + +   - + - +   - + - +   + + + +

   flipx     flipy     flipplus  flipminus
   noxpose   noxpose   xpose     xpose
   - - - -   - + - +   + + + +   + - + -
   + + + +   - + - +   + + + +   - + - +
   - - - -   - + - +   + + + +   + - + -
   + + + +   - + - +   + + + +   - + - +
   -------------------------------------
*/

func (p *PDQHasher) dct16OriginalToRotate90(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if (j & 1) != 0 {
				(*B)[j][i] = (*A)[i][j]
			} else {
				(*B)[j][i] = -(*A)[i][j]
			}
		}
	}
}

func (p *PDQHasher) dct16OriginalToRotate180(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if ((i + j) & 1) != 0 {
				(*B)[i][j] = -(*A)[i][j]
			} else {
				(*B)[i][j] = (*A)[i][j]
			}
		}
	}
}

func (p *PDQHasher) dct16OriginalToRotate270(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if (i & 1) != 0 {
				(*B)[j][i] = (*A)[i][j]
			} else {
				(*B)[j][i] = -(*A)[i][j]
			}
		}
	}
}

func (p *PDQHasher) dct16OriginalToFlipX(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if (i & 1) != 0 {
				(*B)[i][j] = (*A)[i][j]
			} else {
				(*B)[i][j] = -(*A)[i][j]
			}
		}
	}
}

func (p *PDQHasher) dct16OriginalToFlipY(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if (j & 1) != 0 {
				(*B)[i][j] = (*A)[i][j]
			} else {
				(*B)[i][j] = -(*A)[i][j]
			}
		}
	}
}

func (p *PDQHasher) dct16OriginalToFlipPlus1(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			(*B)[j][i] = (*A)[i][j]
		}
	}
}

func (p *PDQHasher) dct16OriginalToFlipMinus1(A, B *[][]float64) {
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if ((i + j) & 1) != 0 {
				(*B)[j][i] = -(*A)[i][j]
			} else {
				(*B)[j][i] = (*A)[i][j]
			}
		}
	}
}

/**
 * Each bit of the 16x16 output hash is for whether the given frequency
 * component is greater than the median frequency component or not.
 */
func (p *PDQHasher) pdqBuffer16x16ToBits(dctOutput16x16 [][]float64) *types.Hash256 {
	hash := types.Hash256{}
	dctMedian := helpers.Torben(dctOutput16x16, 16, 16)
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			if dctOutput16x16[i][j] > dctMedian {
				hash.SetBit(i*16 + j)
			}
		}
	}
	return &hash
}

// Round up.
func (p *PDQHasher) computeJaroszWindowSize(dimension int) int {
	result := (float64(dimension) + float64(PDQ_JAROSZ_WINDOW_SIZE_DIVISOR) - 1.0) / float64(PDQ_JAROSZ_WINDOW_SIZE_DIVISOR)
	return int(result)
}

func (p *PDQHasher) jaroszFilterFloat(buffer1, buffer2 *[]float64, numRows, numCols, windowSizeAlongRows, windowSizeAlongCols, nreps int) {
	for i := 0; i < nreps; i++ {
		p.boxAlongRowsFloat(buffer1, buffer2, numRows, numCols, windowSizeAlongRows)
		p.boxAlongColsFloat(buffer2, buffer1, numRows, numCols, windowSizeAlongCols)
	}
}

func (p *PDQHasher) box1DFloat(invec *[]float64, inStartOffset int, outvec *[]float64, outStartOffset, vectorLength, stride, fullWindowSize int) {
	halfWindowSize := int(((float64(fullWindowSize) + 2.0) / 2.0))
	phase_1_nreps := int(halfWindowSize - 1)
	phase_2_nreps := int(fullWindowSize - halfWindowSize + 1)
	phase_3_nreps := int(vectorLength - fullWindowSize)
	phase_4_nreps := int(halfWindowSize - 1)
	li := 0 // Index of left edge of read window, for subtracts
	ri := 0 // Index of right edge of read windows, for adds
	oi := 0 // Index of output vector
	sum := float64(0.0)
	currentWindowSize := 0

	// PHASE 1: ACCUMULATE FIRST SUM NO WRITES
	for i := 0; i < phase_1_nreps; i++ {
		sum += (*invec)[inStartOffset+ri]
		currentWindowSize += 1
		ri += stride
	}

	// PHASE 2: INITIAL WRITES WITH SMALL WINDOW
	for i := 0; i < phase_2_nreps; i++ {
		sum += (*invec)[inStartOffset+ri]
		currentWindowSize += 1
		(*outvec)[outStartOffset+oi] = sum / float64(currentWindowSize)
		ri += stride
		oi += stride
	}

	// PHASE 3: WRITES WITH FULL WINDOW
	for i := 0; i < phase_3_nreps; i++ {
		sum += (*invec)[inStartOffset+ri]
		sum -= (*invec)[inStartOffset+li]
		(*outvec)[outStartOffset+oi] = sum / float64(currentWindowSize)
		li += stride
		ri += stride
		oi += stride
	}

	// PHASE 4: FINAL WRITES WITH SMALL WINDOW
	for i := 0; i < phase_4_nreps; i++ {
		sum -= (*invec)[inStartOffset+li]
		currentWindowSize -= 1
		(*outvec)[outStartOffset+oi] = sum / float64(currentWindowSize)
		li += stride
		oi += stride
	}
}

/**
 * input - matrix as numRows x numCols in row-major order
 * output - matrix as numRows x numCols in row-major order
 */
func (p *PDQHasher) boxAlongRowsFloat(input, output *[]float64, numRows, numCols, windowSize int) {
	for i := 0; i < numRows; i++ {
		p.box1DFloat(input, i*numCols, output, i*numCols, numCols, 1, windowSize)
	}
}

func (p *PDQHasher) boxAlongColsFloat(input, output *[]float64, numRows, numCols, windowSize int) {
	for i := 0; i < numCols; i++ {
		p.box1DFloat(input, i, output, i, numRows, numCols, windowSize)
	}
}
