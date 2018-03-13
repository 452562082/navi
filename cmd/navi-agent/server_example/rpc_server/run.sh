#!/bin/bash
g++ -Wall -O3 -g -DHAVE_NETINET_IN_H -I. -I/usr/local/include/ -L/usr/local/lib/ *.cpp  -o mytest -lthriftnb -levent -lthrift -lrt -lz
#./asv-rpc