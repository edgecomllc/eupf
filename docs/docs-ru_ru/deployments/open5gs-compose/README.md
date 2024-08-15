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
    -o "com.docker.network.bridge.name"="br-open5gs-main" \
    open5gs-main
sudo ethtool -K br-open5gs-main tx off
```

<details><summary> <i>Здесь мы отключаем tx offload чтобы не было TCP checksum error на внутренних пакетах, для тестов iperf.</summary>
<p>

Видимо дело в том, что по-умолчанию включен offloading расчета контрольных сумм и ядро до последнего откладывает вычисление csum в расчете на то, что csum посчитается в драйвере при отправке пакета. Но у нас виртуальное окружение и пакет в итоге улетает в туннель GTP на UPF. Видимо из-за этого не происходит корректного расчета csum.

При этом если отключить offloading, то контрольная сумма считается сразу же на iperf и всё работает.

</p>
</details> 

<!---
# настройте правила firewall

`bash fw.sh`
--><br>
# запустите iperf-сервисы

`sudo apt install iperf3`

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

`docker network rm open5gs-main`