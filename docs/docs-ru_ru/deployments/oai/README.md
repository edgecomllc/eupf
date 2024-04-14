# OpenAir Core + OpenAir RAN

![](./schema.png)

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF и проекта OpenAirInterface с помощью следующих компонентов:
- OpenAirInterface в качестве радиочасти
- OpenAirInterface ядро сети
- eUPF в качестве модуля UPF

## Требования

- Установлена утилита[helm](https://helm.sh/docs/intro/install/)
- Сalico настроен на использование BIRD

    Для этого измените значение параметра `calico_backend` на `bird` в настройках (configmap) `calico-config` и перезапустите все поды с именем `calico-node-*`

## Шаги развертывания

1. перейдите в папку docs/deployments/oai
1. настройте параметры calico BGP. В частности, настройки Calico BGP пиринга, Calico IP Pool (для корректного NAT) и параметры модуля Felix для того, чтобы корректно сохранять маршруты в абонентскую подсеть (получаемые по BGP от eUPF)

    `make calico`

1. разверните eupf

    `make upf`
3. установите OpenAir core

`git clone -b master https://gitlab.eurecom.fr/oai/cn5g/oai-cn5g-fed`

перейдите в папку `oai-cn5g-fed/charts/oai-5g-core/oai-5g-basic`

примените чарты согласно инструкции https://gitlab.eurecom.fr/oai/cn5g/oai-cn5g-fed/-/blob/master/docs/DEPLOY_SA5G_HC.md#4-deploying-helm-charts, но при этом используйте namespace `open5gs`

4. установите gNB

    `make gnb`

5. установите UE

    `make ue`

## Проверка

1. перейдите в оболочку shell UE пода

    `kubectl -n open5gs exec -ti deployment/ue-oai-nr-ue -- /bin/bash`

2. запустите команду ping

    `ping -I oaitun_ue1 1.1.1.1`


## Удаление конфигурации

1. выполните команду

    `make clean`
