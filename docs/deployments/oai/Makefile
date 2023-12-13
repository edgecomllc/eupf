NAMESPACE?=open5gs
HELM_VALUES_DIR?=./values

HELM_ARGS?=--timeout 30s --create-namespace
HELM_ACTION?=upgrade --install

gnb:
	helm $(HELM_ACTION) \
		gnb charts/oai-gnb \
		--values $(HELM_VALUES_DIR)/gnb.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ue:
	helm $(HELM_ACTION) \
		ue charts/oai-nr-ue \
		--values $(HELM_VALUES_DIR)/ue.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

clean:
	helm delete mysql oai-amf oai-ausf oai-nrf oai-smf oai-spgwu-tiny oai-udm oai-udr -n $(NAMESPACE) || true
	helm delete open5gs-upf -n $(NAMESPACE) || true
	helm delete gnb -n $(NAMESPACE) || true
	helm delete ue -n $(NAMESPACE) || true
