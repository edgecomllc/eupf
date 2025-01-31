# open5gs-compose

This deployment starts Open5gs 5G mobile core with eUPF and UERANSIM using docker-compose. 

## 1. Install docker, docker-compose

https://docs.docker.com/engine/install/

## 2. Start services

Run `docker-compose up -d`

## 3. Verify environment up and running

Run `docker-compose logs ue` and check UE logs.

If UE have successfully connected you'll see message like below:

```
[2025-01-31 15:09:30.550] [app] [info] Connection setup for PDU session[1] is successful, TUN interface[uesimtun0, 10.46.0.2] is up.
```

## 4. Test UE connectivity

Fall into UE container `docker-compose exec ue bash`.

And use `uesimtun0` interface to send packets via it:

```
# ping -I uesimtun0 8.8.8.8 
PING 8.8.8.8 (8.8.8.8) from 10.46.0.2 uesimtun0: 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=58 time=15.2 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=58 time=15.2 ms
...

curl --interface uesimtun0 -vv http://google.com
...
```