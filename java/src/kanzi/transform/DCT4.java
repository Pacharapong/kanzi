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

package kanzi.transform;

import kanzi.IntTransform;


// Implementation of Discrete Cosine Transform of dimension 4 
// Due to rounding errors, the recosntruction may not be perfect
public final class DCT4 implements IntTransform
{
    // Weights
    private final static int W0  = 64;
    private final static int W4  = 83;
    private final static int W8  = 64;
    private final static int W12 = 36;
    private final static int W1  = 64;
    private final static int W5  = 36;
    private final static int W9  = -64;
    private final static int W13 = -83;
    
    private static final int MAX_VAL = 1<<16;
    private static final int MIN_VAL = -(MAX_VAL+1);
            
    private final int fShift;
    private final int iShift;
    private final int[] data;
 

    public DCT4()
    {
       this.fShift = 8;
       this.iShift = 20;
       this.data = new int[16];
    }
    

    public int[] forward(int[] block)
    {
       return this.forward(block, 0);
    }


    @Override
    public int[] forward(int[] block, int blkptr)
    {
       this.computeForward(block, blkptr, this.data, 0, 4);
       this.computeForward(this.data, 0, block, blkptr, this.fShift-4);
       return block;
    }
    
    
    private int[] computeForward(int[] input, int iIdx, int[] output, int oIdx, int shift)
    {
       final int round = (shift == 0) ? 0 : 1 << (shift - 1);
       
       final int x0  = input[iIdx];
       final int x1  = input[iIdx+1];
       final int x2  = input[iIdx+2];
       final int x3  = input[iIdx+3];
       final int x4  = input[iIdx+4];
       final int x5  = input[iIdx+5];
       final int x6  = input[iIdx+6];
       final int x7  = input[iIdx+7];
       final int x8  = input[iIdx+8];
       final int x9  = input[iIdx+9];
       final int x10 = input[iIdx+10];
       final int x11 = input[iIdx+11];
       final int x12 = input[iIdx+12];
       final int x13 = input[iIdx+13];
       final int x14 = input[iIdx+14];
       final int x15 = input[iIdx+15];
       
       final int a0  = x0 + x3;
       final int a1  = x1 + x2;
       final int a2  = x0 - x3;
       final int a3  = x1 - x2;
       final int a4  = x4 + x7;
       final int a5  = x5 + x6;
       final int a6  = x4 - x7;
       final int a7  = x5 - x6;
       final int a8  = x8 + x11;
       final int a9  = x9 + x10;
       final int a10 = x8 - x11;
       final int a11 = x9 - x10;
       final int a12 = x12 + x15;
       final int a13 = x13 + x14;
       final int a14 = x12 - x15;
       final int a15 = x13 - x14;
    
       output[oIdx]    = ((W0  * a0)  + (W1  * a1)  + round) >> shift;
       output[oIdx+1]  = ((W0  * a4)  + (W1  * a5)  + round) >> shift;
       output[oIdx+2]  = ((W0  * a8)  + (W1  * a9)  + round) >> shift;
       output[oIdx+3]  = ((W0  * a12) + (W1  * a13) + round) >> shift;
       output[oIdx+4]  = ((W4  * a2)  + (W5  * a3)  + round) >> shift;
       output[oIdx+5]  = ((W4  * a6)  + (W5  * a7)  + round) >> shift;
       output[oIdx+6]  = ((W4  * a10) + (W5  * a11) + round) >> shift;
       output[oIdx+7]  = ((W4  * a14) + (W5  * a15) + round) >> shift;
       output[oIdx+8]  = ((W8  * a0)  + (W9  * a1)  + round) >> shift;
       output[oIdx+9]  = ((W8  * a4)  + (W9  * a5)  + round) >> shift;
       output[oIdx+10] = ((W8  * a8)  + (W9  * a9)  + round) >> shift;
       output[oIdx+11] = ((W8  * a12) + (W9  * a13) + round) >> shift;
       output[oIdx+12] = ((W12 * a2)  + (W13 * a3)  + round) >> shift;
       output[oIdx+13] = ((W12 * a6)  + (W13 * a7)  + round) >> shift;
       output[oIdx+14] = ((W12 * a10) + (W13 * a11) + round) >> shift;
       output[oIdx+15] = ((W12 * a14) + (W13 * a15) + round) >> shift;
       
       return output;
    }


    public int[] inverse(int[] block)
    {
        return this.inverse(block, 0);
    }


    @Override
    public int[] inverse(int[] block, int blkptr)
    {
       this.computeInverse(block, blkptr, this.data, 0, 10);
       this.computeInverse(this.data, 0, block, blkptr, this.iShift-10);
       return block;
    }
    
    
    private int[] computeInverse(int[] input, int iIdx, int[] output, int oIdx, int shift)
    {
       final int round = (shift == 0) ? 0 : 1 << (shift - 1);

       final int x0  = input[iIdx];
       final int x1  = input[iIdx+1];
       final int x2  = input[iIdx+2];
       final int x3  = input[iIdx+3];
       final int x4  = input[iIdx+4];
       final int x5  = input[iIdx+5];
       final int x6  = input[iIdx+6];
       final int x7  = input[iIdx+7];
       final int x8  = input[iIdx+8];
       final int x9  = input[iIdx+9];
       final int x10 = input[iIdx+10];
       final int x11 = input[iIdx+11];
       final int x12 = input[iIdx+12];
       final int x13 = input[iIdx+13];
       final int x14 = input[iIdx+14];
       final int x15 = input[iIdx+15];

       final int a0  = (W4  * x4) + (W12 * x12);
       final int a1  = (W5  * x4) + (W13 * x12);
       final int a2  = (W0  * x0) + (W8  * x8);
       final int a3  = (W1  * x0) + (W9  * x8);
       final int a4  = (W4  * x5) + (W12 * x13);
       final int a5  = (W5  * x5) + (W13 * x13);
       final int a6  = (W0  * x1) + (W8  * x9);
       final int a7  = (W1  * x1) + (W9  * x9);
       final int a8  = (W4  * x6) + (W12 * x14);
       final int a9  = (W5  * x6) + (W13 * x14);
       final int a10 = (W0  * x2) + (W8  * x10);
       final int a11 = (W1  * x2) + (W9  * x10);
       final int a12 = (W4  * x7) + (W12 * x15);
       final int a13 = (W5  * x7) + (W13 * x15);
       final int a14 = (W0  * x3) + (W8  * x11);
       final int a15 = (W1  * x3) + (W9  * x11);
       
       final int b0  = (a2  + a0  + round) >> shift;
       final int b1  = (a3  + a1  + round) >> shift;
       final int b2  = (a3  - a1  + round) >> shift;
       final int b3  = (a2  - a0  + round) >> shift;
       final int b4  = (a6  + a4  + round) >> shift;
       final int b5  = (a7  + a5  + round) >> shift;
       final int b6  = (a7  - a5  + round) >> shift;
       final int b7  = (a6  - a4  + round) >> shift;
       final int b8  = (a10 + a8  + round) >> shift;
       final int b9  = (a11 + a9  + round) >> shift;
       final int b10 = (a11 - a9  + round) >> shift;
       final int b11 = (a10 - a8  + round) >> shift;
       final int b12 = (a14 + a12 + round) >> shift;
       final int b13 = (a15 + a13 + round) >> shift;
       final int b14 = (a15 - a13 + round) >> shift;
       final int b15 = (a14 - a12 + round) >> shift;
       
       output[oIdx]    = (b0  >= MAX_VAL) ? MAX_VAL : ((b0  <= MIN_VAL) ? MIN_VAL : b0);
       output[oIdx+1]  = (b1  >= MAX_VAL) ? MAX_VAL : ((b1  <= MIN_VAL) ? MIN_VAL : b1);
       output[oIdx+2]  = (b2  >= MAX_VAL) ? MAX_VAL : ((b2  <= MIN_VAL) ? MIN_VAL : b2);
       output[oIdx+3]  = (b3  >= MAX_VAL) ? MAX_VAL : ((b3  <= MIN_VAL) ? MIN_VAL : b3);
       output[oIdx+4]  = (b4  >= MAX_VAL) ? MAX_VAL : ((b4  <= MIN_VAL) ? MIN_VAL : b4);
       output[oIdx+5]  = (b5  >= MAX_VAL) ? MAX_VAL : ((b5  <= MIN_VAL) ? MIN_VAL : b5);
       output[oIdx+6]  = (b6  >= MAX_VAL) ? MAX_VAL : ((b6  <= MIN_VAL) ? MIN_VAL : b6);
       output[oIdx+7]  = (b7  >= MAX_VAL) ? MAX_VAL : ((b7  <= MIN_VAL) ? MIN_VAL : b7);
       output[oIdx+8]  = (b8  >= MAX_VAL) ? MAX_VAL : ((b8  <= MIN_VAL) ? MIN_VAL : b8);
       output[oIdx+9]  = (b9  >= MAX_VAL) ? MAX_VAL : ((b9  <= MIN_VAL) ? MIN_VAL : b9);
       output[oIdx+10] = (b10 >= MAX_VAL) ? MAX_VAL : ((b10 <= MIN_VAL) ? MIN_VAL : b10);
       output[oIdx+11] = (b11 >= MAX_VAL) ? MAX_VAL : ((b11 <= MIN_VAL) ? MIN_VAL : b11);
       output[oIdx+12] = (b12 >= MAX_VAL) ? MAX_VAL : ((b12 <= MIN_VAL) ? MIN_VAL : b12);
       output[oIdx+13] = (b13 >= MAX_VAL) ? MAX_VAL : ((b13 <= MIN_VAL) ? MIN_VAL : b13);
       output[oIdx+14] = (b14 >= MAX_VAL) ? MAX_VAL : ((b14 <= MIN_VAL) ? MIN_VAL : b14);
       output[oIdx+15] = (b15 >= MAX_VAL) ? MAX_VAL : ((b15 <= MIN_VAL) ? MIN_VAL : b15);
       
       return output;
    }

}