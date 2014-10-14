/*
Copyright 2011-2013 Frederic Langlet
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
you may obtain a copy of the License at

                http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package entropy

import (
	"errors"
	"fmt"
	"kanzi"
)

const (
	TOP_RANGE                = uint64(0x00FFFFFFFFFFFFFF)
	BOTTOM_RANGE             = uint64(0x00000000FFFFFFFF)
	MASK                     = uint64(0x00FFFF0000000000)
	DEFAULT_RANGE_CHUNK_SIZE = uint(1 << 16) // 64 KB by default
	DEFAULT_RANGE_LOG_RANGE  = uint(13)
)

type RangeEncoder struct {
	low       uint64
	range_    uint64
	invSum    uint64
	bitstream kanzi.OutputBitStream
	freqs     []int
	cumFreqs  []int
	alphabet  []byte
	eu        *EntropyUtils
	chunkSize int
	logRange  uint
}

// The chunk size indicates how many bytes are encoded (per block) before
// resetting the frequency stats. 0 means that frequencies calculated at the
// beginning of the block apply to the whole block.
// Since the number of args is variable, this function can be called like this:
// NewRangeEncoder(bs) or NewRangeEncoder(bs, 16384, 14)
// The default chunk size is 65536 bytes.
func NewRangeEncoder(bs kanzi.OutputBitStream, args ...uint) (*RangeEncoder, error) {
	if bs == nil {
		return nil, errors.New("Invalid null bitstream parameter")
	}

	if len(args) > 2 {
		return nil, errors.New("At most one chunk size and one log range can be provided")
	}

	chkSize := DEFAULT_RANGE_CHUNK_SIZE
	logRange := DEFAULT_RANGE_LOG_RANGE

	if len(args) == 2 {
		chkSize = args[0]
		logRange = args[1]
	}

	if chkSize != 0 && chkSize < 1024 {
		return nil, errors.New("The chunk size must be at least 1024")
	}

	if chkSize > 1<<30 {
		return nil, errors.New("The chunk size must be at most 2^30")
	}

	if logRange < 8 || logRange > 16 {
		return nil, fmt.Errorf("Invalid range parameter: %v (must be in [8..16])", logRange)
	}

	this := new(RangeEncoder)
	this.bitstream = bs
	this.alphabet = make([]byte, 256)
	this.freqs = make([]int, 256)
	this.cumFreqs = make([]int, 257)
	this.logRange = logRange
	this.chunkSize = int(chkSize)
	var err error
	this.eu, err = NewEntropyUtils()
	return this, err
}

func (this *RangeEncoder) updateFrequencies(frequencies []int, size int, lr uint) (int, error) {
	if frequencies == nil || len(frequencies) != 256 {
		return 0, errors.New("Invalid frequencies parameter")
	}

	alphabetSize, err := this.eu.NormalizeFrequencies(frequencies, this.alphabet, size, 1<<lr)

	if err != nil {
		return alphabetSize, err
	}

	if alphabetSize > 0 {
		this.cumFreqs[0] = 0

		// Create histogram of frequencies scaled to 'range'
		for i := 0; i < 256; i++ {
			this.cumFreqs[i+1] = this.cumFreqs[i] + frequencies[i]
		}

		this.invSum = uint64(1 << 24) / uint64(this.cumFreqs[256])
		this.encodeHeader(alphabetSize, this.alphabet, frequencies, lr)
	}

	return alphabetSize, nil
}

func (this *RangeEncoder) encodeHeader(alphabetSize int, alphabet []byte, frequencies []int, lr uint) bool {
	EncodeAlphabet(this.bitstream, alphabet[0:alphabetSize])

	if alphabetSize == 0 {
		return true
	}

	this.bitstream.WriteBits(uint64(lr-8), 3) // logRange
	inc := 16

	if alphabetSize <= 64 {
		inc = 8
	}

	llr := uint(3)

	for 1<<llr <= lr {
		llr++
	}

	/// Encode all frequencies (but the first one) by chunks of size 'inc'
	for i := 1; i < alphabetSize; i += inc {
		max := 0
		logMax := uint(1)
		endj := i + inc

		if endj > alphabetSize {
			endj = alphabetSize
		}

		// Search for max frequency log size in next chunk
		for j := i; j < endj; j++ {
			if frequencies[alphabet[j]] > max {
				max = frequencies[alphabet[j]]
			}
		}

		for 1<<logMax <= max {
			logMax++
		}

		this.bitstream.WriteBits(uint64(logMax-1), llr)

		// Write frequencies
		for j := i; j < endj; j++ {
			this.bitstream.WriteBits(uint64(frequencies[alphabet[j]]), logMax)
		}
	}

	return true
}

func (this *RangeEncoder) Encode(block []byte) (int, error) {
	if block == nil {
		return 0, errors.New("Invalid null block parameter")
	}

	if len(block) == 0 {
		return 0, nil
	}

	sizeChunk := this.chunkSize

	if sizeChunk == 0 {
		sizeChunk = len(block)
	}

	frequencies := this.freqs // aliasing
	startChunk := 0
	end := len(block)

	for startChunk < end {
		this.range_ = TOP_RANGE
		this.low = 0
		lr := this.logRange

		endChunk := startChunk + sizeChunk

		if endChunk > end {
			endChunk = end
		}

		// Lower log range if the size of the data block is small
		for lr > 8 && 1<<lr > endChunk-startChunk {
			lr--
		}

		for i := range frequencies {
			frequencies[i] = 0
		}

		for i := startChunk; i < endChunk; i++ {
			frequencies[block[i]]++
		}

		// Rebuild statistics
		if _, err := this.updateFrequencies(frequencies, endChunk-startChunk, lr); err != nil {
			return startChunk, err
		}

		for i := startChunk; i < endChunk; i++ {
			this.encodeByte(block[i])
		}

		// Flush 'low'
		this.bitstream.WriteBits(this.low, 56)
		startChunk = endChunk
	}

	return len(block), nil
}

func (this *RangeEncoder) encodeByte(b byte) {
	value := int(b)
	symbolLow := uint64(this.cumFreqs[value])
	symbolHigh := uint64(this.cumFreqs[value+1])

	// Compute next low and range
	this.range_ = (this.range_ >> 24) * this.invSum
	this.low += (symbolLow * this.range_)
	this.range_ *= (symbolHigh - symbolLow)

	// If the left-most digits are the same throughout the range, write bits to bitstream
	for {
		if (this.low^(this.low+this.range_))&MASK != 0 {
			if this.range_ > BOTTOM_RANGE {
				break
			}

			// Normalize
			this.range_ = -this.low & BOTTOM_RANGE
		}

		this.bitstream.WriteBits(this.low>>40, 16)
		this.range_ <<= 16
		this.low <<= 16
	}
}

func (this *RangeEncoder) BitStream() kanzi.OutputBitStream {
	return this.bitstream
}

func (this *RangeEncoder) Dispose() {
}

type RangeDecoder struct {
	code      uint64
	low       uint64
	range_    uint64
	invSum    uint64
	bitstream kanzi.InputBitStream
	freqs     []int
	cumFreqs  []int
	f2s       []byte // mapping frequency -> symbol
	alphabet  []byte
	chunkSize int
}

// The chunk size indicates how many bytes are encoded (per block) before
// resetting the frequency stats. 0 means that frequencies calculated at the
// beginning of the block apply to the whole block
// Since the number of args is variable, this function can be called like this:
// NewRangeDecoder(bs) or NewRangeDecoder(bs, 16384, 14)
// The default chunk size is 65536 bytes.
func NewRangeDecoder(bs kanzi.InputBitStream, args ...uint) (*RangeDecoder, error) {
	if bs == nil {
		return nil, errors.New("Invalid null bitstream parameter")
	}

	if len(args) > 1 {
		return nil, errors.New("At most one chunk size can be provided")
	}

	chkSize := DEFAULT_RANGE_CHUNK_SIZE

	if len(args) == 1 {
		chkSize = args[0]
	}

	if chkSize != 0 && chkSize < 1024 {
		return nil, errors.New("The chunk size must be at least 1024")
	}

	if chkSize > 1<<30 {
		return nil, errors.New("The chunk size must be at most 2^30")
	}

	this := new(RangeDecoder)
	this.bitstream = bs
	this.alphabet = make([]byte, 256)
	this.freqs = make([]int, 256)
	this.cumFreqs = make([]int, 257)
	this.f2s = make([]byte, 0)
	this.chunkSize = int(chkSize)
	return this, nil
}

func (this *RangeDecoder) decodeHeader(frequencies []int) (int, uint, error) {
	alphabetSize, err := DecodeAlphabet(this.bitstream, this.alphabet)

	if err != nil || alphabetSize == 0 {
		return alphabetSize, 0, nil
	}

	if alphabetSize != 256 {
		for i := range frequencies {
			frequencies[i] = 0
		}
	}

	// Decode frequencies
	logRange := uint(8 + this.bitstream.ReadBits(3))
	sum := 0
	inc := 16
	llr := uint(3)

	if alphabetSize <= 64 {
		inc = 8
	}

	for 1<<llr <= logRange {
		llr++
	}

	// Decode all frequencies (but the first one) by chunks of size 'inc'
	for i := 1; i < alphabetSize; i += inc {
		logMax := uint(1 + this.bitstream.ReadBits(llr))
		endj := i + inc

		if endj > alphabetSize {
			endj = alphabetSize
		}

		// Read frequencies
		for j := i; j < endj; j++ {
			val := int(this.bitstream.ReadBits(logMax))

			if val <= 0 || val >= 1<<logRange {
				error := fmt.Errorf("Invalid bitstream: incorrect frequency %v  for symbol '%v' in ANS range decoder", val, this.alphabet[j])
				return alphabetSize, logRange, error
			}

			frequencies[this.alphabet[j]] = val
			sum += val
		}
	}

	// Infer first frequency
	frequencies[this.alphabet[0]] = (1 << logRange) - sum

	if frequencies[this.alphabet[0]] <= 0 || frequencies[this.alphabet[0]] > 1<<logRange {
		error := fmt.Errorf("Invalid bitstream: incorrect frequency %v  for symbol '%v' in ANS range decoder", frequencies[this.alphabet[0]], this.alphabet[0])
		return alphabetSize, logRange, error
	}

	this.cumFreqs[0] = 0

	if len(this.f2s) < 1<<logRange {
		this.f2s = make([]byte, 1<<logRange)
	}

	// Create histogram of frequencies scaled to 'range' and reverse mapping
	for i := 0; i < 256; i++ {
		this.cumFreqs[i+1] = this.cumFreqs[i] + frequencies[i]

		for j := frequencies[i] - 1; j >= 0; j-- {
			this.f2s[this.cumFreqs[i]+j] = byte(i)
		}
	}

	this.invSum = uint64(1 << 24) / uint64(this.cumFreqs[256])
	return alphabetSize, logRange, nil
}

// Initialize once (if necessary) at the beginning, the use the faster decodeByte_()
// Reset frequency stats for each chunk of data in the block
func (this *RangeDecoder) Decode(block []byte) (int, error) {
	if block == nil {
		return 0, errors.New("Invalid null block parameter")
	}

	end := len(block)
	startChunk := 0
	sizeChunk := this.chunkSize

	if sizeChunk == 0 {
		sizeChunk = len(block)
	}

	for startChunk < end {
		alphabetSize, _, err := this.decodeHeader(this.freqs)

		if err != nil || alphabetSize == 0 {
			return startChunk, err
		}

		this.range_ = TOP_RANGE
		this.low = 0
		this.code = this.bitstream.ReadBits(56)
		endChunk := startChunk + sizeChunk

		if endChunk > end {
			endChunk = end
		}

		for i := startChunk; i < endChunk; i++ {
			block[i] = this.decodeByte()
		}

		startChunk = endChunk
	}

	return len(block), nil
}

func (this *RangeDecoder) decodeByte() byte {
	this.range_ = (this.range_ >> 24) * this.invSum
	count := int((this.code - this.low) / this.range_)
	value := int(this.f2s[count])

	// Compute next low and range
	symbolLow := uint64(this.cumFreqs[value])
	symbolHigh := uint64(this.cumFreqs[value+1])
	this.low += (symbolLow * this.range_)
	this.range_ *= (symbolHigh - symbolLow)

	for {
		if (this.low^(this.low+this.range_))&MASK != 0 {
			if this.range_ > BOTTOM_RANGE {
				break
			}

			// Normalize
			this.range_ = -this.low & BOTTOM_RANGE
		}

		this.code = (this.code << 16) | this.bitstream.ReadBits(16)
		this.range_ <<= 16
		this.low <<= 16
	}

	return byte(value)
}

func (this *RangeDecoder) BitStream() kanzi.InputBitStream {
	return this.bitstream
}

func (this *RangeDecoder) Dispose() {
}
