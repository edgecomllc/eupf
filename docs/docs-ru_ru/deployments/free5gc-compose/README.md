# Free5GC + eUPF with docker-compose

Данный пример конфигурации показывает возможность развертывания 5G сети на основе eUPF и проекта Free5GC с помощью следующих компонентов:
- Docker-compose
- UERANSIM в качестве радиочасти
- Free5GS ядро сети
- eUPF в качестве модуля UPF

Исходный пример развертывания проекта free5gc(https://github.com/free5gc/free5gc-compose) был модифицирован таким образом, чтобы заменить штатный модуль UPF на модуль eUPF. 

Также в конфигурацию добавлены:
- модуль NAT (поскольку eUPF не поддерживает функции NAT)
- модули prometheus и grafana для отображения статистики работы eUPF

Основные внесенные изменения находятся в файле `docker-compose.override.yml`, который является стандартным механзмом внесения изменений в конфигурацию docker-compose.

## Предварительные требования

- установите модуль ядра gtp5g по инструкции https://github.com/free5gc/gtp5g

- установите утилиты docker, docker-compose

См. инструкцию по ссылке https://docs.docker.com/engine/install/

## Шаги развертывания

1. Скопируйте репозиторий `edgecomllc/free5gc-compose`

`git clone https://github.com/edgecomllc/free5gc-compose`

1. перейдите в папку free5gc-compose

`cd free5gc-compose`

1. выполните запуск модулей ядра free5gc и эмуляторов gNB

    `docker-compose up -d`

1. добавьте абонента в сеть free5gc через WebUI

    откройте в браузере http://127.0.0.1:5000 (используйте "admin"/"free5gc" для авторизации), перейдите в меню "subscribers", нажмите "new subscriber" и кнопку "submit"

    остановите проброс портов `Ctrl + C`

## Проверка

1. запустите оболочку shell в контейнере UERANSIM

    `docker-compose exec ueransim bash`

1. запустите эмулятор UE в оболочке shell контейнера UERANSIM
    `./nr-ue -c config /uecfg.yaml`

1. запустите ещё одну оболочку shell в контейнере UERANSIM

    `docker-compose exec ueransim bash`

1. проверьте доступность сети с помошью команды ping в новой оболочке

    `ping -I uesimtun0 1.1.1.1`

## Удаление конфигурации

1. выполните команду

    `make clean`
