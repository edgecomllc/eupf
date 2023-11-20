NAMESPACE?=open5gs
HELM_VALUES_DIR?=./values

HELM_ARGS?=--timeout 30s --create-namespace
HELM_ACTION?=upgrade --install
# HELM_ACTION?=template --debug

open5gs:
	helm $(HELM_ACTION) \
		open5gs openverso/open5gs \
		--version 2.0.12 \
		--values $(HELM_VALUES_DIR)/main.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

upf:
	helm $(HELM_ACTION) \
		open5gs-upf openverso/open5gs-upf \
		--version 2.0.8 \
		--values $(HELM_VALUES_DIR)/upf.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

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

# gnb:
# 	helm $(HELM_ACTION) \
# 		gnb openverso/ueransim-gnb \
# 		--version 0.2.5 \
# 		--values $(HELM_VALUES_DIR)/old/gnb.yaml \
# 		-n $(NAMESPACE) \
# 		$(HELM_ARGS)

# ue:
# 	helm $(HELM_ACTION) \
# 		ueransim openverso/ueransim-ues \
# 		--version 0.1.1 \
# 		--values $(HELM_VALUES_DIR)/old/ue.yaml \
# 		-n $(NAMESPACE) \
# 		$(HELM_ARGS)

clean:
	helm delete open5gs -n $(NAMESPACE) || true
	helm delete open5gs-upf -n $(NAMESPACE) || true
	helm delete gnb -n $(NAMESPACE) || true
	helm delete ue -n $(NAMESPACE) || true
