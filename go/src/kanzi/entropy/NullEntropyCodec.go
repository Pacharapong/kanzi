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
	"kanzi"
)

type NullEntropyEncoder struct {
	bitstream kanzi.OutputBitStream
}

func NewNullEntropyEncoder(bs kanzi.OutputBitStream) (*NullEntropyEncoder, error) {
	this := new(NullEntropyEncoder)
	this.bitstream = bs
	return this, nil
}

func (this *NullEntropyEncoder) Encode(block []byte) (int, error) {
	return EntropyEncodeArray(this, block)
}

func (this *NullEntropyEncoder) EncodeByte(val byte) error {
	_, err := this.bitstream.WriteBits(uint64(val), 8)
	return err
}

func (this *NullEntropyEncoder) BitStream() kanzi.OutputBitStream {
	return this.bitstream
}

func (this *NullEntropyEncoder) Dispose() {
}

type NullEntropyDecoder struct {
	bitstream kanzi.InputBitStream
}

func NewNullEntropyDecoder(bs kanzi.InputBitStream) (*NullEntropyDecoder, error) {
	this := new(NullEntropyDecoder)
	this.bitstream = bs
	return this, nil
}

func (this *NullEntropyDecoder) Decode(block []byte) (int, error) {
	return EntropyDecodeArray(this, block)
}

func (this *NullEntropyDecoder) DecodeByte() (byte, error) {
	r, err := this.bitstream.ReadBits(8)
	return byte(r), err
}

func (this *NullEntropyDecoder) BitStream() kanzi.InputBitStream {
	return this.bitstream
}

func (this *NullEntropyDecoder) Dispose() {
}
