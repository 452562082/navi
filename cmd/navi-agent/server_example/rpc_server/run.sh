#!/bin/bash
g++ -Wall -O3 -g -DHAVE_NETINET_IN_H -I. -I/usr/local/include/  -L/usr/local/lib/ *.cpp  -o rpc-server -lthriftnb -levent -lthrift
#g++ -o asv-rpc *.cpp -DHAVE_NETINET_IN_H -I. -I/usr/local/include/ -I/kaldi/kaldi/kaldi/src/kvp/ -L/usr/local/lib/ -lthriftnb -lthrift -L/kaldi/kaldi/kaldi/src/lib/ -lkvp-asv -lfaiss -L/usr/OpenBLAS/lib -lopenblas -levent -lrt -fPIC -m64 -Wall -g -O3 -mavx -msse4 -mpopcnt -Wno-sign-compare -std=c++11
#g++ -g -DHAVE_NETINET_IN_H -I. -I/usr/local/include/ -L/usr/local/lib/  kvpClient.cc -o rpc-client -lthriftnb -levent -lthrift -lrt
#export LD_LIBRARY_PATH=/kaldi/kaldi/kaldi/src/lib:/usr/local/lib
#./asv-rpc