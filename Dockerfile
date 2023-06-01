FROM ubuntu:20.04
WORKDIR /
COPY --from=0 bin/k8-ipam .