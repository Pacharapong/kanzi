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

package transform

import "errors"

// Sort by Rank Transform is a family of transforms typically used after
// a BWT to reduce the variance of the data prior to entropy coding.
// SBR(alpha) is defined by sbr(x, alpha) = (1-alpha)*(t-w1(x,t)) + alpha*(t-w2(x,t))
// where x is an item in the data list, t is the current access time and wk(x,t) is
// the k-th access time to x at time t (with 0 <= alpha <= 1).
// See [Two new families of list update algorihtms] by Frank Schulz for details.
// It turns out that SBR(0)= Move to Front Transform
// It turns out that SBR(1)= Time Stamp Transform
// This code implements SBR(0), SBR(1/2) and SBR(1). Code derived from openBWT

const (
	MODE_MTF       = 1 // alpha = 0
	MODE_RANK      = 2 // alpha = 1/2
	MODE_TIMESTAMP = 3 // alpha = 1
)

type SBRT struct {
	size    uint
	prev    []int // size 256
	curr    []int // size 256
	symbols []int // size 256
	ranks   []int // size 256
	mode    int
}

func NewSBRT(mode int, sz uint) (*SBRT, error) {
	if mode != MODE_MTF && mode != MODE_RANK && mode != MODE_TIMESTAMP {
		return nil, errors.New("Invalid mode parameter")
	}

	this := new(SBRT)
	this.size = sz
	this.mode = mode
	this.prev = make([]int, 256)
	this.curr = make([]int, 256)
	this.symbols = make([]int, 256)
	this.ranks = make([]int, 256)
	return this, nil
}

func (this *SBRT) Size() uint {
	return this.size
}

func (this *SBRT) SetSize(sz uint) bool {
	this.size = sz
	return true
}

func (this *SBRT) Forward(src, dst []byte) (uint, uint, error) {
	count := int(this.size)

	if count == 0 {
		count = len(src)
	}

	// Aliasing
	p := this.prev
	q := this.curr
	s2r := this.symbols
	r2s := this.ranks

	var mask1, mask2 int
	var shift uint

	if this.mode == MODE_TIMESTAMP {
		mask1 = 0
	} else {
		mask1 = -1
	}

	if this.mode == MODE_MTF {
		mask2 = 0
	} else {
		mask2 = -1
	}

	if this.mode == MODE_RANK {
		shift = 1
	} else {
		shift = 0
	}

	for i := 0; i < 256; i++ {
		p[i] = 0
		q[i] = 0
		s2r[i] = i
		r2s[i] = i
	}

	for i := 0; i < count; i++ {
		c := uint(src[i])
		r := s2r[c]
		dst[i] = byte(r)
		q[c] = ((i & mask1) + (p[c] & mask2)) >> shift
		p[c] = i
		curVal := q[c]

		// Move up symbol to correct rank
		for r > 0 && q[r2s[r-1]] <= curVal {
			r2s[r] = r2s[r-1]
			s2r[r2s[r]] = r
			r--
		}

		r2s[r] = int(c)
		s2r[c] = r
	}

	return uint(count), uint(count), nil
}

func (this *SBRT) Inverse(src, dst []byte) (uint, uint, error) {
	count := int(this.size)

	if count == 0 {
		count = len(src)
	}

	// Aliasing
	p := this.prev
	q := this.curr
	r2s := this.ranks

	var mask1, mask2 int
	var shift uint

	if this.mode == MODE_TIMESTAMP {
		mask1 = 0
	} else {
		mask1 = -1
	}

	if this.mode == MODE_MTF {
		mask2 = 0
	} else {
		mask2 = -1
	}

	if this.mode == MODE_RANK {
		shift = 1
	} else {
		shift = 0
	}

	for i := 0; i < 256; i++ {
		p[i] = 0
		q[i] = 0
		r2s[i] = i
	}

	for i := 0; i < count; i++ {
		r := uint(src[i])
		c := r2s[r]
		dst[i] = byte(c)
		q[c] = ((i & mask1) + (p[c] & mask2)) >> shift
		p[c] = i
		curVal := q[c]

		// Move up symbol to correct rank
		for r > 0 && q[r2s[r-1]] <= curVal {
			r2s[r] = r2s[r-1]
			r--
		}

		r2s[r] = c
	}

	return uint(count), uint(count), nil
}
