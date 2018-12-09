#!/usr/bin/env bash

kubectl delete -f cluster/examples/kubernetes/cassandra/cluster.yaml
kubectl delete -f cluster/examples/kubernetes/cassandra/operator.yaml


go test -v -timeout 1800s -count 1 -run CassandraSuite github.com/rook/rook/tests/integration
