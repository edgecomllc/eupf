NAMESPACE?=open5gs
HELM_VALUES_DIR?=./values

HELM_ARGS?=--timeout 30s --create-namespace
HELM_ACTION?=upgrade --install
HELM_REGISTRY?=oci://registry-1.docker.io/gradiant

open5gs:
	helm $(HELM_ACTION) \
		open5gs $(HELM_REGISTRY)/open5gs \
		--version 2.1.0 \
		--values $(HELM_VALUES_DIR)/main.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

smf:
	helm $(HELM_ACTION) \
		open5gs-smf $(HELM_REGISTRY)/open5gs-smf \
		--version 2.1.0 \
		--values $(HELM_VALUES_DIR)/smf.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)
	# change default configmap to our
	kubectl apply -f $(HELM_VALUES_DIR)/../manifests/smf-configmap.yaml -n $(NAMESPACE)
	# scale deployment from 0 to 1
	kubectl scale --replicas=1 deployment open5gs-smf -n $(NAMESPACE)

smf2:
	helm $(HELM_ACTION) \
		open5gs-smf2 $(HELM_REGISTRY)/open5gs-smf \
		--version 2.1.0 \
		--values $(HELM_VALUES_DIR)/smf.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)
	# change default configmap to our
	kubectl apply -f $(HELM_VALUES_DIR)/../manifests/smf2-configmap.yaml -n $(NAMESPACE)
	# scale deployment from 0 to 1
	kubectl scale --replicas=1 deployment open5gs-smf2 -n $(NAMESPACE)

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

upf2:
	helm $(HELM_ACTION) \
		eupf2 oci://ghcr.io/edgecomllc/charts/eupf \
		--version 0.5.0 \
		--values $(HELM_VALUES_DIR)/eupf2.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

gnb:
	helm $(HELM_ACTION) \
		gnb $(HELM_REGISTRY)/ueransim-gnb \
		--version 0.2.5 \
		--values $(HELM_VALUES_DIR)/gnb.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ue1:
	helm $(HELM_ACTION) \
		ueransim1 $(HELM_REGISTRY)/ueransim-ues \
		--version 0.1.1 \
		--values $(HELM_VALUES_DIR)/ue.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ue2:
	helm $(HELM_ACTION) \
		ueransim2 $(HELM_REGISTRY)/ueransim-ues \
		--version 0.1.1 \
		--values $(HELM_VALUES_DIR)/ue2.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

clean:
	helm delete open5gs -n $(NAMESPACE) || true
	helm delete open5gs-smf -n $(NAMESPACE) || true
	helm delete open5gs-smf2 -n $(NAMESPACE) || true
	helm delete eupf -n $(NAMESPACE) || true
	helm delete eupf2 -n $(NAMESPACE) || true
	helm delete gnb -n $(NAMESPACE) || true
	helm delete ueransim1 -n $(NAMESPACE) || true
	helm delete ueransim2 -n $(NAMESPACE) || true
	kubectl delete -f $(HELM_VALUES_DIR)/../manifests/calico-pools.yaml || true
	kubectl delete -f $(HELM_VALUES_DIR)/../manifests/calico-bgp.yaml || true
