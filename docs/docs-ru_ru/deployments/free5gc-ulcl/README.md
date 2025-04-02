Данная инструкция содержит описание развертывания реализации 5G ядра Free5GC с использованием eUPF в Kubernetes на основе конфигурации развертывания(helmcharts) из проекта [Orange-OpenSource/towards5gs-helm](https://github.com/Orange-OpenSource/towards5gs-helm) и настроек docs/deployments/free5gc-ulcl. 

## Опция UpLink CLassifier (ULCL)

Опция ULCL подразумевает развертывание дополнительного модуля UPF(I-UPF), который маршрутизирует трафик между терминальными UPF (PSA-UPF).  

В данном пример конфигурация маршрутизации следующая:
- UE--gNodeB--upfb--upf1--Internet (маршрут по-умолчанию)
- UE--gNodeB--upfb--upf2--Internet--1.1.1.1/32 (маршрут для абонента с imsi 208930000000003)

eUPF разворачивается как upfb. В качестве upf1 и upf2 используются модули из free5gc.

Разницу в маршрутизации можно проверить с использованием команды traceroute, выполняемой на эмуляторе UE:
```powershell
bash-5.1# traceroute -i uesimtun0 www.google.com -w 1
traceroute to www.google.com (173.194.222.103), 30 hops max, 46 byte packets
 1  10.233.64.41 (10.233.64.41)  1.518 ms  1.805 ms  1.459 ms
 ......
 bash-5.1# traceroute -i uesimtun0 -w1 1.1.1.1
traceroute to 1.1.1.1 (1.1.1.1), 30 hops max, 46 byte packets
 1  10.233.64.56 (10.233.64.56)  1.512 ms  1.176 ms  0.778 ms
```

## Краткая инструкция

### подготовить kubernetes хост - установить модуль ядра gtp5g

Команды для сборки и установки модуля gtp5g, который требуется для Free5gc UPFs:

```
apt-get update; apt-get install git build-essential -y; \
cd /tmp; \
git clone --depth 1 https://github.com/free5gc/gtp5g.git; \
cd gtp5g/; \
make && make install
```

После установки можно проверить, что модуль загрузился:

`lsmod | grep ^gtp5g`

* [установить утилиту helm](https://helm.sh/docs/intro/install/)

* добавить helm-репозиторий towards5gs

    ```
    helm repo add towards5gs 'https://raw.githubusercontent.com/Orange-OpenSource/towards5gs-helm/main/repo/'
    helm repo update
    ```

### Развернуть конфигурацию free5gculcl с помощью команд make
📝 Чтобы избежать конфликтов IP-адресации из-за использования сети ipvlan рекомендуется остановить любые другие поды, которые могут также использовать ipvlan настройки.
1. Перейдите в папку docs/deployments/free5gc-ulcl
1. обновите файлы values, задав корректное название сетевого интерфейса вашего нода:
    - file `global.yaml`: parameter `masterIf` в 5 строках
    - file `eupf-b.yaml`:  `"master": ` в 2 строках
    - file `kustomize/patch_rm_default_route_from_nad.yaml`: `"master": ` в 1 строке 
1. Выполните `make eupf` для развернывания eUPF в качестве upfb
1. Выполните `make upf` для развернывания Free5gc модулей UPF в качестве upf1, upf2
1. `make free5gc` to install free5gc core
1. Добавьте нового абонента в систему через web-интерфейс настройки free5gc

   Для этого пробросьте порт пода с веб-интерфейсом на localhost

   ```powershell
   kubectl port-forward service/webui-service 5000:5000 -n free5gc
   ```

   Откройте в браузере http://127.0.0.1:5000 (Пользователь "admin" с паролем "free5gc"), перейдите в меню "subscribers", нажмите "new subscriber", оставьте все значения без изменений, нажмите "submit"

   Остановите проброс порта командой `Ctrl + C`

1. Выполните `make ueransim` для развертывания эмуляторов gNodeB and UE.

После установки войдите в оболочку пода ueransim:

* `make ueransim_shell`

  Команда `make clean` удалит все установленные поды из кластера
