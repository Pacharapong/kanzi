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

package main

import (
	"fmt"
	"kanzi/bitstream"
	"kanzi/entropy"
	"kanzi/util"
	"math/rand"
	"os"
	"time"
)

func main() {
	fmt.Printf("\nTestPAQCodec")
	TestCorrectness()
	TestSpeed()
}

func TestCorrectness() {
	fmt.Printf("\n\nCorrectness test")

	// Test behavior
	for ii := 1; ii < 20; ii++ {
		fmt.Printf("\nTest %v", ii)
		var values []byte
		rand.Seed(time.Now().UTC().UnixNano())

		if ii == 3 {
			values = []byte{0, 0, 32, 15, -4 & 0xFF, 16, 0, 16, 0, 7, -1 & 0xFF, -4 & 0xFF, -32 & 0xFF, 0, 31, -1 & 0xFF}
		} else if ii == 2 {
			values = []byte{0x3d, 0x4d, 0x54, 0x47, 0x5a, 0x36, 0x39, 0x26, 0x72, 0x6f, 0x6c, 0x65, 0x3d, 0x70, 0x72, 0x65}
		} else if ii == 1 {
			values = []byte{65, 71, 74, 66, 76, 65, 69, 77, 74, 79, 68, 75, 73, 72, 77, 68, 78, 65, 79, 79, 78, 66, 77, 71, 64, 70, 74, 77, 64, 67, 71, 64}
		} else {
			values = make([]byte, 32)

			for i := range values {
				values[i] = byte(64 + 3*ii + rand.Intn(ii+1))
			}
		}

		fmt.Printf("\nOriginal: ")

		for i := range values {
			fmt.Printf("%d ", values[i])
		}

		fmt.Printf("\nEncoded: ")
		buffer := make([]byte, 16384)
		oFile, _ := util.NewByteArrayOutputStream(buffer, true)
		defer oFile.Close()
		obs, _ := bitstream.NewDefaultOutputBitStream(oFile, 16384)
		dbgbs, _ := bitstream.NewDebugOutputBitStream(obs, os.Stdout)
		dbgbs.ShowByte(true)
		dbgbs.Mark(true)
		predictor1, _ := entropy.NewPAQPredictor()
		fpc, _ := entropy.NewBinaryEntropyEncoder(dbgbs, predictor1)

		if _, err := fpc.Encode(values); err != nil {
			fmt.Printf("Error during encoding: %s", err)
			os.Exit(1)
		}

		fpc.Dispose()
		dbgbs.Close()
		println()
		
		iFile, _ := util.NewByteArrayInputStream(buffer, true)
		defer iFile.Close()
		ibs, _ := bitstream.NewDefaultInputBitStream(iFile, 16384)
		dbgbs2, _ := bitstream.NewDebugInputBitStream(ibs, os.Stdout)
		dbgbs2.ShowByte(true)
		//dbgbs2.Mark(true)

		predictor2, _ := entropy.NewPAQPredictor()
		fpd, _ := entropy.NewBinaryEntropyDecoder(dbgbs2, predictor2)

		ok := true
		values2 := make([]byte, len(values))

		if _, err := fpd.Decode(values2); err != nil {
			fmt.Printf("Error during decoding: %s", err)
			os.Exit(1)
		}

		fmt.Printf("\nDecoded: ")

		for i := range values2 {
			fmt.Printf("%v ", values2[i])

			if values[i] != values2[i] {
				ok = false
			}
		}

		if ok == true {
			fmt.Printf("\nIdentical")
		} else {
			fmt.Printf("\n! *** Different *** !")
			os.Exit(1)
		}

		fpd.Dispose()
		dbgbs2.Close()
		fmt.Printf("\n")
	}
}

func TestSpeed() {
	fmt.Printf("\n\nSpeed test\n")
	repeats := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5, 8, 9, 7, 9, 3}

	for jj := 0; jj < 3; jj++ {
		fmt.Printf("\nTest %v\n", jj+1)
		size := 50000
		iter := 1000
		buffer := make([]byte, size*2)
		values1 := make([]byte, size)
		values2 := make([]byte, size)
		delta1 := int64(0)
		delta2 := int64(0)

		for ii := 0; ii < iter; ii++ {
			idx := jj

			for i := 0; i < len(values1); i++ {
				i0 := i

				length := repeats[idx]
				idx = (idx + 1) & 0x0F

				if i0+length >= len(values1) {
					length = 1
				}

				for j := i0; j < i0+length; j++ {
					values1[j] = byte(i0)
					i++
				}
			}

			oFile, _ := util.NewByteArrayOutputStream(buffer, false)
			defer oFile.Close()
			predictor1, _ := entropy.NewPAQPredictor()
			obs, _ := bitstream.NewDefaultOutputBitStream(oFile, uint(size))
			rc, _ := entropy.NewBinaryEntropyEncoder(obs, predictor1)

			// Encode
			before := time.Now()

			if _, err := rc.Encode(values1); err != nil {
				fmt.Printf("An error occured during encoding: %v\n", err)
				os.Exit(1)
			}

			rc.Dispose()
			obs.Close()

			if _, err := obs.Close(); err != nil {
				fmt.Printf("Error during close: %v\n", err)
				os.Exit(1)
			}

			after := time.Now()
			delta1 += after.Sub(before).Nanoseconds()
		}

		for ii := 0; ii < iter; ii++ {
			iFile, _ := util.NewByteArrayInputStream(buffer, false)
			defer iFile.Close()
			predictor2, _ := entropy.NewPAQPredictor()
			ibs, _ := bitstream.NewDefaultInputBitStream(iFile, uint(size))
			rd, _ := entropy.NewBinaryEntropyDecoder(ibs, predictor2)

			// Decode
			before := time.Now()

			if _, err := rd.Decode(values2); err != nil {
				fmt.Printf("An error occured during decoding: %v\n", err)
				os.Exit(1)
			}

			rd.Dispose()
			ibs.Close()

			if _, err := ibs.Close(); err != nil {
				fmt.Printf("Error during close: %v\n", err)
				os.Exit(1)
			}

			after := time.Now()
			delta2 += after.Sub(before).Nanoseconds()

		}

		prod := int64(iter) * int64(size)
		fmt.Printf("Encode [ms]      : %d\n", delta1/1000000)
		fmt.Printf("Throughput [KB/s]: %d\n", prod*1000000/delta1*1000/1024)
		fmt.Printf("Decode [ms]      : %d\n", delta2/1000000)
		fmt.Printf("Throughput [KB/s]: %d\n", prod*1000000/delta2*1000/1024)
	}
}
