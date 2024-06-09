#!/bin/sh

cat > all.yaml
kubectl kustomize ./kustomize
rm all.yaml