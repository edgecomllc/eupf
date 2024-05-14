# Open5GS и балансировка нагрузки между несколькими eUPF

![](./schema.png)

Данный пример конфигурации показывает возможности масштабирования плоскости обработки данных в 5G сети на основе eUPF и open-source решения Open5GS в составе:
- UERANSIM в качестве радиочасти
- Open5GS ядро сети
- eUPF в качестве модулей UPF

При наличии нескольких подключенных модулей UPF модуль SMF производит распределение новых PDU-сессий абонентов поочередно на каждый из подключенных UPF.

## Предварительные требования

- Kubernetes кластер с Calico и Multus плагинами CNI
- [Утилита helm](https://helm.sh/docs/intro/install/)
- Calico настроен на использование BIRD

    Для этого измените значение параметра `calico_backend` на `bird` в настройках (configmap) `calico-config` и перезапустите все поды с именем `calico-node-*`

## Ограничения

---

На данный момент нет возможности выполнять роутинг трафика в сторону абонентов, т.к. оба UPF работают с одинаковой абонентской подсетью и не реализуют функции NAT.

Т.о. шаги проверки не будут работать корректно.

---

# Шаги развертывания

0. перейдите в папку docs/deployments/open5gs-with-scaling-eupf

    `cd docs/deployments/open5gs-with-scaling-eupf/`

1. разверните eupf

    `make upf`

Проверьте, что созданы 2 пода eUPF и запомните их IP-вдреса:

```bash
$ kubectl get po -n open5gs -l "app.kubernetes.io/name=eupf" -o wide
NAME     READY   STATUS    RESTARTS   AGE     IP             NODE      NOMINATED NODE   READINESS GATES
eupf-0   1/1     Running   0          6m29s   10.233.64.17   edgecom   <none>           <none>
eupf-1   1/1     Running   0          6m19s   10.233.64.44   edgecom   <none>           <none>
```

2. разверните open5gs

    `make open5gs`

3. разверните SMF

    `make smf`

4. разверните gNB

    `make gnb`

5. разверните 2 эмулятора UE на основе UERANSim

    `make ue1`

    `make ue2`


6. проверьте подключение SMF к обоим UPF по логам SMF:

ue1 должен установить PDU-сессию через eUPF 0 (10.233.64.17)

```
10/16 19:05:30.573: [smf] INFO: [Added] Number of SMF-UEs is now 1 (../src/smf/context.c:898)
10/16 19:05:30.573: [smf] INFO: [Added] Number of SMF-Sessions is now 1 (../src/smf/context.c:2975)
10/16 19:05:30.596: [smf] INFO: UE SUPI[imsi-999700000000001] DNN[internet] IPv4[10.11.0.2] IPv6[] (../src/smf/npcf-handler.c:495)
10/16 19:05:30.599: [gtp] INFO: gtp_connect() [10.233.64.17]:2152 (../lib/gtp/path.c:60)
```

ue2 должен установить PDU-сессию через eUPF 1 (10.233.64.44)

```
10/16 19:05:42.749: [smf] INFO: [Added] Number of SMF-UEs is now 2 (../src/smf/context.c:898)
10/16 19:05:42.749: [smf] INFO: [Added] Number of SMF-Sessions is now 2 (../src/smf/context.c:2975)
10/16 19:05:42.767: [smf] INFO: UE SUPI[imsi-999700000000002] DNN[internet] IPv4[10.11.0.3] IPv6[] (../src/smf/npcf-handler.c:495)
10/16 19:05:42.774: [gtp] INFO: gtp_connect() [10.233.64.44]:2152 (../lib/gtp/path.c:60)
```

## Проверка

1. запустите оболочку shell в поде UE1

    `kubectl -n open5gs exec -ti deployment/ueransim1-ueransim-ues -- /bin/bash`

2. проверьте доступность сети с помошью команды ping

    `ping -I uesimtun0 1.1.1.1`

## Удаление конфигурации

1. выполните команду

    `make clean`