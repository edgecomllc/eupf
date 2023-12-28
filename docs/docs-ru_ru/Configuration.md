# Конфигурация UPF

## Описание 

В настоящее время UPF имеет несколько параметров конфигурации, показанных ниже.<br>Параметры можно настроить через интерфейс командной строки, файлы конфигурации (YAML, JSON) или переменные среды.

| Параметр                      | Описание                                                                                                                                                                                                                                                                                                                                     | yaml              | env                   | cli arg     | Значение по умолчанию    |
|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|-----------------------|-------------|-------------|
| Interface name<br>`Обязательный`  | Список сетевых интерфейсов, обрабатывающих трафик N3 (GTP) и N6 (SGi). eUPF присоединяет перехватчик XDP к каждому интерфейсу в этом списке. Format: `[ifnameA, ifnameB, ...]`.                                                                                                                                                                                     | `interface_name`  | `UPF_INTERFACE_NAME`  | `--iface`   | `lo`        |
| N3 address <br>`Обязательный`     | IPv4 адрея для N3 интерфейса                                                                                                                                                                                                                                                                                                                   | `n3_address`      | `UPF_N3_ADDRESS`      | `--n3addr`  | `127.0.0.1` |
| XDP mode <br>`Дополнительный`        | XDP attach mode: <br> ∘ **generic** – Реализация на уровне ядра. В целях оценки. <br> ∘ **native** – реализация на уровне драйвера <br> ∘ **offloaded** – реализация на уровне NIC. XDP можно загрузить и выполнить непосредственно на сетевой карте. <br> См. [Как работает XDP](https://www.tigera.io/learn/guides/ebpf/ebpf-xdp/#How-XDP-Works) | `xdp_attach_mode` | `UPF_XDP_ATTACH_MODE` | `--attach`  | `generic`   |
| API address <br>`Дополнительный`     | Локальный адрес для обслуживания сервера [REST API](../../docs/api.md)                                                                                                                                                                                                                                                                                              | `api_address`     | `UPF_API_ADDRESS`     | `--aaddr`   | `:8080`     |
| PFCP address <br>`Дополнительный`    | Локальный адрес, по которому буедт доступен PFCP server                                                                                                                                                                                                                                                                                                    | `pfcp_address`    | `UPF_PFCP_ADDRESS`    | `--paddr`   | `:8805`     |
| PFCP NodeID <br>`Дополнительный`     | Локальный NodeID для PFCP protocol. Формет -  IPv4 address.                                                                                                                                                                                                                                                                                         | `pfcp_node_id`    | `UPF_PFCP_NODE_ID`    | `--nodeid`  | `127.0.0.1` |
| Metrics address <br>`Дополнительный` | Локальный адрес для обслуживания метрик Prometheus.                                                                                                                                                                                                                                                                                         | `metrics_address` | `UPF_METRICS_ADDRESS` | `--maddr`   | `:9090`     |
| QER map size <br>`Дополнительный`    | Размер eBPF map для параметров QER                                                                                                                                                                                                                                                                                                          | `qer_map_size`    | `UPF_QER_MAP_SIZE`    | `--qersize` | `1024  `    |
| FAR map size <br>`Дополнительный`    | Размер eBPF map для параметров FAR                                                                                                                                                                                                                                                                                                         | `far_map_size`    | `UPF_FAR_MAP_SIZE`    | `--farsize` | `1024  `    |
| PDR map size <br>`Дополнительный`    | Размер eBPF map для параметров PDR                                                                                                                                                                                                                                                                                                         | `pdr_map_size`    | `UPF_PDR_MAP_SIZE`    | `--pdrsize` | `1024  `    |
| Logging level <br>`Дополнительный`   | Журналы уровня <= выбранного уровня будут записаны на stdout.                                                                                                                                                                                                                                                                                   | `logging_level`   | `UPF_LOGGING_LEVEL`   | `--loglvl`  | `info`         |
| FTUP Feature <br>`Дополнительный`    | Поддержка опции распределения TEID                                                                                                                                                                                                                                                                                                              | `feature_ftup`    | `UPF_FEATURE_FTUP`    | `--feature_ftup`          | `false`        |
| TEID Pool <br>`Дополнительный`       | Пул TEID, необходимый для выделения TEID, когда опция FTUP включена                                                                                                                                                                                                                                                                          | `teid_pool`       | `UPF_TEID_POOL`       | `--teid_pool`          | `65536`        |

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
