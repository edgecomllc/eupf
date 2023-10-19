NAMESPACE?=open5gs
HELM_VALUES_DIR?=./values

HELM_ARGS?=--timeout 30s --create-namespace
HELM_ACTION?=upgrade --install

upf:
	helm $(HELM_ACTION) \
		eupf oci://ghcr.io/edgecomllc/charts/eupf \
		--version 0.5.0 \
		--values $(HELM_VALUES_DIR)/eupf.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

open5gs:
	helm $(HELM_ACTION) \
		open5gs openverso/open5gs \
		--version 2.0.12 \
		--values $(HELM_VALUES_DIR)/main.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

smf:
	helm $(HELM_ACTION) \
		open5gs-smf openverso/open5gs-smf \
		--version 2.0.10 \
		--values $(HELM_VALUES_DIR)/smf.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)
	# change default configmap to our
	kubectl apply -f $(HELM_VALUES_DIR)/../manifests/smf-configmap.yaml -n $(NAMESPACE)
	# scale deployment from 0 to 1
	kubectl scale --replicas=1 deployment open5gs-smf -n $(NAMESPACE)

gnb:
	helm $(HELM_ACTION) \
		gnb openverso/ueransim-gnb \
		--version 0.2.5 \
		--values $(HELM_VALUES_DIR)/gnb.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ue1:
	helm $(HELM_ACTION) \
		ueransim1 openverso/ueransim-ues \
		--version 0.1.1 \
		--values $(HELM_VALUES_DIR)/ue1.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ue2:
	helm $(HELM_ACTION) \
		ueransim2 openverso/ueransim-ues \
		--version 0.1.1 \
		--values $(HELM_VALUES_DIR)/ue2.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

clean:
	helm delete open5gs -n $(NAMESPACE) || true
	helm delete open5gs-smf -n $(NAMESPACE) || true
	helm delete eupf -n $(NAMESPACE) || true
	helm delete gnb -n $(NAMESPACE) || true
	helm delete ueransim1 -n $(NAMESPACE) || true
	helm delete ueransim2 -n $(NAMESPACE) || true
