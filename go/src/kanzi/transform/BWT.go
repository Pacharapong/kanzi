/*
Copyright 2011, 2012 Frederic Langlet
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

type BWT struct {
	size         uint
	data         []int
	buffer1      []int
	buffer2      []byte
	buckets      []int
	primaryIndex uint
}

func NewBWT(sz uint) (*BWT, error) {
	this := new(BWT)
	this.size = sz
	this.data = make([]int, sz)
	this.buffer1 = make([]int, sz)
	this.buffer2 = make([]byte, sz)
	this.buckets = make([]int, 256)
	return this, nil
}

func (this *BWT) PrimaryIndex() uint {
	return this.primaryIndex
}

func (this *BWT) SetPrimaryIndex(primaryIndex uint) bool {
	if primaryIndex < 0 {
		return false
	}

	this.primaryIndex = primaryIndex
	return true
}

func (this *BWT) Size() uint {
	return this.size
}

func (this *BWT) SetSize(sz uint) {
	this.size = sz
}

func (this *BWT) Forward(input []byte) []byte {
	length := int(this.size)

	if this.size == 0 {
		length = len(input)
	}

	if length < 2 {
		return input
	}

	// Dynamic memory allocation
	if len(this.data) < length {
		this.data = make([]int, length)
	}

	if len(this.buffer1) < length {
		this.buffer1 = make([]int, length)
	}

	for i := 0; i < length; i++ {
		this.data[i] = int(input[i]) & 0xFF
	}

	sa := this.buffer1 // alias
	pIdx := computeSuffixArray(this.data, sa, 0, length, 256, true)
	input[0] = byte(this.data[length-1])

	for i := uint(0); i < pIdx; i++ {
		input[i+1] = byte(sa[i])
	}

	for i := pIdx + 1; i < uint(length); i++ {
		input[i] = byte(sa[i])
	}

	this.SetPrimaryIndex(pIdx + 1)
	return input
}

func (this *BWT) Inverse(input []byte) []byte {
	length := int(this.size)

	if this.size == 0 {
		length = len(input)
	}

	// Dynamic memory allocation
	if len(this.data) < length {
		this.data = make([]int, length)
	}

	if len(this.buffer2) < length {
		this.buffer2 = make([]byte, length)
	}

	for i := 0; i < 256; i++ {
		this.buckets[i] = 0
	}

	// Create histogram
	for i := 0; i < length; i++ {
		idx := int(input[i]) & 0xFF
		this.data[i] = this.buckets[idx]
		this.buckets[idx]++
	}

	sum := 0

	// Create cumulative histogram
	for i := 0; i < 256; i++ {
		val := this.buckets[i]
		this.buckets[i] = sum
		sum += val
	}

	pIdx := int(this.PrimaryIndex())
	buffer := this.buffer2 // alias
	val := 0

	for i := length - 1; i >= 0; i-- {
		idx := input[val]
		buffer[i] = idx
		val = this.data[val] + this.buckets[int(idx)&0xFF]

		if val < pIdx {
			val++
		}
	}

	copy(input, buffer[0:length])
	return input
}

func getCounts(src []int, dst []int, n, k int) {
	for i := 0; i < k; i++ {
		dst[i] = 0
	}

	for i := 0; i < n; i++ {
		dst[src[i]]++
	}
}

func getBuckets(src []int, dst []int, k int, end bool) {
	sum := 0

	if end == true {
		for i := 0; i < k; i++ {
			sum += src[i]
			dst[i] = sum
		}
	} else {
		for i := 0; i < k; i++ {
			// The temp variable is required if src == dst
			tmp := src[i]
			dst[i] = sum
			sum += tmp
		}
	}
}

// sort all type LMS suffixes
func sortLMSSuffixes(src []int, sa []int, ptrC *[]int, ptrB *[]int, n, k int) {
	// compute sal
	if ptrC == ptrB {
		getCounts(src, *ptrC, n, k)
	}

	B := *ptrB
	C := *ptrC

	// find starts of buckets
	getBuckets(C, B, k, false)

	j := n - 1
	c1 := src[j]
	b := B[c1]
	j--

	if src[j] < c1 {
		sa[b] = ^j
	} else {
		sa[b] = j
	}

	b++

	for i := 0; i < n; i++ {
		j = sa[i]

		if j > 0 {
			c0 := src[j]

			if c0 != c1 {
				B[c1] = b
				c1 = c0
				b = B[c1]
			}

			j--

			if src[j] < c1 {
				sa[b] = ^j
			} else {
				sa[b] = j
			}

			b++
			sa[i] = 0

		} else if j < 0 {
			sa[i] = ^j
		}
	}

	// compute sas
	if ptrC == ptrB {
		getCounts(src, C, n, k)
	}

	// find ends of buckets
	getBuckets(C, B, k, true)
	c1 = 0
	b = B[c1]

	for i := n - 1; i >= 0; i-- {
		j = sa[i]

		if j > 0 {
			c0 := src[j]

			if c0 != c1 {
				B[c1] = b
				c1 = c0
				b = B[c1]
			}

			j--
			b--

			if src[j] > c1 {
				sa[b] = ^(j + 1)
			} else {
				sa[b] = j
			}

			sa[i] = 0
		}
	}
}

func postProcessLMS(src []int, sa []int, n, m int) int {
	i := 0
	j := 0

	// compact all the sorted substrings into the first m items of sa
	// 2*m must be not larger than n        
	for p := sa[i]; p < 0; i++ {
		sa[i] = ^p
		p = sa[i+1]
	}

	if i < m {
		j = i
		i++

		for true {
			p := sa[i]
			i++

			if p >= 0 {
				continue
			}

			sa[j] = ^p
			sa[i-1] = 0
			j++

			if j == m {
				break
			}
		}
	}

	// store the length of all substrings
	i = n - 2
	j = n - 1
	c0 := src[n-2]
	c1 := src[n-1]

	for i >= 0 && c0 >= c1 {
		c1 = c0
		i--

		if i >= 0 {
			c0 = src[i]
		}
	}

	for i >= 0 {
		c1 = c0
		i--

		if i >= 0 {
			c0 = src[i]
		}

		for i >= 0 && c0 <= c1 {
			c1 = c0
			i--

			if i >= 0 {
				c0 = src[i]
			}
		}

		if i < 0 {
			break
		}

		sa[m+((i+1)>>1)] = j - i
		j = i + 1
		c1 = c0
		i--

		if i >= 0 {
			c0 = src[i]
		}

		for i >= 0 && c0 >= c1 {
			c1 = c0
			i--

			if i >= 0 {
				c0 = src[i]
			}
		}
	}

	// find the lexicographic names of all substrings
	name := 0
	q := n
	qlen := 0

	for ii := 0; ii < m; ii++ {
		p := sa[ii]
		plen := sa[m+(p>>1)]
		diff := true

		if plen == qlen && q+plen < n {
			j = 0

			for j < plen && src[p+j] == src[q+j] {
				j++
			}

			if j == plen {
				diff = false
			}
		}

		if diff == true {
			name++
			q = p
			qlen = plen
		}

		sa[m+(p>>1)] = name
	}

	return name
}

func induceSuffixArray(src []int, sa []int, ptrBuf1 *[]int, ptrBuf2 *[]int, n int, k int) {
	buf1 := *ptrBuf1
	buf2 := *ptrBuf2

	// compute sal
	if ptrBuf1 == ptrBuf2 {
		getCounts(src, buf1, n, k)
	}

	// find starts of buckets
	getBuckets(buf1, buf2, k, false)

	j := n - 1
	c1 := src[j]
	b := buf2[c1]

	if j > 0 && src[j-1] < c1 {
		sa[b] = ^j
	} else {
		sa[b] = j
	}

	b++

	for i := 0; i < n; i++ {
		j = sa[i]
		sa[i] = ^j

		if j > 0 {
			j--
			c0 := src[j]

			if c0 != c1 {
				buf2[c1] = b
				c1 = c0
				b = buf2[c1]
			}

			if j > 0 && src[j-1] < c1 {
				sa[b] = ^j
			} else {
				sa[b] = j
			}

			b++
		}
	}

	// compute sas
	if ptrBuf1 == ptrBuf2 {
		getCounts(src, buf1, n, k)
	}

	// find ends of buckets
	getBuckets(buf1, buf2, k, true)
	c1 = 0
	b = buf2[c1]

	for i := n - 1; i >= 0; i-- {
		j = sa[i]

		if j > 0 {
			j--
			c0 := src[j]

			if c0 != c1 {
				buf2[c1] = b
				c1 = c0
				b = buf2[c1]
			}

			b--

			if j == 0 || src[j-1] > c1 {
				sa[b] = ^j
			} else {
				sa[b] = j
			}
		} else {
			sa[i] = ^j
		}
	}
}

func computeBWT(data []int, sa []int, ptrBuf1 *[]int, ptrBuf2 *[]int, n int, k int) int {
	buf1 := *ptrBuf1
	buf2 := *ptrBuf2

	// compute sal
	if ptrBuf1 == ptrBuf2 {
		getCounts(data, buf1, n, k)
	}

	// find starts of buckets
	getBuckets(buf1, buf2, k, false)
	j := n - 1
	c1 := data[j]
	b := buf2[c1]

	if j > 0 && data[j-1] < c1 {
		sa[b] = ^j
	} else {
		sa[b] = j
	}

	b++

	for i := 0; i < n; i++ {
		j = sa[i]

		if j > 0 {
			j--
			c0 := data[j]
			sa[i] = ^c0

			if c0 != c1 {
				buf2[c1] = b
				c1 = c0
				b = buf2[c1]
			}

			if j > 0 && data[j-1] < c1 {
				sa[b] = ^j
			} else {
				sa[b] = j
			}

			b++
		} else if j != 0 {
			sa[i] = ^j
		}
	}

	// compute sas
	if ptrBuf1 == ptrBuf2 {
		getCounts(data, buf1, n, k)
	}

	// find ends of buckets
	getBuckets(buf1, buf2, k, true)
	c1 = 0
	b = buf2[c1]
	pidx := -1

	for i := n - 1; i >= 0; i-- {
		j = sa[i]

		if j > 0 {
			j--
			c0 := data[j]
			sa[i] = c0

			if c0 != c1 {
				buf2[c1] = b
				c1 = c0
				b = buf2[c1]
			}

			b--

			if j > 0 && data[j-1] > c1 {
				sa[b] = ^data[j-1]
			} else {
				sa[b] = j
			}
		} else if j != 0 {
			sa[i] = ^j
		} else {
			pidx = i
		}
	}

	return pidx
}

// find the suffix array sa of T[0..n-1] in {0..k-1}^n
func computeSuffixArray(data []int, sa []int, fs int, n int, k int, isbwt bool) uint {
	var B, C []int
	var ptrB, ptrC *[]int
	flags := 0

	if k <= 256 {
		C = make([]int, k)
		ptrC = &C

		if k <= fs {
			B = sa[n+fs-k:]
			flags = 1
		} else {
			B = make([]int, k)
			flags = 3
		}

		ptrB = &B

	} else if k <= fs {
		C = sa[n+fs-k:]
		ptrC = &C

		if k <= fs-k {
			B = sa[n+fs-(k+k):]
			ptrB = &B
			flags = 0
		} else if k <= 1024 {
			B = make([]int, k)
			ptrB = &B
			flags = 2
		} else {
			ptrB = ptrC
			B = *ptrB
			flags = 8
		}
	} else {
		B = make([]int, k)
		ptrB = &B
		ptrC = ptrB
		C = *ptrC
		flags = 12
	}

	// stage 1: reduce the problem by at least 1/2, sort all the LMS-substrings
	// find ends of buckets
	getCounts(data, C, n, k)
	getBuckets(C, B, k, true)

	for ii := 0; ii < n; ii++ {
		sa[ii] = 0
	}

	b := -1
	i := n - 1
	j := n
	m := 0
	c0 := data[n-1]
	c1 := c0
	i--

	if i >= 0 {
		c0 = data[i]
	}

	for i >= 0 && c0 >= c1 {
		c1 = c0
		i--

		if i >= 0 {
			c0 = data[i]
		}
	}

	for i >= 0 {
		c1 = c0
		i--

		if i >= 0 {
			c0 = data[i]
		}

		for i >= 0 && c0 <= c1 {
			c1 = c0
			i--

			if i >= 0 {
				c0 = data[i]
			}
		}

		if i >= 0 {
			if b >= 0 {
				sa[b] = j
			}

			B[c1]--
			b = B[c1]
			j = i
			m++
			c1 = c0
			i--

			if i >= 0 {
				c0 = data[i]
			}

			for i >= 0 && c0 >= c1 {
				c1 = c0
				i--

				if i >= 0 {
					c0 = data[i]
				}
			}
		}
	}

	name := 0

	if m > 1 {
		sortLMSSuffixes(data, sa, ptrC, ptrB, n, k)
		name = postProcessLMS(data, sa, n, m)
	} else if m == 1 {
		sa[b] = j + 1
		name = 1
	}

	// stage 2: solve the reduced problem recurse if names are not yet unique
	if name < m {
		newfs := (n + fs) - (m + m)

		if (flags & 13) == 0 {
			if k+name <= newfs {
				newfs -= k
			} else {
				flags |= 8
			}
		}

		j = m + m + newfs - 1

		for ii := m + (n >> 1) - 1; ii >= m; ii-- {
			if sa[ii] != 0 {
				sa[j] = sa[ii] - 1
				j--
			}
		}

		computeSuffixArray(sa[m+newfs:], sa, newfs, m, name, false)

		i = n - 2
		j = m + m - 1
		c1 = data[n-1]
		c0 = data[n-2]

		for i >= 0 && c0 >= c1 {
			c1 = c0
			i--

			if i >= 0 {
				c0 = data[i]
			}
		}

		for i >= 0 {
			c1 = c0
			i--

			if i >= 0 {
				c0 = data[i]
			}

			for i >= 0 && c0 <= c1 {
				c1 = c0
				i--

				if i >= 0 {
					c0 = data[i]
				}
			}

			if i >= 0 {
				sa[j] = i + 1
				j--
				c1 = c0
				i--

				if i >= 0 {
					c0 = data[i]
				}

				for i >= 0 && c0 >= c1 {
					c1 = c0
					i--

					if i >= 0 {
						c0 = data[i]
					}
				}
			}
		}

		for ii := 0; ii < m; ii++ {
			sa[ii] = sa[m+sa[ii]]
		}

		if (flags & 4) != 0 {
			B = make([]int, k)
			ptrB = &B
			ptrC = ptrB
			C = *ptrC
		} else if (flags & 2) != 0 {
			B = make([]int, k)
			ptrB = &B
		}
	}

	// stage 3: induce the result for the original problem
	if (flags & 8) != 0 {
		getCounts(data, C, n, k)
	}

	// put all left-most S characters into their buckets
	if m > 1 {
		// find ends of buckets
		getBuckets(C, B, k, true)
		i = m - 1
		j = n
		p := sa[m-1]
		c1 = data[p]

		for true {
			c0 = c1
			q := B[c0]

			for q < j {
				j--
				sa[j] = 0
			}

			firstIter := true

			for firstIter == true || c1 == c0 {
				j--
				sa[j] = p
				i--

				if i < 0 {
					break
				}

				p = sa[i]
				c1 = data[p]
				firstIter = false
			}

			if i < 0 {
				break
			}
		}

		for j > 0 {
			j--
			sa[j] = 0
		}
	}

	pidx := 0

	if isbwt == false {
		induceSuffixArray(data, sa, ptrC, ptrB, n, k)
	} else {
		pidx = computeBWT(data, sa, ptrC, ptrB, n, k)
	}

	return uint(pidx)
}
