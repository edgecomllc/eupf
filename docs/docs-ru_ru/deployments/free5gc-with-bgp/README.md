# Free5GC + eUPF with Calico BGP


![](/docs/deployments/free5gc-with-bgp/schema.png)

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF и проекта Free5GC с помощью следующих компонентов:
- UERANSIM в качестве радиочасти
- Free5GS ядро сети
- eUPF в качестве модуля UPF

## Предварительные требования

- Kubernetes кластер с Calico и Multus плагинами CNI
- [Утилита helm](https://helm.sh/docs/intro/install/)
- calico настроено на использование BIRD

    Для этого измените значение параметра `calico_backend` на `bird` в настройках (configmap) `calico-config` и перезапустите все поды с именем `calico-node-*`


- настроенные helm репозитории

    ```
    helm repo add towards5gs https://raw.githubusercontent.com/Orange-OpenSource/towards5gs-helm/main/repo/
    helm repo update
    ```


## Шаги развертывания
1. перейдите в папку docs/deployments/free5gc-with-bgp
1. обновите файлы values, задав корректное название сетевого интерфейса вашего нода:
    - file `values/global.yaml`: parameter `masterIf` в 5 строках
    - file `values/eupf.yaml`:  `"master": ` в 1 строке
    - file `kustomize/patch_rm_default_route_from_nad.yaml`: `"master": ` в 1 строке 

1. установитe free5gc

    `make free5gc`

1. добавьте абонента в сеть free5gc через WebUI

    Для этого пробросьте порт веб-интерфейса на локальную машину

    ```powershell
    kubectl port-forward service/webui-service 5000:5000 -n free5gc
    ```

    откройте в браузере http://127.0.0.1:5000 (используйте "admin"/"free5gc" для авторизации), перейдите в меню "subscribers", нажмите "new subscriber" и кнопку "submit"

    остановите проброс портов `Ctrl + C`

1. настройте параметры calico BGP. В частномти, настройки Calico BGP пиринга, Calico IP Pool (для корректного NAT) и параметры модуля Felix для того, чтобы корректно сохранять маршруты в абонентскую подсеть (получаемые по BGP от eUPF)

    `make calico`

1. разверните eupf

    `make upf`

1. разверните UERANSim

    `make ue1`

## Проверка

1. запустите оболочку shell в поде UE

    `kubectl -n free5gc exec -ti deployment/ueransim1-ue -- /bin/bash`

1. проверьте доступность сети с помошью команды ping

    `ping -I uesimtun0 1.1.1.1`

## Удаление конфигурации

1. выполните команду

    `make clean`
