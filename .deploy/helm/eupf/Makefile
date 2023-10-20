HELM_REGISTRY?=oci://ghcr.io/edgecomllc/charts
HELM_PACKAGE_NAME?=$(notdir $(CURDIR))
HELM_PACKAGE_VERSION?=$(shell cat Chart.yaml | grep ^version | awk -F ": " '{ print $$2}')

# create PAT https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens
# login with PAT: helm registry login -u username ghcr.io

package:
	helm package .

push:
	helm push $(HELM_PACKAGE_NAME)-$(HELM_PACKAGE_VERSION).tgz $(HELM_REGISTRY)
