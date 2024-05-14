# Open5GS + eUPF with Calico BGP + srsUE

![](./schema.png)

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF и open-source решения srsRAN в составе:
- srsUE и srsRAN gNodeB в качестве радиочасти
- Open5GS ядро сети
- eUPF в качестве модуля UPF

## Предварительные требования

- Kubernetes кластер с Calico и Multus плагинами CNI
- [Утилита helm](https://helm.sh/docs/intro/install/)
- Calico настроен на использование BIRD

    Для этого измените значение параметра `calico_backend` на `bird` в настройках (configmap) `calico-config` и перезапустите все поды с именем `calico-node-*`

## Шаги развертывания

0. перейдите в папку docs/deployments/srsran-gnb

    `cd docs/deployments/srsran-gnb/`

1. разверните eupf

    `make upf`

2. настройте параметры calico BGP. В частности, настройки Calico BGP пиринга, Calico IP Pool (для корректного NAT) и параметры модуля Felix для того, чтобы корректно сохранять маршруты в абонентскую подсеть (получаемые по BGP от eUPF)

    `make calico`

3. разверните open5gs

    `make open5gs`

4. разверните SMF

    `make smf`

5. разверните gNB из проекта srsUE

    `make srs`


## Проверка

1. запустите оболочку shell в поде UE1

    `kubectl -n open5gs exec -ti statefulset/srsran-srs-5g -- /bin/bash`

2. проверьте доступность сети с помошью команды ping

    `ping -I uesimtun0 1.1.1.1`

## Удаление конфигурации

1. выполните команду

    `make clean`