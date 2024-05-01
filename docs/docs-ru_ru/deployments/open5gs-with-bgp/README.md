# Open5GS + eUPF with Calico BGP

![](./schema.png)

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF с помощью следующих компонентов:
- UERANSIM в качестве радиочасти
- Open5GS ядро сети
- eUPF в качестве модуля UPF

## Предварительные требования

- Kubernetes кластер с Calico и Multus плагинами CNI
- [Утилита helm](https://helm.sh/docs/intro/install/)
- Calico настроен на использование BIRD

    Для этого измените значение параметра `calico_backend` на `bird` в настройках (configmap) `calico-config` и перезапустите все поды с именем `calico-node-*`

## Шаги развертывания

1. перейдите в папку docs/deployments/open5gs-with-bgp
1. разверните eupf

    `make upf`

1. настройте параметры calico BGP. В частности, настройки Calico BGP пиринга, Calico IP Pool (для корректного NAT) и параметры модуля Felix для того, чтобы корректно сохранять маршруты в абонентскую подсеть (получаемые по BGP от eUPF)

    `make calico`

3. разверните open5gs

    `make open5gs`

4. разверните SMF

    `make smf`

5. разверните gNB

    `make gnb`

6. разверните UERANSim

    `make ue1`

## Проверка

1. запустите оболочку shell в поде UE

    `kubectl -n open5gs exec -ti deployment/ueransim1-ueransim-ues -- /bin/bash`

1. проверьте доступность сети с помошью команды ping

    `ping -I uesimtun0 1.1.1.1`

## Удаление конфигурации

1. выполните команду

    `make clean`
