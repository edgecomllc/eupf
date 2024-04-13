# OpenAir Core + OpenAir RAN as a docker-compose

В данной инструкции используется 5G L2 nFAPI симулятор из проекта OpenAirInterface для эмуляции радиосети и eUPF модуль в качестве замены штатного модуля UPF. Развертывание основано на инструкции Eurecom для 5G SA режиме с 1 абонентом: [OAI Full Stack 5G-NR L2 simulation with containers and a proxy](https://gitlab.eurecom.fr/oai/openairinterface5g/-/tree/develop/ci-scripts/yaml_files/5g_l2sim_tdd), в которой штатный UPF заменен на eUPF.

📝В развертывании используется настройка `network_mode: "host"` для того, чтобы обеспечить взаимодействие через `lo` инерфейс хоста между контейнерами oai-gnb, proxy, oai-nr-ue0.

В развертывании используются 2 дополнительных сервиса `edgecom-upf` и `edgecom-nat` с отдельной сетью для их взаимодействия. Роутинг в рамках этой сети организован таким образом, что eUPF передает сетевые пакеты от UE через NAT-контейнер, а ответные пакета идут через NAT-контейнер в сторону eUPF и далее в сторону UE.

## Инструкция по развертыванию
1. Разверните ядро сети и эмулятор радиосети согласно инструкции "OAI Full Stack 5G-NR L2 simulation with containers and a proxy"
    по ссылке https://gitlab.eurecom.fr/oai/openairinterface5g/-/blob/develop/ci-scripts/yaml_files/5g_l2sim_tdd/README.md

    <details><summary>TLDR: Пример команд для хоста с сетевым интерфейсом `ens3` и IP-адресом `188.120.232.247`</summary>
    <p>

    ```ruby
    sergo@edgecom:~/gitlab$ git clone https://gitlab.eurecom.fr/oai/openairinterface5g.git
    sergo@edgecom:~/gitlab$ cd openairinterface5g/ci-scripts/yaml_files/5g_l2sim_tdd/

    nano docker-compose.yaml
                - DEFAULT_DNS_IPV4_ADDRESS=169.254.25.10  #172.21.3.100

    nano ../../conf_files/gnb.sa.band78.106prb.l2sim.conf
        NETWORK_INTERFACES :
        {
            GNB_INTERFACE_NAME_FOR_NG_AMF            = "ens3";
            GNB_IPV4_ADDRESS_FOR_NG_AMF              = "188.120.232.247";
            GNB_INTERFACE_NAME_FOR_NGU               = "ens3";
            GNB_IPV4_ADDRESS_FOR_NGU                 = "188.120.232.247";
            GNB_PORT_FOR_NGU                         = 2152; # Spec 2152
        };

    nano ../../conf_files/nrue.band78.106prb.l2sim.conf
    MACRLCs = (
            {
            num_cc = 1;
            tr_n_preference = "nfapi";
            local_n_if_name  = "ens3";
            remote_n_address = "127.0.0.1"; //Proxy IP
            local_n_address  = "127.0.0.1";


    sudo docker pull oaisoftwarealliance/proxy:develop
    sudo docker image tag oaisoftwarealliance/proxy:develop oai-lte-multi-ue-proxy:latest

    sudo ifconfig lo: 127.0.0.2 netmask 255.0.0.0 up
    ```

    </p>
    </details> 
    </p>

    **Don't start it.**

2. Скопируйте папку `edgecomllc/eupf/docs/deployments/oai-nfapi-sim-compose` по ссылке https://github.com/edgecomllc/eupf внутрь `openairinterface5g/ci-scripts/yaml_files/5g_l2sim_tdd`

    📝 В папук находится файл `docker-compose.override.yml`, который должен оказаться в одной папке с файлом `docker-compose.yml` из основного проекта. Проверьте это перед тем как продолжить.
3. Запустите docker-контейнеры в следующем порядке:
    ```ruby
    sudo docker-compose up -d mysql oai-nrf oai-amf oai-smf edgecom-nat edgecom-upf
    sudo docker-compose up -d  oai-gnb
    sudo docker-compose up -d proxy oai-nr-ue0
    ```
4. Проверьте, что в системе появился сетевой интерфейс `oaitun_ue1` и на нем настроен IP-адрес: `ip a |grep oaitun`
    ```ruby
    ip a |grep oaitun
    8808: oaitun_ue1: <POINTOPOINT,MULTICAST,NOARP,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UNKNOWN group default qlen 500
        inet 12.1.1.5/24 brd 12.1.1.255 scope global oaitun_ue1
    ```
5. Проверьте, что через этот интерфейс есть доступ в сеть Интернет
    ```ruby
    ping -I oaitun_ue1 -c 3 1.1.1.1
    PING 1.1.1.1 (1.1.1.1) from 12.1.1.5 oaitun_ue1: 56(84) bytes of data.
    64 bytes from 1.1.1.1: icmp_seq=1 ttl=54 time=192 ms
    64 bytes from 1.1.1.1: icmp_seq=2 ttl=54 time=68.0 ms
    64 bytes from 1.1.1.1: icmp_seq=3 ttl=54 time=150 ms
    
    --- 1.1.1.1 ping statistics ---
    3 packets transmitted, 3 received, 0% packet loss, time 2000ms
    rtt min/avg/max/mdev = 67.998/136.369/191.533/51.290 ms
    ```
6. Поскольку модули OpenAirInterface могут сильно нагружить системы, то используйти команду `sudo docker-compose down oai-nr-ue0 proxy` для того, чтобы остановить основные модули, которые могут нагружать систему.

### Удаление конфигурации
`sudo docker-compose down`

