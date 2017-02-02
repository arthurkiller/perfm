# perfm [![Build Status](https://travis-ci.org/arthurkiller/perfm.svg?branch=master)](https://travis-ci.org/arthurkiller/perfm) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/perfm)](https://goreportcard.com/report/github.com/arthurkiller/perfm)
a golang performence testing platform

## Testing Data
The testing data was generate by the CPP random engin , the code is in the normal distribute.
[![pic](demo/screen.png)](github.com/arthurkiller/perfm)

```cpp
#include <iostream>
#include <iomanip>
#include <string>
#include <random>
#include <cmath>
int main()
{
    std::random_device rd;
    std::mt19937 gen(rd());
    std::normal_distribution<> d(5,2);
    for(int n=0; n<100000; ++n) {
        std::cout<<std::round(d(gen))<<std::endl;
    }
    return 0;
}
```

## TODO & Milestone
* version 0.1 
    support the qps and average cost counting
* version 0.2
    add the presur testing feature to support the presure performence testing
