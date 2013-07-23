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

import (
	"kanzi"
)

const (
	W4_0  = 64
	W4_4  = 83
	W4_8  = 64
	W4_12 = 36
	W4_1  = 64
	W4_5  = 36
	W4_9  = -64
	W4_13 = -83

	MAX_VAL4 = 1 << 16
	MIN_VAL4 = -(MAX_VAL4 + 1)
)

type DCT4 struct {
	fShift uint  // default 8
	iShift uint  // default 20
	data   []int // int[16]
}

func NewDCT4() (*DCT4, error) {
	this := new(DCT4)
	this.fShift = 8
	this.iShift = 20
	this.data = make([]int, 16)
	return this, nil
}

func (this *DCT4) Forward(src, dst []int) (uint, uint, error) {
	computeForward4(src, this.data, 4)
	computeForward4(this.data, dst, this.fShift-4)
	return 16, 16, nil
}

func computeForward4(input []int, output []int, shift uint) {
	round := (1 << shift) >> 1

	x0 := input[0]
	x1 := input[1]
	x2 := input[2]
	x3 := input[3]
	x4 := input[4]
	x5 := input[5]
	x6 := input[6]
	x7 := input[7]
	x8 := input[8]
	x9 := input[9]
	x10 := input[10]
	x11 := input[11]
	x12 := input[12]
	x13 := input[13]
	x14 := input[14]
	x15 := input[15]

	a0 := x0 + x3
	a1 := x1 + x2
	a2 := x0 - x3
	a3 := x1 - x2
	a4 := x4 + x7
	a5 := x5 + x6
	a6 := x4 - x7
	a7 := x5 - x6
	a8 := x8 + x11
	a9 := x9 + x10
	a10 := x8 - x11
	a11 := x9 - x10
	a12 := x12 + x15
	a13 := x13 + x14
	a14 := x12 - x15
	a15 := x13 - x14

	output[0] = ((W4_0 * a0) + (W4_1 * a1) + round) >> shift
	output[1] = ((W4_0 * a4) + (W4_1 * a5) + round) >> shift
	output[2] = ((W4_0 * a8) + (W4_1 * a9) + round) >> shift
	output[3] = ((W4_0 * a12) + (W4_1 * a13) + round) >> shift
	output[4] = ((W4_4 * a2) + (W4_5 * a3) + round) >> shift
	output[5] = ((W4_4 * a6) + (W4_5 * a7) + round) >> shift
	output[6] = ((W4_4 * a10) + (W4_5 * a11) + round) >> shift
	output[7] = ((W4_4 * a14) + (W4_5 * a15) + round) >> shift
	output[8] = ((W4_8 * a0) + (W4_9 * a1) + round) >> shift
	output[9] = ((W4_8 * a4) + (W4_9 * a5) + round) >> shift
	output[10] = ((W4_8 * a8) + (W4_9 * a9) + round) >> shift
	output[11] = ((W4_8 * a12) + (W4_9 * a13) + round) >> shift
	output[12] = ((W4_12 * a2) + (W4_13 * a3) + round) >> shift
	output[13] = ((W4_12 * a6) + (W4_13 * a7) + round) >> shift
	output[14] = ((W4_12 * a10) + (W4_13 * a11) + round) >> shift
	output[15] = ((W4_12 * a14) + (W4_13 * a15) + round) >> shift
}

func (this *DCT4) Inverse(src, dst []int) (uint, uint, error) {
	computeInverse4(src, this.data, 10)
	computeInverse4(this.data, dst, this.iShift-10)
	return 16, 16, nil
}

func computeInverse4(input, output []int, shift uint) {
	round := (1 << shift) >> 1

	x0 := input[0]
	x1 := input[1]
	x2 := input[2]
	x3 := input[3]
	x4 := input[4]
	x5 := input[5]
	x6 := input[6]
	x7 := input[7]
	x8 := input[8]
	x9 := input[9]
	x10 := input[10]
	x11 := input[11]
	x12 := input[12]
	x13 := input[13]
	x14 := input[14]
	x15 := input[15]

	a0 := (W4_4 * x4) + (W4_12 * x12)
	a1 := (W4_5 * x4) + (W4_13 * x12)
	a2 := (W4_0 * x0) + (W4_8 * x8)
	a3 := (W4_1 * x0) + (W4_9 * x8)
	a4 := (W4_4 * x5) + (W4_12 * x13)
	a5 := (W4_5 * x5) + (W4_13 * x13)
	a6 := (W4_0 * x1) + (W4_8 * x9)
	a7 := (W4_1 * x1) + (W4_9 * x9)
	a8 := (W4_4 * x6) + (W4_12 * x14)
	a9 := (W4_5 * x6) + (W4_13 * x14)
	a10 := (W4_0 * x2) + (W4_8 * x10)
	a11 := (W4_1 * x2) + (W4_9 * x10)
	a12 := (W4_4 * x7) + (W4_12 * x15)
	a13 := (W4_5 * x7) + (W4_13 * x15)
	a14 := (W4_0 * x3) + (W4_8 * x11)
	a15 := (W4_1 * x3) + (W4_9 * x11)

	b0 := (a2 + a0 + round) >> shift
	b1 := (a3 + a1 + round) >> shift
	b2 := (a3 - a1 + round) >> shift
	b3 := (a2 - a0 + round) >> shift
	b4 := (a6 + a4 + round) >> shift
	b5 := (a7 + a5 + round) >> shift
	b6 := (a7 - a5 + round) >> shift
	b7 := (a6 - a4 + round) >> shift
	b8 := (a10 + a8 + round) >> shift
	b9 := (a11 + a9 + round) >> shift
	b10 := (a11 - a9 + round) >> shift
	b11 := (a10 - a8 + round) >> shift
	b12 := (a14 + a12 + round) >> shift
	b13 := (a15 + a13 + round) >> shift
	b14 := (a15 - a13 + round) >> shift
	b15 := (a14 - a12 + round) >> shift

	output[0] = kanzi.Clamp(b0, MIN_VAL4, MAX_VAL4)
	output[1] = kanzi.Clamp(b1, MIN_VAL4, MAX_VAL4)
	output[2] = kanzi.Clamp(b2, MIN_VAL4, MAX_VAL4)
	output[3] = kanzi.Clamp(b3, MIN_VAL4, MAX_VAL4)
	output[4] = kanzi.Clamp(b4, MIN_VAL4, MAX_VAL4)
	output[5] = kanzi.Clamp(b5, MIN_VAL4, MAX_VAL4)
	output[6] = kanzi.Clamp(b6, MIN_VAL4, MAX_VAL4)
	output[7] = kanzi.Clamp(b7, MIN_VAL4, MAX_VAL4)
	output[8] = kanzi.Clamp(b8, MIN_VAL4, MAX_VAL4)
	output[9] = kanzi.Clamp(b9, MIN_VAL4, MAX_VAL4)
	output[10] = kanzi.Clamp(b10, MIN_VAL4, MAX_VAL4)
	output[11] = kanzi.Clamp(b11, MIN_VAL4, MAX_VAL4)
	output[12] = kanzi.Clamp(b12, MIN_VAL4, MAX_VAL4)
	output[13] = kanzi.Clamp(b13, MIN_VAL4, MAX_VAL4)
	output[14] = kanzi.Clamp(b14, MIN_VAL4, MAX_VAL4)
	output[15] = kanzi.Clamp(b15, MIN_VAL4, MAX_VAL4)
}
