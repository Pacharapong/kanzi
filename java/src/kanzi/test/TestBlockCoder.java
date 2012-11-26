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

package kanzi.test;

import java.util.HashSet;
import java.util.Set;
import kanzi.app.BlockCompressor;
import kanzi.app.BlockDecompressor;


public class TestBlockCoder
{
    public static void main(String[] args)
    {
       BlockCompressor enc = new BlockCompressor(args);
       long status = enc.call();
       
       if (status < 0L)
       {
          System.out.println("Compression failed with status " + status + ", skipping decompression");
          System.exit((int) status);
       }

       Set<String> set = new HashSet<String>();

       for (int i=0; i<args.length; i++)
       {
          if (args[i].startsWith("-input="))
             set.add(args[i]+".knz");
          else if (args[i].equals("-debug"))
             set.add(args[i]);
          else if (args[i].equals("-silent"))
             set.add(args[i]);
          else if (args[i].equals("-overwrite"))
             set.add(args[i]);
       }

       args = (String[]) set.toArray(new String[set.size()]);
       BlockDecompressor dec = new BlockDecompressor(args);
       status = dec.call();

       if (status < 0L)
       {
          System.out.println("Decompression failed with status " + status);
          System.exit((int) status);
       }
    }
    
}