FROM ubuntu:20.04
WORKDIR /
COPY  bin/k8-ipam .
COPY  bin/k8-ipam-cni .
