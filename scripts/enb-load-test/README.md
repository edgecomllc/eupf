# GTP-U Echo Server and Client

## Deployment

1. Initialize Docker Swarm: `docker swarm init`
2. Build image: `docker-compose build`
3. Deploy the stack: `docker stack deploy -c docker-compose.yaml test`
4. Check that services are running: `docker service ls`
5. Check service logs: `docker service logs -f test_eupf`

## Cleanup

1. Remove the stack: `docker stack rm test`
