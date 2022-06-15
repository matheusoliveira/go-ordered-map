# Benchmarks

This is the result of a series of benchmarks on each implementation to validate the assumptions
about design decision of each. Bellow is a formatted version of the results with my own
conclusions, you can see the raw results at [bench.txt](bench.txt) file.

To run the benchmarks your self, just do

```sh
make bench
```

You can generate the formatted output as bellow with [utilities/benchtable.go], the command is provided
in the Makefile as well (must run after generating the `bench.txt` with previous command):

```sh
make doc-bench
```


## Benchmark Iteration

Just put many values in the map, outside of benchmark, and then iterate through the map to
check time taken for full iteration.

| Implemenation | Nruns |      ns/op | B/op | allocs/op | % perf relative |
| ------------- | ----: | ---------: | ---: | --------: | --------------: |
| map           | 3,936 |  1,399,521 |    0 |         0 |        baseline |
| Builtin       |   303 | 19,250,757 |  201 |         4 |        -92.73 % |
| Simple        |   500 | 12,606,482 |   16 |         1 |        -88.90 % |
| Linked        | 7,786 |    848,627 |   16 |         1 |         64.92 % |
| LinkedHash    | 7,083 |    844,225 |   16 |         1 |         65.78 % |
| Sync          | 2,134 |  2,827,290 |   48 |         2 |        -50.50 % |

Conclusion: omap implementations using linked list (Linked and LinkedHash) are faster to iterate
than builtin map, since they have a data struct well design and optimized for that.


## Benchmark MarshalJSON

Calls json.Marshal to convert a single map of (string, int) to JSON.

| Implemenation | Nruns |      ns/op |      B/op | allocs/op | % perf relative |
| ------------- | ----: | ---------: | --------: | --------: | --------------: |
| map           | 1,164 |  5,449,427 | 1,010,914 |    20,006 |        baseline |
| Builtin       |   962 |  6,347,832 | 1,134,340 |    20,010 |        -14.15 % |
| Simple        | 1,009 |  6,163,324 |   857,160 |    39,762 |        -11.58 % |
| Linked        | 1,328 |  4,773,561 |   860,148 |    39,761 |         14.16 % |
| LinkedHash    | 1,280 |  5,021,098 |   865,523 |    39,762 |          8.53 % |
| Sync          | 1,123 |  5,801,274 |   990,457 |    39,764 |         -6.06 % |

Conclusion: the results here are similar to BenchmarkIteration, since all cases have to iterate
while building the final JSON.


## Benchmark ShortStrKeysPut

Put values into the map with a short key length, pre-generating the keys before the benchmarks,
so key size is not accounted in memory.

| Implemenation | Nruns |      ns/op |       B/op | allocs/op | % perf relative |
| ------------- | ----: | ---------: | ---------: | --------: | --------------: |
| map           |   219 | 29,021,513 |  8,088,043 |     3,999 |        baseline |
| Builtin       |   214 | 27,240,396 |  8,087,120 |     3,997 |          6.54 % |
| Simple        |   100 | 52,066,508 | 17,028,371 |     4,018 |        -44.26 % |
| Linked        |   132 | 55,644,086 | 12,887,104 |   103,997 |        -47.84 % |
| LinkedHash    |    78 | 65,659,897 | 16,530,445 |   404,046 |        -55.80 % |
| Sync          |    96 | 57,953,905 | 12,888,188 |   104,003 |        -49.92 % |

Conclusion: since all implementations of omap have to build a separate data structure on Put, it
is expected that they are slower than builtin map, the trade-off seems acceptable if you you
need to iterate (or serialize) the map or if have few keys.


## Benchmark LargeStrKeysPut

Put large string keys in map of int value, pre-generating the keys before the benchmarks,
so key size is not accounted in memory.

| Implemenation | Nruns |         ns/op |       B/op | allocs/op | % perf relative |
| ------------- | ----: | ------------: | ---------: | --------: | --------------: |
| map           |     6 |   915,548,530 |  8,090,290 |     4,010 |        baseline |
| Builtin       |     6 |   933,213,966 |  8,093,778 |     4,029 |         -1.89 % |
| Simple        |     5 | 1,148,190,670 | 17,030,152 |     4,027 |        -20.26 % |
| Linked        |     5 | 1,033,555,280 | 12,891,305 |   104,017 |        -11.42 % |
| LinkedHash    |     6 | 1,009,584,296 | 16,533,028 |   404,057 |         -9.31 % |
| Sync          |     5 | 1,106,828,341 | 12,884,448 |   103,985 |        -17.28 % |

Conclusion: the trade-off here is very similar to BenchmarkShortStrKeysPut, with the advantage
that using a large key actually improve the relative performance, compared to short key.


## Benchmark LargeStrKeysPutGen

Put large string keys in map of int value, but unlike BenchmarkShortStrKeysPut, this benchmark
generates the key inside the benchmark, so both key generation time and key memory is accounted
in the result.

| Implemenation | Nruns |         ns/op |        B/op | allocs/op | % perf relative |
| ------------- | ----: | ------------: | ----------: | --------: | --------------: |
| map           |    33 |   172,837,803 | 401,134,506 |    40,082 |        baseline |
| Builtin       |    30 |   171,893,459 | 401,133,821 |    40,080 |          0.55 % |
| Simple        |    30 |   205,032,312 | 401,819,511 |    40,101 |        -15.70 % |
| Linked        |    28 |   179,232,175 | 401,614,090 |    50,081 |         -3.57 % |
| LinkedHash    |    28 |   183,706,741 | 401,999,293 |    80,112 |         -5.92 % |
| Sync          |    32 |   181,802,519 | 401,614,492 |    50,084 |         -4.93 % |

Conclusion: when the time of large keys generation is accounted in the benchmark, the relative
performance loss compared to BenchmarkLargeStrKeysPut is actually better.


## Benchmark LargeStrKeysGet

Generate a map of large string keys, same as BenchmarkShortStrKeysPut, and then run the
benchmark only to get the values of a random key. All sub-benchmarks use same random
seed.

| Implemenation |     Nruns |         ns/op |        B/op | allocs/op | % perf relative |
| ------------- | --------: | ------------: | ----------: | --------: | --------------: |
| map           | 1,470,026 |         3,609 |           0 |         0 |        baseline |
| Builtin       | 1,769,174 |         3,978 |           0 |         0 |         -9.28 % |
| Simple        | 1,911,117 |         3,221 |           0 |         0 |         12.05 % |
| Linked        | 1,826,122 |         3,221 |           0 |         0 |         12.05 % |
| LinkedHash    | 1,000,000 |         5,294 |          16 |         1 |        -31.83 % |
| Sync          | 1,747,059 |         3,164 |           0 |         0 |         14.06 % |

Conclusion: except for LinkedHash, the implementations basically map the Get operation to a
builtin map, so it is expected that the difference is minor. LinkedHash is more complex, so
it is expected to be slower. All good here.


## Benchmark LargeStrKeysIterate

Generate a map of large string keys, same as BenchmarkShortStrKeysPut, and then run the
benchmark to iterate over all key/value pairs.

| Implemenation |     Nruns |         ns/op |        B/op | allocs/op | % perf relative |
| ------------- | --------: | ------------: | ----------: | --------: | --------------: |
| Builtin       |       345 |    17,452,293 |         200 |         4 |        baseline |
| Simple        |        31 |   200,420,724 |          16 |         1 |        -91.29 % |
| Linked        |    17,079 |       332,097 |          16 |         1 |       5155.18 % |
| LinkedHash    |    21,812 |       265,814 |          16 |         1 |       6465.60 % |
| Sync          |     2,899 |     1,939,563 |          48 |         2 |        799.81 % |

Conclusion: the performance iteration with large keys is even better than short keys.


## Benchmark LargeStrKeysPutGet

Generate a map of large strings keys and int value, and get all values one by one.

| Implemenation |     Nruns |         ns/op |        B/op | allocs/op | % perf relative |
| ------------- | --------: | ------------: | ----------: | --------: | --------------: |
| map           |         9 |   597,743,002 |   8,089,458 |     4,006 |        baseline |
| Builtin       |         8 |   704,752,021 |   8,082,226 |     3,973 |        -15.18 % |
| Simple        |         7 |   813,976,451 |  17,035,262 |     4,051 |        -26.57 % |
| Linked        |         7 |   817,070,427 |  12,883,924 |   103,981 |        -26.84 % |
| LinkedHash    |         6 |   925,364,101 |  18,131,789 |   504,052 |        -35.40 % |
| Sync          |         7 |   786,297,455 |  12,886,646 |   103,995 |        -23.98 % |


## Benchmark LargeObjectKey

Generate a map of large strings keys and int value, and get all values one by one.

| Implemenation |     Nruns |         ns/op |          B/op | allocs/op | % perf relative |
| ------------- | --------: | ------------: | ------------: | --------: | --------------: |
| map           |        36 |   157,263,414 |   410,275,740 |    10,211 |        baseline |
| Builtin       |        24 |   241,444,763 |   410,358,049 |    10,217 |        -34.87 % |
| Simple        |         9 |   612,520,670 | 1,935,593,816 |    10,237 |        -74.33 % |
| Linked        |        16 |   340,107,150 |   819,876,043 |    20,215 |        -53.76 % |
| LinkedHash    |        28 |   198,635,101 |   820,759,129 |    40,308 |        -20.83 % |
| Sync          |        14 |   381,237,194 |   819,876,630 |    20,221 |        -58.75 % |

Conclusion: this test is designed specifically for LinkedHash implementation, and is actually
the only use-case where this implementation is a good fit, and as expected it is the fastest
of omap implementations, although still slower than builtin map. Albeit, it seems a very
specific and unusual use case.


