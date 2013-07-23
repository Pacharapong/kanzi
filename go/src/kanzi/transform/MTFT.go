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

const (
	RESET_THRESHOLD = 64
	LIST_LENGTH     = 17
)

type Payload struct {
	previous *Payload
	next     *Payload
	value    byte
}

type MTFT struct {
	size    uint
	lengths []int      // size 16
	buckets []byte     // size 256
	heads   []*Payload // size 16
	anchor  *Payload
}

func NewMTFT(sz uint) (*MTFT, error) {
	this := new(MTFT)
	this.size = sz
	this.heads = make([]*Payload, 16)
	this.lengths = make([]int, 16)
	this.buckets = make([]byte, 256)
	return this, nil
}

func (this *MTFT) Size() uint {
	return this.size
}

func (this *MTFT) SetSize(sz uint) bool {
	this.size = sz
	return true
}

func (this *MTFT) Inverse(src, dst []byte) (uint, uint, error) {
	indexes := this.buckets

	for i := range indexes {
		indexes[i] = byte(i)
	}

	count := int(this.size)

	if count == 0 {
		count = len(src)
	}

	for i := 0; i < count; i++ {
		idx := src[i]

		if idx == 0 {
			dst[i] = indexes[0]
			continue
		}

		value := indexes[idx]
		dst[i] = value

		if idx < 16 {
			for j := int(idx - 1); j >= 0; j-- {
				indexes[j+1] = indexes[j]
			}
		} else {
			copy(indexes[1:], indexes[0:idx])
		}

		indexes[0] = value
	}

	return uint(count), uint(count), nil
}

// Initialize the linked lists: 1 item in bucket 0 and LIST_LENGTH in each other
// Used by forward() only
func (this *MTFT) initLists() {
	array := make([]*Payload, 256)
	array[0] = &Payload{value: 0}
	previous := array[0]
	this.heads[0] = previous
	this.lengths[0] = 1
	this.buckets[0] = 0
	listIdx := byte(0)

	for i := 1; i < 256; i++ {
		array[i] = &Payload{value: byte(i)}

		if (i-1)%LIST_LENGTH == 0 {
			listIdx++
			this.heads[listIdx] = array[i]
			this.lengths[listIdx] = LIST_LENGTH
		}

		this.buckets[i] = listIdx
		previous.next = array[i]
		array[i].previous = previous
		previous = array[i]
	}

	// Create a fake end payload so that every payload in every list has a successor
	this.anchor = &Payload{value: 0}
	previous.next = this.anchor
}

// Recreate one list with 1 item and 15 lists with LIST_LENGTH items
// Update lengths and buckets accordingly.
// Used by forward() only
func (this *MTFT) balanceLists(resetValues bool) {
	this.lengths[0] = 1
	p := this.heads[0].next
	val := byte(0)

	if resetValues == true {
		this.heads[0].value = byte(0)
		this.buckets[0] = 0
	}

	for listIdx := byte(1); listIdx < 16; listIdx++ {
		this.heads[listIdx] = p
		this.lengths[listIdx] = LIST_LENGTH

		for n := 0; n < LIST_LENGTH; n++ {
			if resetValues == true {
				val++
				p.value = val
			}

			this.buckets[int(p.value)] = listIdx
			p = p.next
		}
	}
}

func (this *MTFT) Forward(src, dst []byte) (uint, uint, error) {
	if this.anchor == nil {
		this.initLists()
	} else {
		this.balanceLists(true)
	}

	count := int(this.size)

	if count == 0 {
		count = len(src)
	}

	previous := this.heads[0].value

	for ii := 0; ii < count; ii++ {
		current := src[ii]

		if current == previous {
			dst[ii] = byte(0)
			continue
		}

		// Find list index
		listIdx := int(this.buckets[int(current)])
		p := this.heads[listIdx]
		idx := 0

		for i := 0; i < listIdx; i++ {
			idx += this.lengths[i]
		}

		// Find index in list (less than RESET_THRESHOLD iterations)
		for p.value != current {
			p = p.next
			idx++
		}

		dst[ii] = byte(idx)

		// Unlink (the end anchor ensures p.next != nil)
		p.previous.next = p.next
		p.next.previous = p.previous

		// Add to head of first list
		p.next = this.heads[0]
		p.next.previous = p
		this.heads[0] = p

		// Update list information
		if listIdx != 0 {
			// Update head if needed
			if p == this.heads[listIdx] {
				this.heads[listIdx] = p.previous.next
			}

			this.lengths[listIdx]--
			this.lengths[0]++
			this.buckets[int(current)] = 0

			if this.lengths[0] > RESET_THRESHOLD || this.lengths[listIdx] == 0 {
				this.balanceLists(false)
			}
		}

		previous = current
	}

	return uint(count), uint(count), nil
}
