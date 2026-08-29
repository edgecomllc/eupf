# Конфигурация UPF

## Описание

В настоящее время UPF имеет несколько параметров конфигурации, показанных ниже.<br>Параметры можно настроить через интерфейс командной строки, файлы конфигурации (YAML, JSON) или переменные среды.

| Параметр                      | Описание                                                                                                                                                                                                                                                                                                                                     | yaml              | env                   | cli arg     | Значение по умолчанию    |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|-----------------------|-------------|-------------|
| Interface name<br>`Обязательный`  | Список сетевых интерфейсов, обрабатывающих трафик N3 (GTP) и N6 (SGi). eUPF присоединяет перехватчик XDP к каждому интерфейсу в этом списке. Формат: `[ifnameA, ifnameB, ...]`.                                                                                                                                                                                     | `interface_name`  | `UPF_INTERFACE_NAME`  | `--iface`   | `lo`        |
| N3 address <br>`Обязательный`     | IPv4 адреc для N3 интерфейса                                                                                                                                                                                                                                                                                                                   | `n3_address`      | `UPF_N3_ADDRESS`      | `--n3addr`  | `127.0.0.1` |
| N9 address <br>`Дополнительный`     | IPv4 адреc для N9 интерфейса.                                                                                                                                                                                                                                                                                                                   | `n9_address`      | `UPF_N9_ADDRESS`      | `--n9addr`  | `n3_address` |
| XDP mode <br>`Дополнительный`        | XDP attach mode: <br> ∘ **generic** – Реализация на уровне ядра. В целях оценки. <br> ∘ **native** – реализация на уровне драйвера <br> ∘ **offload** – реализация на уровне NIC. XDP можно загрузить и выполнить непосредственно на сетевой карте. <br> См. [Как работает XDP](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works) | `xdp_attach_mode` | `UPF_XDP_ATTACH_MODE` | `--attach`  | `generic`   |
| API address <br>`Дополнительный`     | Локальный адрес для обслуживания сервера [REST API](../../docs/api.md)                                                                                                                                                                                                                                                                                              | `api_address`     | `UPF_API_ADDRESS`     | `--aaddr`   | `:8080`     |
| PFCP address <br>`Дополнительный`    | Локальный адрес, по которому буедт доступен PFCP server                                                                                                                                                                                                                                                                                                    | `pfcp_address`    | `UPF_PFCP_ADDRESS`    | `--paddr`   | `:8805`     |
| PFCP NodeID <br>`Дополнительный`     | Локальный NodeID для PFCP protocol. Формет -  IPv4 address.                                                                                                                                                                                                                                                                                         | `pfcp_node_id`    | `UPF_PFCP_NODE_ID`    | `--nodeid`  | `127.0.0.1` |
| GTP peer <br>`Дополнительный`        | Список GTP-узлов, в сторону которых будут отправляться запросы GTP Echo Request. Формат: `[хостA:портA, хостB:портB, ...]`.                                                                                                                                                                                                                                                               | `gtp_peer`    | `UPF_GTP_PEER`        | `--peer`    | `-`         |
| Echo request iterval <br>`Дополнительный`        | Интервал отправки GTP Echo Request сообщений. Значение указывается в секундах.                                                                                                                                                                                                                                                                                    | `gtp_echo_interval`    | `UPF_GTP_ECHO_INTERVAL`        | `--echo`    | `10`         |
| Metrics address <br>`Дополнительный` | Локальный адрес для обслуживания метрик Prometheus.                                                                                                                                                                                                                                                                                         | `metrics_address` | `UPF_METRICS_ADDRESS` | `--maddr`   | `:9090`     |
| QER map size <br>`Дополнительный`    | Размер eBPF map для параметров QER                                                                                                                                                                                                                                                                                                          | `qer_map_size`    | `UPF_QER_MAP_SIZE`    | `--qersize` | `1024  `    |
| FAR map size <br>`Дополнительный`    | Размер eBPF map для параметров FAR                                                                                                                                                                                                                                                                                                         | `far_map_size`    | `UPF_FAR_MAP_SIZE`    | `--farsize` | `1024  `    |
| PDR map size <br>`Дополнительный`    | Размер eBPF map для параметров PDR                                                                                                                                                                                                                                                                                                         | `pdr_map_size`    | `UPF_PDR_MAP_SIZE`    | `--pdrsize` | `1024  `    |
| Logging level <br>`Дополнительный`   | Журналы уровня <= выбранного уровня будут записаны на stdout.                                                                                                                                                                                                                                                                                   | `logging_level`   | `UPF_LOGGING_LEVEL`   | `--loglvl`  | `info`         |
| FTUP Feature <br>`Дополнительный`    | Поддержка опции распределения TEID                                                                                                                                                                                                                                                                                                              | `feature_ftup`    | `UPF_FEATURE_FTUP`    | `--feature_ftup`          | `false`        |
| TEID Pool <br>`Дополнительный`       | Пул TEID, необходимый для выделения TEID, когда опция FTUP включена                                                                                                                                                                                                                                                                          | `teid_pool`       | `UPF_TEID_POOL`       | `--teid_pool`          | `65536`        |
| Пиры PFCP <br>`Дополнительный` | Список пиров PFCP (имена хостов SMF или IP-адреса), к которым UPF попытается подключиться | `pfcp_node` | `UPF_PFCP_NODE` | `--pfcprnode` | |
| Тайм-аут настройки ассоциации <br>`Дополнительный` | Тайм-аут между запросами на настройку ассоциации, инициированными UPF | `association_setup_timeout` | `UPF_ASSOCIATION_SETUP_TIMEOUT` | `--astimeout` | `5` |
|Профиль PFCP <br>`Дополнительный`|Диалект PFCP: "default" (стандарт 3GPP) или "huawei" (Huawei SPGW-C)|`pfcp_profile`|`UPF_PFCP_PROFILE`|`--pfcp_profile`|`default`|
| <br>`Дополнительный` | Адрес для связи через интерфейс S1-U | s1u_address | UPF_S1U_ADDRESS | --s1uaddr string | 127.0.0.1 |
| <br>`Дополнительный` | Адрес для связи через интерфейс S5/S8 | s5s8_address | UPF_S5S8_ADDRESS | --s5s8addr string | 127.0.0.1 |
| <br>`Дополнительный` | Адрес Sxa для привязки сервера PFCP к | sxa_address | UPF_SXA_ADDRESS | --sxaaddr string | 127.0.0.2:8805 |
| Список узлов<br>`Дополнительный` | Адрес удаленного узла Sxa | sxa_node | UPF_SXA_NODE | --sxanode stringArray | |
| <br>`Дополнительный` | Идентификатор узла сервера Sxa | sxa_node_id | UPF_SXA_NODE_ID | --sxanodeid string | 127.0.0.2 |
| <br>`Дополнительный` | Адрес Sxb для привязки сервера PFCP к | sxb_address | UPF_SXB_ADDRESS | --sxbaddr string | 127.0.0.3:8805 |
| Список узлов<br>`Дополнительный` | Адрес удаленного узла Sxb | sxb_node | UPF_SXB_NODE | --sxbnode stringArray | |
| <br>`Дополнительный` | Идентификатор узла сервера Sxb | sxb_node_id | UPF_SXB_NODE_ID | --sxbnodeid string | 127.0.0.3 |
| <br>`Дополнительный` | Адрес для связи через интерфейс PA | pa_address | UPF_PA_ADDRESS | --paaddr string | 127.0.0.1 |
| <br>`Дополнительный` | Трассировка сообщений ассоциации PFCP (Установить/Изменить/Освободить) | trace_association | UPF_TRACE_ASSOCIATION | --traceassoc | TRUE |
| <br>`Дополнительный` | Трассировка сброшенных пакетов dataplane | trace_blocked | UPF_TRACE_BLOCKED | --traceblock | TRUE |
| <br>`Дополнительный` | Максимальное количество ротируемых файлов дампа трассировки | trace_files | UPF_TRACE_FILES | --tracefcnt int | 10 |
| <br>`Дополнительный` | Максимальное количество пакетов в одном файле дампа трассировки | trace_max_packets | UPF_TRACE_MAX_PACKETS | --tracefpackets int | 100000 |
| <br>`Дополнительный` | Максимальный размер (в байтах) одного файла дампа трассировки | trace_max_size | UPF_TRACE_MAX_SIZE | --tracefsize int | 10485760 |
| <br>`Дополнительный` | Трассировка сообщений Heartbeat PFCP | trace_heartbeat | UPF_TRACE_HEARTBEAT | --tracehb | ЛОЖЬ |
| <br>`Дополнительный` | Включение или отключение поддержки объявлений маршрутизатора IPv6 | ip6_ra_support | UPF_IP6_RA_SUPPORT | --ip6ra | ЛОЖЬ |
| <br>`Дополнительный` | Префикс IPv6 абонента | ip6_ra_prefix | UPF_IP6_RA_PREFIX | --ip6rapref string | 2a03:d000:29a0:509::/64 |

Мы использьуем [Viper](https://github.com/spf13/viper) для работы с конфигурациями, [Viper](https://github.com/spf13/viper) использует следующий порядок приоритета. Каждый элемент имеет приоритет над элементом, находящимся под ним:

- аргумент CLI
- переменная окружения
- значение в конфигурационных файлах
- значение по умолчанию

*ЗАМЕЧАНИЕ:* [commit](https://github.com/edgecomllc/eupf/commit/ea56431df2f74cb2eabe85052d8762fe95848711) на текущий момент мы поддерживаем только IPv4 NodeID.

## Примеры конфигураций

### Значения по умолчанию  YAML

```yaml
interface_name: [lo]
xdp_attach_mode: generic
api_address: :8080
pfcp_address: :8805
pfcp_node_id: 127.0.0.1
metrics_address: :9090
n3_address: 127.0.0.1
n9_address: 127.0.0.1
qer_map_size: 1024
far_map_size: 1024
pdr_map_size: 1024
feature_ftup: true
teid_pool: 65536
```

### Переменные окружения

```env
UPF_INTERFACE_NAME="[eth0, n6]"
UPF_XDP_ATTACH_MODE=generic
UPF_API_ADDRESS=:8081
UPF_PFCP_ADDRESS=:8806
UPF_METRICS_ADDRESS=:9091
UPF_PFCP_NODE_ID: 10.100.50.241  # address on n4 interface
UPF_N3_ADDRESS: 10.100.50.233
UPF_N9_ADDRESS: 10.100.50.233
```

### CLI

```bash
eupf \
 --iface n3 \
 --iface n6 \
 --attach generic \
 --aaddr :8081 \
 --paddr :8086 \
 --nodeid 127.0.0.1 \
 --maddr :9090 \
 --n3addr 10.100.50.233
```
