#include <iostream>
#include <iomanip>
#include <string>
#include <map>
#include <random>
#include <cmath>
int main()
{
    std::random_device rd;
    std::mt19937 gen(rd());
    std::normal_distribution<> d(5,2);
    for(int n=0; n<10000; ++n) {
        std::cout<<std::round(d(gen))<<std::endl;
    }
    return 0;
}
