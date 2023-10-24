NAMESPACE?=free5gc
HELM_VALUES_DIR?=./values

HELM_ARGS?=--timeout 30s --create-namespace
HELM_ACTION?=upgrade --install

free5gc:
	helm $(HELM_ACTION) \
		free5gc towards5gs/free5gc \
		--version 1.1.7 \
		--values $(HELM_VALUES_DIR)/global.yaml \
		--values $(HELM_VALUES_DIR)/free5gc.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

calico:
	kubectl apply -f $(HELM_VALUES_DIR)/../manifests/calico-pools.yaml
	kubectl apply -f $(HELM_VALUES_DIR)/../manifests/calico-bgp.yaml
	kubectl patch felixconfiguration default --type='merge' -p '{"spec":{"removeExternalRoutes":false}}'

upf:
	helm $(HELM_ACTION) \
		eupf oci://ghcr.io/edgecomllc/charts/eupf \
		--version 0.5.0 \
		--values $(HELM_VALUES_DIR)/eupf.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ue1:
	helm $(HELM_ACTION) \
		ueransim1 towards5gs/ueransim \
		--values $(HELM_VALUES_DIR)/global.yaml \
		--values $(HELM_VALUES_DIR)/ueransim.yaml \
		--version 2.0.17 \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

clean:
	helm delete free5gc -n $(NAMESPACE) || true
	helm delete eupf -n $(NAMESPACE) || true
	helm delete ueransim1 -n $(NAMESPACE) || true
	kubectl delete -f $(HELM_VALUES_DIR)/../manifests/calico-pools.yaml || true
	kubectl delete -f $(HELM_VALUES_DIR)/../manifests/calico-bgp.yaml || true
