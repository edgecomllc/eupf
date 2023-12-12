# OpenAir Core + OpenAir RAN as a docker-compose
We will use 5G L2 nFAPI simulator to test L2 and above Layers. Let's pull Eurecom's deployment for 5G SA mode with 1 User: [OAI Full Stack 5G-NR L2 simulation with containers and a proxy](https://gitlab.eurecom.fr/oai/openairinterface5g/-/tree/develop/ci-scripts/yaml_files/5g_l2sim_tdd) and replace the UPF with our eUPF.

📝This deploy uses `network_mode: "host"` for communications over `lo` interface of the host between containers oai-gnb, proxy, oai-nr-ue0.

We will add two services `edgecom-upf` and `edgecom-nat` with a dedicated network between. We have configured routing between the both, so that eUPF passes packets from the UE towards the NAT container, and the replies come through NAT to eUPF and down to the UE.

## How to deploy:
1. Deploy  the whole project "OAI Full Stack 5G-NR L2 simulation with containers and a proxy" 
    following instructions https://gitlab.eurecom.fr/oai/openairinterface5g/-/blob/develop/ci-scripts/yaml_files/5g_l2sim_tdd/README.md

    <details><summary>TLDR: Look at our example where host interface name is `ens3` with IP addr `188.120.232.247`</summary>
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

2. Copy content of folder `edgecomllc/eupf/docs/deployments/oai-nfapi-sim-compose` from https://github.com/edgecomllc/eupf repository to your `openairinterface5g/ci-scripts/yaml_files/5g_l2sim_tdd`

    📝 There is `docker-compose.override.yml` file which should happen to be in the same folder with `docker-compose.yml`. Check it before continue.
3. Start containers in the following order:
    ```ruby
    sudo docker-compose up -d mysql oai-nrf oai-amf oai-smf edgecom-nat edgecom-upf
    sudo docker-compose up -d  oai-gnb
    sudo docker-compose up -d proxy oai-nr-ue0
    ```
4. Check the interface `oaitun_ue1` is created and ip address is set: `ip a |grep oaitun`
    ```ruby
    ip a |grep oaitun
    8808: oaitun_ue1: <POINTOPOINT,MULTICAST,NOARP,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UNKNOWN group default qlen 500
        inet 12.1.1.5/24 brd 12.1.1.255 scope global oaitun_ue1
    ```
5. Check the Internet is accessible over created GTP tunnel
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
6. `sudo docker-compose down oai-nr-ue0 proxy` to free up resources before the system sluggish.

### Undeploy all
`sudo docker-compose down`

