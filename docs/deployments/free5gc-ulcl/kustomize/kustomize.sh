#!/bin/sh

cat > ./kustomize/all.yaml
kubectl kustomize ./kustomize
rm ./kustomize/all.yaml