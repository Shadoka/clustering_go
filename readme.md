**This repository contains a simple implementation of hierarchic clustering.**

It uses parallelism to build all possible combinations of clusters, but in this specific case it does not yield any performance gain compared to the sequential implementation.

Time (without parallelism): 33ms

Time (with parallelism)   : 35ms

I noticed a huge (compared to the total runtime of the program) performance boost when I set the size of the slice in line 137 from 10 to 4950 (the actual amount of possible combinations with 100 cluster).
As soon as I changed that line, the runtime went from 53ms to 35ms.
Although much too oversized for the last passes of clustering, it fits the first pass just right. For this I conclude that reallocating memory is quite expensive and if one wants to write performant Go programms
he needs to find a good tradeoff between setting the size too low (too many reallocations) and too high (too much memory reserved).