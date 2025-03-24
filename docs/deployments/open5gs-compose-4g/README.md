# open5gs-compose

This deployment starts Open5gs 4G mobile core with UPF or eUPF.

---

## 1. Start services with Open5GS UPF

Run `docker-compose -f docker-compose.yaml up -d`

## 2. Verify environment up and running

Run `docker-compose logs srsue_zmq` and check UE logs.

If UE have successfully connected you'll see message like below:

```bash
Network attach successful. IP: 10.167.71.194
```

## 3. Test UE connectivity

Fall into UE container `docker-compose exec srsue_zmq bash`.

And use `tun_srsue` interface to send packets via it:

```bash
# ping -I tun_srsue 8.8.8.8 
PING 8.8.8.8 (8.8.8.8) from 10.46.0.2 uesimtun0: 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=58 time=15.2 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=58 time=15.2 ms
...
```

---

## 1. Start services with eUPF

Run `docker-compose up -d`

## 2. Verify environment up and running

Run `docker-compose logs srsue_zmq` and check UE logs.

If UE have successfully connected you'll see message like below:

```bash
Network attach successful. IP: 10.167.71.194
```

## 3. Test UE connectivity

Fall into UE container `docker-compose exec srsue_zmq bash`.

And use `tun_srsue` interface to send packets via it:

```bash
# ping -I tun_srsue 8.8.8.8 
PING 8.8.8.8 (8.8.8.8) from 10.46.0.2 uesimtun0: 56(84) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=1 ttl=58 time=15.2 ms
64 bytes from 8.8.8.8: icmp_seq=2 ttl=58 time=15.2 ms
...
```
