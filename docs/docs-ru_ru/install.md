# Как установить и запустить eUPF
Самый простой способ установить eUPF — использовать helm charts для одного из поддерживаемых основных проектов 5G с открытым исходным кодом в вашем собственном кластере Kubernetes.
В качестве альтернативы eUPF можно развернуть с помощью docker-compose (конфигурации free5gc или OpenAirInterface).

Мы подготовили шаблоны для развертывания в двух средах с открытым исходным кодом: **open5gs** и **free5gc** на ваш выбор.

Проект [UERANSIM](https://github.com/aligungr/UERANSIM) используется для эмуляции конечной точки радиосвязи, поэтому вы сможете проверить сквозное соединение с виртуального пользовательского устройства до сети Интернет.

Варианты развертывания:
- baremetall/ВМ
- [Среда создания Docker](install.md/#deploy-with-docker-compose)
- [Среда Kubernetes](install.md/#deploy-with-kubernetes)

## Общие требования к узлу

**eUPF необходимо ядро Linux версии > 5.14 (мы использовали Ubuntu 22.04 LTS)**

### Поддержка драйверов

#### Драйверы, поддерживающие универсальный XDP

Начиная с ядра 4.12, вы можете запускать XDP (и eUPF) в универсальном режиме где угодно. Но его следует использовать только для целей тестирования или отладки (без производительности).

#### Драйверы, поддерживающие собственный XDP

Чтобы запустить eUPF в native режиме, вам понадобится совместимый драйвер. Список поддерживаемых драйверов можно найти в документации Cilium или IOVisor:

- См. главу «Драйверы, поддерживающие встроенный XDP» в [cilium](https://docs.cilium.io/en/latest/bpf/progtypes/#xdp).
- См. [документацию проекта bcc](https://github.com/iovisor/bcc/blob/master/docs/kernel-versions.md#xdp).

Собственный режим поддерживается в большинстве современных облаков и сетевых адаптеров виртуальных машин:
- Amazon `ena`
- Microsoft `hv_netvsc`
- VirtIO `virtio_net`
- VMWare `vmxnet3`
- SR-IOV `ixgbevf`

#### Драйверы, поддерживающие offloaded XDP

На данный момент только поддерживаются сетевые карты Netronome

# Развертывание с помощью docker-compose

## Предварительные требования

- [Docker Engine](https://docs.docker.com/engine/install): необходимо для запуска the eUPF container
- [Docker Compose v2](https://docs.docker.com/compose/install): необходимо для загрузки the eUPF container

## Конфигурация и запуск

### Созадйте docker-compose.yaml
<details><summary>docker-compose.yaml template</summary>

```yaml
version: '2.4'

services:
  eupf:
    image: ghcr.io/edgecomllc/eupf:main
    privileged: true
    volumes:
      - /sys/fs/bpf:/sys/fs/bpf
    environment:
      - GIN_MODE=release
      - UPF_INTERFACE_NAME=eth0
      - UPF_XDP_ATTACH_MODE=generic
      - UPF_API_ADDRESS=:8081
      - UPF_PFCP_ADDRESS=:8805
      - UPF_METRICS_ADDRESS=:9091
      - UPF_PFCP_NODE_ID=172.21.0.100
      - UPF_N3_ADDRESS=172.21.0.100
    ulimits:
      memlock: -1
    cap_add:
      - NET_ADMIN
      - SYS_ADMIN
    ports:
      - 2152:2152/udp
      - 8805:8805/udp
      - 8080:8080
      - 9090:9090
    restart: unless-stopped
    networks:
      local-dc:
        ipv4_address: 172.21.0.100
    sysctls:
      - net.ipv4.conf.all.forwarding=1

  net-tools:
    image: praqma/network-multitool:alpine-extra@sha256:47b259d4463950f5c10d9c0bf63d9e71ec456618f5549a414afa0c04392e0ac1
    network_mode: host
    privileged: true
    restart: unless-stopped
    command:
      - /bin/sh
      - -c
      - |
        ip ro add 10.33.0.0/16 via 172.21.0.100
        echo "done"
        tail -f /dev/null

networks:
  local-dc:
    external: true
```
</details>

### Укажите конфигурационные параметры eUPF 

Смотрите [configuration guide](Configuration.md)

### Запустите eUPF

```
docker-compose up -d
```

## Примеры запуска в docker-compose

Смотрите примеры развертывания docker-compose с помощью **open5gs**, **free5gc** и **OpenAirInterface** в [the Deployment examples table](../../docs/deployments/README.md#docker-compose-deployments).

# Развертывание в среде Kubernetes

## Предварительные требования
- Kubernetes кластер с поддержкой Calico и Multus CNI
  - с [Enabled Unsafe Sysctls](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/#enabling-unsafe-sysctls) net.ipv4.ip_forward:
    `kubelet --allowed-unsafe-sysctls 'net.ipv4.ip_forward,net.ipv6.conf.all.forwarding'`
- установлен [helm](https://helm.sh/docs/intro/install/)
- установлен [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack), или <br>
  создан CustomResource ServiceMonitor: <br>
  ```kubectl apply -f https://github.com/prometheus-community/helm-charts/raw/main/charts/kube-prometheus-stack/crds/crd-servicemonitors.yaml```
<!-- - развернуто 5g core (open5gs or free5gc) -->

В наших K8s средах мы используем кластер K8s с одним узлом, развернутый с помощью [kubespray](https://github.com/kubernetes-sigs/kubespray). <!-- Вы можете увидеть примеры конфигурации в этом [репозитории](https://github.com/edgecomllc/ansible) -private -->
<details><summary>С дополнительным файлом Inventory/mycluster/group_vars/kube_node.yaml</summary>
<p>

```yaml
---

kubelet_node_config_extra_args:
  allowedUnsafeSysctls:
    - "net.ipv4.ip_forward"
``` 
</p>
</details> 

## Маршрутизация подсетей UE

Для маршрутизации трафика нисходящей линии связи обратно в UE предлагаются следующие варианты:

1. Использовать BGP
2. Для каждого UE использовать статические маршруты

### Использование BGP

Демон BGP (BIRD) работает как дополнительный модуль в модуле eUPF. Подсеть UE объявляется узлам K8s. K8s CNI должен быть настроен на использование BGP.

Это решение подходит для развертывания eUPF в одном экземпляре.

### Использование статических маршрутов

Чтобы использовать масштабируемое развертывание eUPF (более одной реплики eUPF), маршрут нисходящей линии связи к конкретному UE должен передать UPF с соответствующим сеансом PDU.

В качестве доказательства концепции предоставляется простая утилита маршрутизации. Утилита считывает активные PDU-сессии для каждого UPF через API и обновляет таблицу маршрутизации узла. Для каждого PDU-сеанса UE существует статический маршрут через соответствующий UPF.

## Примеры

См. примеры развертывания Kubernetes с помощью **open5gs** и **free5gc** [здесь](../../docs/deployments/README.md).

# Тестовые сценарии

## Сценарий 0

<b>Описание:</b>

UE может отправить пакет в Интернет и получить ответ

<b>Действия:</b>

1. Запустите shell в поде

   для open5gs:
   ```powershell
   export NS_NAME=open5gs
   export UE_POD_NAME=$(kubectl get pods -l "app.kubernetes.io/name=ueransim-gnb,app.kubernetes.io/component=ues" --output=jsonpath="{.items..metadata.name}" -n ${NS_NAME})
   kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash
   ```

   для free5gc:

   ```powershell
   export NS_NAME=free5gc
   export UE_POD_NAME=$(kubectl get pods -l "app=ueransim,component=ue" --output=jsonpath="{.items..metadata.name}" -n ${NS_NAME})
   kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash
   ```

1. запустите команду из UE pod's shell.

   `$ ping -I uesimtun0 google.com`


   <b>Ожидаемый результат:</b>

   успешное выполнение команды ping 

# Поддержка

<details><summary>Для сборок с включенной трассировкой, а не продакшн: Подробности смотрите под спойлером.</summary>
<p>

Чтобы просмотреть журнал отладки программ eBPF, введите команду запуска консоли **node**:
`sudo cat /sys/kernel/debug/tracing/trace_pipe`

Далее переключитесь на UE pod's shell. Отправка одного пакета `ping -I use tun0 -c1 1.1.1.1` с успешным ответом, обычно вы увидите такой отладочный вывод:
```ruby
sergo@edgecom:~$ sudo cat /sys/kernel/debug/tracing/trace_pipe

          nr-gnb-4117277 [003] d.s11 266111.395788: bpf_trace_printk: upf: gtp-u received
          nr-gnb-4117277 [003] d.s11 266111.395819: bpf_trace_printk: upf: gtp pdu [ 10.100.50.236 -> 10.100.50.233 ]
          nr-gnb-4117277 [003] d.s11 266111.395825: bpf_trace_printk: upf: uplink session for teid:1 far:1 headrm:0
          nr-gnb-4117277 [003] d.s11 266111.395828: bpf_trace_printk: upf: far:1 action:2 outer_header_creation:0
          nr-gnb-4117277 [003] d.s11 266111.395831: bpf_trace_printk: upf: qer:1 gate_status:0 mbr:200000000
          nr-gnb-4117277 [003] d.s11 266111.395857: bpf_trace_printk: upf: bpf_fib_lookup 10.1.0.1 -> 1.1.1.1: nexthop: 10.100.100.1
          nr-gnb-4117277 [003] d.s11 266111.395861: bpf_trace_printk: upf: bpf_redirect: if=6 18446669071770913972 -> 18446669071770913978
          <idle>-0       [007] d.s.1 266111.396975: bpf_trace_printk: upf: downlink session for ip:10.1.0.1  far:2 action:2
          <idle>-0       [007] dNs.1 266111.396983: bpf_trace_printk: upf: qer:0 gate_status:0 mbr:0
          <idle>-0       [007] dNs.1 266111.396985: bpf_trace_printk: upf: use mapping 10.1.0.1 -> TEID:1
          <idle>-0       [007] dNs.1 266111.396987: bpf_trace_printk: upf: send gtp pdu 10.100.50.233 -> 10.100.50.236
          <idle>-0       [007] dNs.1 266111.396996: bpf_trace_printk: upf: bpf_fib_lookup 10.100.50.233 -> 10.100.50.236: nexthop: 10.100.50.236
          <idle>-0       [007] dNs.1 266111.396998: bpf_trace_printk: upf: bpf_redirect: if=4 18446669071771765924 -> 18446669071771765930
```

</p>
</details> 

## Подключение журналы компонентов:
<details><summary>Успешные соединения eUPF в выводе журналов (stdout)</summary>
<p>

```ruby
2023/04/17 16:09:39 map[api_address::8080 interface_name:n3 metrics_address::9090 pfcp_address::8805 pfcp_node_id:10.100.50.241 xdp_attach_mode:generic]
2023/04/17 16:09:39 {n3 generic :8080 :8805 10.100.50.241 :9090}
2023/04/17 16:09:40 Attached XDP program to iface "n3" (index 4)
2023/04/17 16:09:40 Press Ctrl-C to exit and remove the program
2023/04/17 16:09:40 Start PFCP connection: :8805
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:    export GIN_MODE=release
 - using code:    gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /upf_pipeline             --> main.CreateApiServer.func1 (3 handlers)
[GIN-debug] GET    /qer_map                  --> main.CreateApiServer.func2 (3 handlers)
[GIN-debug] GET    /pfcp_associations        --> main.CreateApiServer.func3 (3 handlers)
[GIN-debug] GET    /config                   --> main.CreateApiServer.func4 (3 handlers)
[GIN-debug] GET    /xdp_stats                --> main.CreateApiServer.func5 (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Listening and serving HTTP on :8080
2023/04/17 16:11:13 Received 30 bytes from 10.100.50.244:8805
2023/04/17 16:11:13 Handling PFCP message from 10.100.50.244:8805
2023/04/17 16:11:13 Got Association Setup Request from: 10.100.50.244:8805.
2023/04/17 16:11:13
Association Setup Request:
  Node ID: 10.100.50.244
  Recovery Time: 2023-04-17 16:11:13 +0000 UTC
2023/04/17 16:11:13 Saving new association: {ID:10.100.50.244 Addr:10.100.50.244:8805 NextSessionID:1 Sessions:map[]}
2023/04/17 16:11:50 Received 287 bytes from 10.100.50.244:8805
2023/04/17 16:11:50 Handling PFCP message from 10.100.50.244:8805
2023/04/17 16:11:50 Got Session Establishment Request from: 10.100.50.244:8805.
2023/04/17 16:11:50
Session Establishment Request:
  CreatePDR ID: 1
    Outer Header Removal: 0
    FAR ID: 1
    Source Interface: 0
    TEID: 1
    Ipv4: 10.100.50.233
    Ipv6: <nil>
  CreatePDR ID: 2
    FAR ID: 2
    Source Interface: 2
    UE IPv4 Address: 10.1.0.1
  CreateFAR ID: 1
    Apply Action: [2]
    Forwarding Parameters:
      Network Instance: internet
  CreateFAR ID: 2
    Apply Action: [2]
    Forwarding Parameters:
  CreateQER ID: 1
    Gate Status DL: 0
    Gate Status UL: 0
    Max Bitrate DL: 100000
    Max Bitrate UL: 200000
    QFI: 9

2023/04/17 16:11:50
Session Establishment Request:
  CreatePDR ID: 1
    Outer Header Removal: 0
    FAR ID: 1
    Source Interface: 0
    TEID: 1
    Ipv4: 10.100.50.233
    Ipv6: <nil>
  CreatePDR ID: 2
    FAR ID: 2
    Source Interface: 2
    UE IPv4 Address: 10.1.0.1
  CreateFAR ID: 1
    Apply Action: [2]
    Forwarding Parameters:
      Network Instance: internet
  CreateFAR ID: 2
    Apply Action: [2]
    Forwarding Parameters:
  CreateQER ID: 1
    Gate Status DL: 0
    Gate Status UL: 0
    Max Bitrate DL: 100000
    Max Bitrate UL: 200000
    QFI: 9
2023/04/17 16:11:50 WARN: No OuterHeaderCreation
2023/04/17 16:11:50 Saving FAR info to session: 1, {Action:2 OuterHeaderCreation:0 Teid:0 RemoteIP:0 LocalIP:0}
2023/04/17 16:11:50 EBPF: Put FAR: i=1, farInfo={Action:2 OuterHeaderCreation:0 Teid:0 RemoteIP:0 LocalIP:0}
2023/04/17 16:11:50 Saving FAR info to session: 2, {Action:2 OuterHeaderCreation:0 Teid:0 RemoteIP:0 LocalIP:0}
2023/04/17 16:11:50 EBPF: Put FAR: i=2, farInfo={Action:2 OuterHeaderCreation:0 Teid:0 RemoteIP:0 LocalIP:0}
2023/04/17 16:11:50 Saving uplink PDR info to session: 1, {PdrInfo:{OuterHeaderRemoval:0 FarId:1} Teid:1 Ipv4:<nil>}
2023/04/17 16:11:50 EBPF: Put PDR Uplink: teid=1, pdrInfo={OuterHeaderRemoval:0 FarId:1}
2023/04/17 16:11:50 Saving downlink PDR info to session: 2, {PdrInfo:{OuterHeaderRemoval:0 FarId:2} Teid:0 Ipv4:10.1.0.1}
2023/04/17 16:11:50 EBPF: Put PDR Downlink: ipv4=10.1.0.1, pdrInfo={OuterHeaderRemoval:0 FarId:2}
2023/04/17 16:11:50 Saving QER info to session: 1, {GateStatusUL:0 GateStatusDL:0 Qfi:9 MaxBitrateUL:200000 MaxBitrateDL:100000}
2023/04/17 16:11:50 Creating QER ID: 1, QER Info: {GateStatusUL:0 GateStatusDL:0 Qfi:9 MaxBitrateUL:200000 MaxBitrateDL:100000}
2023/04/17 16:11:50 EBPF: Put QER: i=1, qerInfo={GateStatusUL:0 GateStatusDL:0 Qfi:9 MaxBitrateUL:200000 MaxBitrateDL:100000}
2023/04/17 16:11:50 Received 148 bytes from 10.100.50.244:8805
2023/04/17 16:11:50 Handling PFCP message from 10.100.50.244:8805
2023/04/17 16:11:50 Got Session Modification Request from: 10.100.50.244:8805.
2023/04/17 16:11:50 Finding association for 10.100.50.244:8805
2023/04/17 16:11:50 Finding session 2
2023/04/17 16:11:50
Session Modification Request:
  UpdatePDR ID: 2
    FAR ID: 2
    Source Interface: 2
    UE IPv4 Address: 10.1.0.1
  UpdateFAR ID: 2
    Apply Action: [2]
    Forwarding Parameters:
2023/04/17 16:11:50 Updating FAR info: 2, {Action:2 OuterHeaderCreation:1 Teid:2 RemoteIP:3962725386 LocalIP:0}
2023/04/17 16:11:50 EBPF: Update FAR: i=2, farInfo={Action:2 OuterHeaderCreation:1 Teid:2 RemoteIP:3962725386 LocalIP:0}
2023/04/17 16:11:50 Updating downlink PDR: 2, {PdrInfo:{OuterHeaderRemoval:0 FarId:2} Teid:0 Ipv4:10.1.0.1}
2023/04/17 16:11:50 EBPF: Update PDR Downlink: ipv4=10.1.0.1, pdrInfo={OuterHeaderRemoval:0 FarId:2}
Stream closed EOF for free5gc/edgecomllc-eupf-universal-chart-d4b54d4b7-t2hr6 (app)
```

</p>
</details>

<details><summary>Журналы SMF free5gc (stdout)</summary>
<p>

```ruby
2023-04-17T16:11:13Z [INFO][SMF][CFG] SMF config version [1.0.2]
2023-04-17T16:11:13Z [INFO][SMF][CFG] UE-Routing config version [1.0.1]
2023-04-17T16:11:13Z [INFO][SMF][Init] SMF Log level is set to [info] level
2023-04-17T16:11:13Z [INFO][LIB][NAS] set log level : info
2023-04-17T16:11:13Z [INFO][LIB][NAS] set report call : false
2023-04-17T16:11:13Z [INFO][LIB][NGAP] set log level : info
2023-04-17T16:11:13Z [INFO][LIB][NGAP] set report call : false
2023-04-17T16:11:13Z [INFO][LIB][Aper] set log level : info
2023-04-17T16:11:13Z [INFO][LIB][Aper] set report call : false
2023-04-17T16:11:13Z [INFO][LIB][PFCP] set log level : info
2023-04-17T16:11:13Z [INFO][LIB][PFCP] set report call : false
2023-04-17T16:11:13Z [INFO][SMF][App] smf
2023-04-17T16:11:13Z [INFO][SMF][App] SMF version:
    free5GC version: v3.2.1
    build time:      2023-03-13T18:13:22Z
    commit hash:     de70bf6c
    commit time:     2022-06-28T04:52:40Z
    go version:      go1.14.4 linux/amd64
2023-04-17T16:11:13Z [INFO][SMF][CTX] smfconfig Info: Version[1.0.2] Description[SMF initial local configuration]
2023-04-17T16:11:13Z [INFO][SMF][CTX] Endpoints: [10.100.50.233]
2023-04-17T16:11:13Z [INFO][SMF][Init] Server started
2023-04-17T16:11:13Z [INFO][SMF][Init] SMF Registration to NRF {4892acc3-b6b3-418f-b791-f2b300277fe9 SMF REGISTERED 0 0xc00024f480 0xc00024f4c0 [] []   [free5gc-free5gc-smf-service] [] <nil> [] [] <nil> 0 0 0 area1 <nil> <nil> <nil> <nil> 0xc00002ee40 <nil> <nil> <nil> <nil> <nil> map[] <nil> false 0xc00024f300 false false []}
2023-04-17T16:11:13Z [INFO][SMF][PFCP] Listen on 10.100.50.244:8805
2023-04-17T16:11:13Z [INFO][SMF][App] Sending PFCP Association Request to UPF[10.100.50.241]
2023-04-17T16:11:13Z [INFO][LIB][PFCP] Remove Request Transaction [1]
2023-04-17T16:11:13Z [INFO][SMF][App] Received PFCP Association Setup Accepted Response from UPF[10.100.50.241]
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Receive Create SM Context Request
2023-04-17T16:11:50Z [INFO][SMF][PduSess] In HandlePDUSessionSMContextCreate
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Send NF Discovery Serving UDM Successfully
2023-04-17T16:11:50Z [INFO][SMF][CTX] Allocated UE IP address: 10.1.0.1
2023-04-17T16:11:50Z [INFO][SMF][CTX] Selected UPF: UPF
2023-04-17T16:11:50Z [INFO][SMF][PduSess] UE[imsi-208930000000003] PDUSessionID[1] IP[10.1.0.1]
2023-04-17T16:11:50Z [INFO][SMF][GSM] In HandlePDUSessionEstablishmentRequest
2023-04-17T16:11:50Z [INFO][NAS][Convert] ProtocolOrContainerList:  [0xc0004aaa80 0xc0004aaac0]
2023-04-17T16:11:50Z [INFO][SMF][GSM] Protocol Configuration Options
2023-04-17T16:11:50Z [INFO][SMF][GSM] &{[0xc0004aaa80 0xc0004aaac0]}
2023-04-17T16:11:50Z [INFO][SMF][GSM] Didn't Implement container type IPAddressAllocationViaNASSignallingUL
2023-04-17T16:11:50Z [INFO][SMF][PduSess] PCF Selection for SMContext SUPI[imsi-208930000000003] PDUSessionID[1]
2023-04-17T16:11:50Z [INFO][SMF][PduSess] SUPI[imsi-208930000000003] has no pre-config route
2023-04-17T16:11:50Z [INFO][SMF][Consumer] SendNFDiscoveryServingAMF ok
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Sending PFCP Session Establishment Request
2023-04-17T16:11:50Z [INFO][SMF][GIN] | 201 |   10.233.78.130 | POST    | /nsmf-pdusession/v1/sm-contexts |
2023-04-17T16:11:50Z [INFO][LIB][PFCP] Remove Request Transaction [2]
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Received PFCP Session Establishment Accepted Response
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Receive Update SM Context Request
2023-04-17T16:11:50Z [INFO][SMF][PduSess] In HandlePDUSessionSMContextUpdate
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Sending PFCP Session Modification Request to AN UPF
2023-04-17T16:11:50Z [INFO][LIB][PFCP] Remove Request Transaction [3]
2023-04-17T16:11:50Z [INFO][SMF][PduSess] Received PFCP Session Modification Accepted Response from AN UPF
2023-04-17T16:11:50Z [INFO][SMF][GIN] | 200 |   10.233.78.130 | POST    | /nsmf-pdusession/v1/sm-contexts/urn:uuid:6dffeab5-0861-490d-8cb0-f5528e8e21a9/modify |
```

</p>
</details>

<details><summary>Журналы UERANSIM UE:</summary>
<p>

```ruby
UERANSIM v3.2.6
[2023-04-25 14:36:41.461] [nas] [info] UE switches to state [MM-DEREGISTERED/PLMN-SEARCH]
[2023-04-25 14:36:41.462] [rrc] [warning] Acceptable cell selection failed, no cell is in coverage
[2023-04-25 14:36:41.462] [rrc] [error] Cell selection failure, no suitable or acceptable cell found
[2023-04-25 14:36:42.464] [rrc] [debug] New signal detected for cell[1], total [1] cells in coverage
[2023-04-25 14:36:43.663] [nas] [error] PLMN selection failure, no cells in coverage
[2023-04-25 14:36:45.865] [nas] [error] PLMN selection failure, no cells in coverage
[2023-04-25 14:36:46.966] [nas] [info] UE switches to state [MM-DEREGISTERED/NO-CELL-AVAILABLE]
[2023-04-25 14:36:47.939] [nas] [info] Selected plmn[208/93]
[2023-04-25 14:36:47.939] [rrc] [info] Selected cell plmn[208/93] tac[1] category[SUITABLE]
[2023-04-25 14:36:47.940] [nas] [info] UE switches to state [MM-DEREGISTERED/PS]
[2023-04-25 14:36:47.940] [nas] [info] UE switches to state [MM-DEREGISTERED/NORMAL-SERVICE]
[2023-04-25 14:36:47.940] [nas] [debug] Initial registration required due to [MM-DEREG-NORMAL-SERVICE]
[2023-04-25 14:36:47.940] [nas] [debug] UAC access attempt is allowed for identity[0], category[MO_sig]
[2023-04-25 14:36:47.940] [nas] [debug] Sending Initial Registration
[2023-04-25 14:36:47.940] [nas] [info] UE switches to state [MM-REGISTER-INITIATED]
[2023-04-25 14:36:47.940] [rrc] [debug] Sending RRC Setup Request
[2023-04-25 14:36:47.941] [rrc] [info] RRC connection established
[2023-04-25 14:36:47.941] [rrc] [info] UE switches to state [RRC-CONNECTED]
[2023-04-25 14:36:47.941] [nas] [info] UE switches to state [CM-CONNECTED]
[2023-04-25 14:36:47.993] [nas] [debug] Authentication Request received
[2023-04-25 14:36:47.994] [nas] [debug] Sending Authentication Failure due to SQN out of range
[2023-04-25 14:36:48.020] [nas] [debug] Authentication Request received
[2023-04-25 14:36:48.048] [nas] [debug] Security Mode Command received
[2023-04-25 14:36:48.048] [nas] [debug] Selected integrity[2] ciphering[0]
[2023-04-25 14:36:48.137] [nas] [debug] Registration accept received
[2023-04-25 14:36:48.137] [nas] [info] UE switches to state [MM-REGISTERED/NORMAL-SERVICE]
[2023-04-25 14:36:48.137] [nas] [debug] Sending Registration Complete
[2023-04-25 14:36:48.137] [nas] [info] Initial Registration is successful
[2023-04-25 14:36:48.137] [nas] [debug] Sending PDU Session Establishment Request
[2023-04-25 14:36:48.137] [nas] [debug] UAC access attempt is allowed for identity[0], category[MO_sig]
[2023-04-25 14:36:48.447] [nas] [debug] PDU Session Establishment Accept received
[2023-04-25 14:36:48.447] [nas] [info] PDU Session establishment is successful PSI[1]
[2023-04-25 14:36:48.478] [app] [info] Connection setup for PDU session[1] is successful, TUN interface[uesimtun0, 10.1.0.1] is up.
Stream closed EOF for free5gc/ueransim-ue-7f76db59c9-c4ltw (ue)
```

</p>
</details>

<details><summary>Статусы UERANSIM UE "<strong>cm-state: CM-CONNECTED</strong>"</summary>
<p>

Откройте UE pod's shell. `kubectl exec -n ${NS_NAME} --stdin --tty ${UE_POD_NAME} -- /bin/bash`

- Команда для open5gs openverso: `nr-cli imsi-999700000000001 -e status`

- Команда для free5gc towards5gs: `./nr-cli imsi-208930000000003 -e status`

```ruby
<<K9s-Shell>> Pod: open5gs/ueransim-ueransim-gnb-ues-5b9d9c577b-zwb6d | Container: ues
bash-5.1# nr-cli imsi-999700000000001 -e status
cm-state: CM-CONNECTED
rm-state: RM-REGISTERED
mm-state: MM-REGISTERED/NORMAL-SERVICE
5u-state: 5U1-UPDATED
sim-inserted: true
selected-plmn: 999/70
current-cell: 1
current-plmn: 999/70
current-tac: 1
last-tai: PLMN[999/70] TAC[1]
stored-suci: no-identity
stored-guti:
 plmn: 999/70
 amf-region-id: 0x02
 amf-set-id: 1
 amf-pointer: 0
 tmsi: 0xf9007746
has-emergency: false
bash-5.1#
bash-5.1# ping -I uesimtun0 -c1 1.1.1.1
PING 1.1.1.1 (1.1.1.1): 56 data bytes
64 bytes from 1.1.1.1: seq=0 ttl=57 time=2.360 ms

--- 1.1.1.1 ping statistics ---
1 packets transmitted, 1 packets received, 0% packet loss
round-trip min/avg/max = 2.360/2.360/2.360 ms
bash-5.1#
bash-5.1# traceroute -i uesimtun0 www.google.com
traceroute to www.google.com (74.125.205.99), 30 hops max, 46 byte packets
 1  10.100.111.1 (10.100.111.1)  1.524 ms  1.246 ms  0.928 ms
 2  10.0.0.1 (10.0.0.1)  0.946 ms  1.722 ms  1.116 ms
 3  172.31.141.1 (172.31.141.1)  1.778 ms  1.990 ms  1.691 ms
 4  172.17.23.111 (172.17.23.111)  1.268 ms  1.822 ms  1.535 ms
 ......
```

</p>
</details>

## Отключите UE 

<details><summary>Отключение UERANSIM UE:</summary>
<p>

**cm-state: CM-IDLE**
```ruby
root@ueransim-ue-7f76db59c9-c4ltw:/ueransim/build# ./nr-cli imsi-208930000000003 -e status
cm-state: CM-IDLE
rm-state: RM-REGISTERED
mm-state: MM-REGISTERED/NORMAL-SERVICE
5u-state: 5U1-UPDATED
sim-inserted: true
selected-plmn: 208/93
current-cell: 2
current-plmn: 208/93
current-tac: 1
last-tai: PLMN[208/93] TAC[1]
stored-suci: no-identity
stored-guti:
 plmn: 208/93
 amf-region-id: 0xca
 amf-set-id: 1016
 amf-pointer: 0
 tmsi: 0x00000001
has-emergency: false
root@ueransim-ue-7f76db59c9-c4ltw:/ueransim/build#
```

Попробуйте пересоединиться:

- Команды для open5gs openverso: `nr-cli imsi-999700000000001 -e "deregister normal"`

- Команды для free5gc towards5gs: `./nr-cli imsi-208930000000003 -e "deregister normal"`

UE отправит начальную регистрацию через 10 секунд.

</p>
</details>

Если соединение не удается установить, рекомендуем перезапустить компоненты в следующей последовательности:
1. SMF
1. AMF
1. UERANSIM GnB
1. UERANSIM UE

## eUPF [API](api.md)
- Чтобы проверить текущую примененную конфигурацию, используйте GET `/api/v1/config`
- Чтобы проверить подключенные сеансы, используйте GET `/api/v1/pfcp_associations`

Вы можете перенаправить API-порт (по умолчанию 8080) из работающего контейнера eUPF на свой компьютер и использовать красивый графический интерфейс, открыв ссылку http://localhost:8080/swagger/index.html в браузере.

Или вы можете просто открыть оболочку внутри контейнера и запустить команды:
`wget  -O - http://localhost:8080/api/v1/config` and `wget  -O - http://localhost:8080/api/v1/pfcp_associations`

<details><summary> примеры вывода успешно подключенного UE в API json</summary>
<p>

```json
/ # wget  -O - http://localhost:8080/api/v1/config
Connecting to localhost:8080 ([::1]:8080)
writing to stdout
{
    "InterfaceName": [
        "n3",
        "n6"
    ],
    "XDPAttachMode": "generic",
    "ApiAddress": ":8080",
    "PfcpAddress": ":8805",
    "PfcpNodeId": "10.100.50.241",
    "MetricsAddress": ":9090",
    "N3Address": "10.100.50.233"
-                    100% |************************************************************************************************************|   246  0:00:00 ETA
written to stdout
/ #
/ # wget  -O - http://localhost:8080/api/v1/pfcp_associations
Connecting to localhost:8080 ([::1]:8080)
writing to stdout
{
    "10.100.50.244:8805": {
        "ID": "10.100.50.244",
        "Addr": "10.100.50.244:8805",
        "NextSessionID": 2,
        "Sessions": {
            "2": {
                "LocalSEID": 2,
                "RemoteSEID": 1,
                "PDRs": {
                    "1": {
                        "PdrInfo": {
                            "OuterHeaderRemoval": 0,
                            "FarId": 1,
                            "QerId": 1
                        },
                        "Teid": 1,
                        "Ipv4": ""
                    },
                    "2": {
                        "PdrInfo": {
                            "OuterHeaderRemoval": 0,
                            "FarId": 2,
                            "QerId": 0
                        },
                        "Teid": 0,
                        "Ipv4": "10.1.0.1"
                    }
                },
                "FARs": {
                    "1": {
                        "Action": 2,
                        "OuterHeaderCreation": 0,
                        "Teid": 0,
                        "RemoteIP": 0,
                        "LocalIP": 0
                    },
                    "2": {
                        "Action": 2,
                        "OuterHeaderCreation": 1,
                        "Teid": 3,
                        "RemoteIP": 3962725386,
                        "LocalIP": 3912393738
                    }
                },
                "QERs": {
                    "1": {
                        "GateStatusUL": 0,
                        "GateStatusDL": 0,
                        "Qfi": 9,
                        "MaxBitrateUL": 200000000,
                        "MaxBitrateDL": 100000000,
                        "StartUL": 0,
                        "StartDL": 0
                    }
                }
            }
        }
    }
-                    100% |************************************************************************************************************|  1954  0:00:00 ETA
written to stdout
/ #
```

</p>
</details>
