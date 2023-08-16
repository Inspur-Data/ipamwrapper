FROM ubuntu:20.04
WORKDIR /
COPY  bin/ipamwrapper .
COPY  bin/ipamwrapper-cni .
COPY  bin/router .

