pull:
	@for f in $$(find ./ -type f -name "docker-compose*yaml"); do \
		docker compose -f $$f pull; \
	done

clean:
	@for f in $$(find ./ -type f -name "docker-compose*yaml"); do \
		docker compose -f $$f down; \
	done

infra:
	docker compose -f docker-compose.infra.yaml up -d

eupf:
	docker compose -f docker-compose.eupf.yaml up -d

core:
	docker compose -f docker-compose.core.yaml up -d

gnb:
	docker compose -f docker-compose.gnb.yaml up -d --scale gnb=6

ue1:
	docker compose -f docker-compose.ue1.yaml up -d

ue2:
	docker compose -f docker-compose.ue2.yaml up -d

test:
	@for c in $$(docker ps --format '{{.Names}}' | grep ue2); do \
		(docker exec -t $$c bash /opt/iperf-test.sh &); \
	done
