# Open5GS + eUPF with Calico BGP + Slices

![](./schema.png)

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF и использования функций слайсинга с помощью следующих компонентов:
- UERANSIM в качестве радиочасти
- Open5GS ядро сети
- eUPF в качестве модуля UPF

## Предварительные требования

- Kubernetes кластер с Calico и Multus плагинами CNI
- [Утилита helm](https://helm.sh/docs/intro/install/)
- Calico настроен на использование BIRD

    Для этого измените значение параметра `calico_backend` на `bird` в настройках (configmap) `calico-config` и перезапустите все поды с именем `calico-node-*`

## Шаги развертывания

1. перейдите в папку docs/deployments/open5gs-with-bgp-and-slices

1. разверните eupf для слайса 1

    `make upf`

2. разверните eupf для слайса 2

    `make upf2`

1. настройте параметры calico BGP. В частности, настройки Calico BGP пиринга, Calico IP Pool (для корректного NAT) и параметры модуля Felix для того, чтобы корректно сохранять маршруты в абонентскую подсеть (получаемые по BGP от eUPF)

    `make calico`

4. разверните open5gs

    `make open5gs`

5. настройте SMF для слайса 1

    `make smf`

6. настройте SMF для слайса 2

    `make smf2`

7. разверните gNB

    `make gnb`

8. разверните UERANSim для слайса 1

    `make ue1`

9. разверните UERANSim для слайса 2

    `make ue2`

## Проверка

1. запустите оболочку shell в поде UE1

    `kubectl -n open5gs exec -ti deployment/ueransim1-ueransim-ues -- /bin/bash`

2. проверьте доступность сети с помошью команды ping

    `ping -I uesimtun0 1.1.1.1`

3. запустите оболочку shell в поде UE2

    `kubectl -n open5gs exec -ti deployment/ueransim2-ueransim-ues -- /bin/bash`

4. проверьте доступность сети с помошью команды ping

    `ping -I uesimtun0 1.1.1.1`

## Удаление конфигурации

1. выполните команду

    `make clean`
