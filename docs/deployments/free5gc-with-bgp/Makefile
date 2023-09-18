NAMESPACE?=free5gc
HELM_VALUES_DIR?=./values

HELM_ARGS?=--timeout 30s --create-namespace
HELM_ACTION?=upgrade --install

free5gc:
	phy_if=$(lshw -C network |grep -Fm 1 'logical name:' |awk -- '{print $NF}')
	sed -i "s/masterIf:.*/masterIf: $phy_if/" $(HELM_VALUES_DIR)/global.yaml
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

upf:
	phy_if=$(lshw -C network |grep -Fm 1 'logical name:' |awk -- '{print $NF}')
	sed -i "s/\"master\":.*/\"master\": \"$phy_if\",/" $(HELM_VALUES_DIR)/eupf.yaml
	helm $(HELM_ACTION) \
		eupf oci://registry-1.docker.io/edgecom/eupf \
		--version 0.4.0 \
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
