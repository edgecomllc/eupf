# open5gs-compose

Данный пример конфигурации показывает возможность развертывания тестовой 5G сети на основе eUPF с помощью следующих компонентов:
- Docker-compose
- UERANSIM в качестве радиочасти
- Open5GS ядро сети
- eUPF в качестве модуля UPF

# установите утилиты docker, docker-compose

https://docs.docker.com/engine/install/

# включите запрет для docker настройки параметров iptables

https://docs.docker.com/network/packet-filtering-firewalls/#prevent-docker-from-manipulating-iptables

# перейдите в директорию проекта

`cd docs/deployments/open5gs-compose/`

# создайте тестовую сеть

```
docker network create \
    --driver=bridge \
    --subnet=172.20.0.0/24 \
    --gateway=172.20.0.1 \
    open5gs-main
```

# настройте правила firewall

`bash fw.sh`

# запустите iperf-сервисы

`for i in $(seq 5201 5208); do (iperf3 -s -p $i &) ; done`

# запустите Open5GS

1. скачайте требуемые образы контейнеров

`make pull`

2. запустите сервисные контейнеры (mongodb)

`make infra`

3. запустите eUPF

`make eupf`

4. запустите ядро сети

`make core`

5. запустите gNB (данная команда запустит несколько экземпляров gNB, см. параметр `scale` в скрипте make)

`make gnb`

6. запустите эмулятор UE1 UERANSim, который будет взаимодействовать со штатным UPF в open5gs ядре

`make ue1`

7. запустите эмулятор UE2 UERANSim, который будет взаимодействовать с eUPF

`make ue2`

# запустите тест iperf

1. запустите iperf тест для каждого эмулятора UE

`make test`

результаты выполнения можно найти в папке `.deploy/docker/local-data/iperf` после окончания теста

# остановка и удаление всех контейнеров

`make clean`
