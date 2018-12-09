#!/usr/bin/env bash


make IMAGES=cassandra build
img=$(docker images | grep -Eo '^build-[a-z0-9]{8}/cassandra-[a-z0-9]+\s')
docker tag $img localhost:5000/rook/cassandra:master
docker --debug push localhost:5000/rook/cassandra:master
kubectl delete sts rook-cassandra-operator --namespace=rook-cassandra-system
kubectl apply -f cluster/examples/kubernetes/cassandra/operator.yaml
