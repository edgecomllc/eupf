# Free5GC + UCLCL + eUPF with docker-compose

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF и проекта Free5GC для демонстрации функции ULCL с помощью следующих компонентов:
- Docker-compose
- UERANSIM в качестве радиочасти
- Free5GS ядро сети
- eUPF в качестве модуля UPF

Исходный пример развертывания проекта free5gc(https://github.com/free5gc/free5gc-compose) был модифицирован таким образом, чтобы заменить штатный модуль UPF на модуль eUPF. 

Также в конфигурацию добавлены:
- модуль NAT (поскольку eUPF не поддерживает функции NAT)
- модули prometheus и grafana для отображения статистики работы eUPF

Основные внесенные изменения находятся в файле `docker-compose.ulcl.yml`, который используется в дополнение к общим настройкам `docker-compose.yml`

В данной конфигурации будут развернуты 3 eUPFs:
- I-UPF
- PSA-UPF
- PSA-UPF 2

Маршрутизация трафика выполняется следующим образом:

- UE--gNodeB--I-UPF--PSA-UPF--Internet (по-умолчанию)
- UE--gNodeB--I-UPF--PSA-UPF-2--Internet--8.8.8.8/32 (для UE c IMSI 208930000000001 и сервера 8.8.8.8)

## Предварительные требования

- установите утилиты docker, docker-compose

См. инструкцию по ссылке https://docs.docker.com/engine/install/

## Шаги развертывания

1. Скопируйте репозиторий `edgecomllc/free5gc-compose`

`git clone --branch ulcl-n9upf-experimetns https://github.com/edgecomllc/free5gc-compose`

1. перейдите в папку free5gc-compose

`cd free5gc-compose`

1. выполните запуск модулей ядра free5gc и эмуляторов gNB

    `docker-compose -f docker-compose.yaml -f docker-compose.ulcl.yml up -d`

1. добавьте абонента в сеть free5gc через WebUI

    откройте в браузере http://127.0.0.1:5000 (используйте "admin"/"free5gc" для авторизации), перейдите в меню "subscribers", нажмите "new subscriber" и кнопку "submit"

    остановите проброс портов `Ctrl + C`

## Проверка

1. запустите оболочку shell в контейнере UERANSIM

    `docker-compose -f docker-compose.yaml -f docker-compose.ulcl.yml exec ueransim bash`

1. запустите эмулятор UE в оболочке shell контейнера UERANSIM
    `./nr-ue -c config/uecfg.yaml`

1. запустите ещё одну оболочку shell в контейнере UERANSIM

    `docker-compose exec ueransim bash`

1. проверьте доступность сети с помошью команды ping в новой оболочке

    `ping -I uesimtun0 1.1.1.1`

1. проверьте доступность сети с помошью команды ping для выделенного маршрута в новой оболочке

    `ping -I uesimtun0 8.8.8.8`

## Удаление конфигурации

1. выполните команду

    `make clean`
