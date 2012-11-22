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

package main

import (
	"kanzi/function"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	fmt.Printf("TestZLT\n")

	for ii := 0; ii < 20; ii++ {
		fmt.Printf("\nCorrectness test %v\n", ii)

		arr := make([]int, 64)

		for i := range arr {
			val := rand.Intn(100) - 16

			if val >= 33 {
				val = 0
			}

			arr[i] = val
		}

		size := len(arr)
		input := make([]byte, size)
		output := make([]byte, size)
		reverse := make([]byte, size)

		for i := range output {
			output[i] = 0x55
		}

		for i := range arr {
			input[i] = byte(arr[i])
		}

		zlt, _ := function.NewZLT(0)
		fmt.Printf("\nOriginal: ")

		for i := range input {
			if i%100 == 0 {
				fmt.Printf("\n")
			}

			fmt.Printf("%v ", input[i])
		}

		fmt.Printf("\nCoded: ")
		srcIdx, dstIdx, _ := zlt.Forward(input, output)

		for i := uint(0); i < dstIdx; i++ {
			if i%100 == 0 {
				fmt.Printf("\n")
			}

			fmt.Printf("%v ", output[i])
		}

		fmt.Printf(" (Compression ratio: %v%%)", dstIdx*100/srcIdx)
		zlt, _ = function.NewZLT(dstIdx) // Required to reset internal attributes
		zlt.Inverse(output, reverse)
		fmt.Printf("\nDecoded: ")
		ok := true

		for i := range input {
			if i%100 == 0 {
				fmt.Printf("\n")
			}

			fmt.Printf("%v ", reverse[i])

			if reverse[i] != input[i] {
				ok = false
			}
		}

		if ok == true {
			fmt.Printf("\nIdentical\n")
		} else {
			fmt.Printf("\nDifferent\n")
			os.Exit(1)
		}
	}

	iter := 50000
	fmt.Printf("\n\nSpeed test\n")
	fmt.Printf("Iterations: %v\n", iter)

	for jj := 0; jj < 3; jj++ {
		input := make([]byte, 30000)
		output := make([]byte, len(input))
		reverse := make([]byte, len(input))

		// Generate random data with runs
		n := 0
		delta1 := int64(0)
		delta2 := int64(0)

		for n < len(input) {
			val := byte(rand.Intn(3))
			input[n] = val
			n++
			run := rand.Intn(255)
			run -= 200
			run--

			for run > 0 && n < len(input) {
				input[n] = val
				n++
				run--
			}
		}

		for ii := 0; ii < iter; ii++ {
			zlt, _ := function.NewZLT(0)
			before := time.Now()
			zlt.Forward(input, output)
			after := time.Now()
			delta1 += after.Sub(before).Nanoseconds()
		}

		for ii := 0; ii < iter; ii++ {
			zlt, _ := function.NewZLT(0)
			before := time.Now()
			zlt.Inverse(output, reverse)
			after := time.Now()
			delta2 += after.Sub(before).Nanoseconds()
		}

		idx := -1

		// Sanity check
		for i := range input {
			if input[i] != reverse[i] {
				idx = i
				break
			}
		}

		if idx >= 0 {
			fmt.Printf("Failure at index %v (%v <-> %v)\n", idx, input[idx], reverse[idx])
			os.Exit(1)
		}

		fmt.Printf("\nZLT encoding [ms]: %v", delta1/1000000)
		fmt.Printf("\nZLT decoding [ms]: %v", delta2/1000000)
	}
}
