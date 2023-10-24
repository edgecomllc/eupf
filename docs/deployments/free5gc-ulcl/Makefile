NAMESPACE?=free5gculcl
HELM_ARGS?=--wait --timeout 100s --create-namespace
HELM_ACTION?=install

free5gc:
	helm $(HELM_ACTION) \
		free5gc towards5gs/free5gc \
		--values global.yaml \
		--values free5gc.yaml \
		--values upf.yaml \
		--version 1.1.7 \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

upf:
	helm $(HELM_ACTION) \
		free5gc-upf towards5gs/free5gc-upf \
		--values global.yaml \
		--values upf.yaml \
		--version 0.2.6 \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

eupf:
	helm $(HELM_ACTION) \
		edgecomllc-eupf oci://ghcr.io/edgecomllc/charts/eupf \
		--version 0.5.0 \
		--values eupf-b.yaml \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

ueransim:
	helm install \
		ueransim towards5gs/ueransim \
		--values global.yaml \
		--version 2.0.17 \
		-n $(NAMESPACE) \
		$(HELM_ARGS)

	kubectl -n $(NAMESPACE) apply -f netad-ueransim.yaml

ueransim_shell:
	$(eval UERANSIM_UE_POD=`kubectl get pods -l "app=ueransim,component=ue" --output=jsonpath="{.items..metadata.name}" -n ${NAMESPACE}`)
	kubectl exec -n $(NAMESPACE) --stdin --tty $(UERANSIM_UE_POD) -- /bin/bash

clean:
	helm delete free5gc edgecomllc-eupf ueransim -n $(NAMESPACE)
