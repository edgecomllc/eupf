# eUPF API Documentation

In addition to prometheus metrics the eUPF API provides a set of endpoints for monitoring the User Plane Function (UPF). It includes endpoints for listing UPF pipeline, QER map content, PFCP associations, displaying configuration, and displaying XDP statistics. This API is built with the Gin Web Framework and provides a Swagger API documentation for easy exploration and testing.

## Endpoints

Method | URL                    | Description                                                            | Example
------ | ---------------------- | ---------------------------------------------------------------------- | ----------------------
`GET`  | `/api/v1/xdp_stats`    | Displays the XDP statistics. Returns an object of `XdpStats`           | `/api/v1/xdp_stats`
`GET`  | `/api/v1/packet_stats` | Displays the PACKET statistics. Returns an object of `PacketStats`     | `/api/v1/packet_stats`
`GET`  | `/api/v1/route_stats`  | Display FIB route lookup statistics. Returns an object of `RouteStats` | `/api/v1/route_stats`

### - QER map

Method | URL                    | Description                                                                | Example
------ | ---------------------- | -------------------------------------------------------------------------- | -------------------
`GET`  | `/api/v1/qer_map`      | Lists the QER map content. Returns a list of `QerMapElement`               | `/api/v1/qer_map`
`GET`  | `/api/v1/qer_map/{id}` | Get QER map element by id. Returns an object of `QerMapElement`            | `/api/v1/qer_map/1`
`PUT`  | `/api/v1/qer_map/{id}` | Set values for QER map element by id. Returns an object of `QerMapElement` | `/api/v1/qer_map/1`

[PUT] Example request body:

```json
{
  "id":0,
  "gate_status_ul": 0,
  "gate_status_dl": 0,
  "qfi": 0,
  "max_bitrate_ul": 200000000,
  "max_bitrate_dl": 100000000
}
```

### - Config

Method | URL                                  | Description                                                               | Example
------ | ------------------------------------ | ------------------------------------------------------------------------- | ------------------------------------
`GET`  | `/api/v1/config`                     | Displays the configuration. Returns an object of `UpfConfig`              | `/api/v1/config`
`POST` | `/api/v1/config/logging_level`       | Set logging level (e.g., `debug`, `info`, `warn`, `error`).               | `/api/v1/config/logging_level`
`POST` | `/api/v1/config/logging_caller`      | Enable or disable displaying caller info in logs.                         | `/api/v1/config/logging_caller`
`POST` | `/api/v1/config/dataplane_ebpf`      | Bind dataplane interfaces and XDP attach mode.                            | `/api/v1/config/dataplane_ebpf`
`POST` | `/api/v1/config/dataplane_addresses` | Update N3/N9 interface IP addresses.                                      | `/api/v1/config/dataplane_addresses`
`POST` | `/api/v1/config/pfcp_n4`             | Update N4 PFCP connection (PFCP address, node ID, and remote nodes list). | `/api/v1/config/pfcp_n4`
`POST` | `/api/v1/config/pfcp_sxa`            | Update Sxa PFCP connection (address, node ID, and remote nodes list).     | `/api/v1/config/pfcp_sxa`
`POST` | `/api/v1/config/pfcp_sxb`            | Update Sxb PFCP connection (address, node ID, and remote nodes list).     | `/api/v1/config/pfcp_sxb`
`POST` | `/api/v1/config/pfcp_timers`         | Update PFCP heartbeat and association setup timers.                       | `/api/v1/config/pfcp_timers`
`POST` | `/api/v1/config/gtp_path`            | Update GTP peers and GTP echo request interval.                           | `/api/v1/config/gtp_path`

[POST] Example request body:

`/api/v1/config/logging_level`

```json
{
  "logging_level": "info"
}
```

`/api/v1/config/logging_caller`

```json
{
  "logging_caller": true
}
```

`/api/v1/config/dataplane_ebpf`

```json
{
  "interface_name": [
    "eth0",
    "eth1"
  ],
  "xdp_attach_mode": "generic"
}
```

`/api/v1/config/dataplane_addresses`

```json
{
  "n3_address": "10.100.200.14",
  "n9_address": "10.100.200.14"
}
```

`/api/v1/config/pfcp_n4`

```json
{
  "pfcp_address": "10.100.200.14:8805",
  "pfcp_node_id": "n4node1",
  "pfcp_remote_node": [
    "10.100.200.15",
    "10.100.200.16"
  ]
}
```

`/api/v1/config/pfcp_sxa`

```json
{
  "sxa_address": "10.100.200.14:8805",
  "sxa_node_id": "sxanode1",
  "sxa_remote_node": [
    "10.100.200.17",
    "10.100.200.18"
  ]
}
```

`/api/v1/config/pfcp_sxb`

```json
{
  "sxb_address": "10.100.200.14:8805",
  "sxb_node_id": "sxbnode1",
  "sxb_remote_node": [
    "10.100.200.19",
    "10.100.200.20"
  ]
}
```

`/api/v1/config/pfcp_timers`

```json
{
  "association_setup_timeout": 10,
  "heartbeat_timeout": 5
}
```

`/api/v1/config/gtp_path`

```json
{
  "gtp_peer": [
    "10.100.200.30",
    "10.100.200.31"
  ],
  "gtp_echo_interval": 10
}
```

### - Uplink PDR

Method | URL                           | Description                                                         | Example
------ | ----------------------------- | ------------------------------------------------------------------- | --------------------------
`GET`  | `/api/v1/uplink_pdr_map/{id}` | Get Uplink PDR values by TEID. Returns an object of `PdrElement`    | `/api/v1/uplink_pdr_map/1`
`PUT`  | `/api/v1/uplink_pdr_map/{id}` | Set Uplink PDR values by TEID. Returns a new object of `PdrElement` | `/api/v1/uplink_pdr_map/1`

[PUT] Example request body:

```json
{
  "teid": 2,
  "outer_header_removal": 0,
  "far_id": 1,
  "qer_id": 1,
  "trace": true
}
```

### - Downlink PDR

Method | URL                             | Description                                                                  | Example
------ | ------------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------------------------------
`GET`  | `/api/v1/downlink_pdr_map/{id}` | Get Uplink PDR values by IPv4 or IPv6\. Returns an object of `PdrElement`    | `/api/v1/downlink_pdr_map/2a03:d000:29a0:0509:0001:0000:1f86:9ec4`
`PUT`  | `/api/v1/downlink_pdr_map/{id}` | Set Uplink PDR values by IPv4 or IPv6\. Returns a new object of `PdrElement` | `/api/v1/downlink_pdr_map/10.45.0.2`

[PUT] Example request body:

```json
{
  "ip": "2a03:d000:29a0:509:1:0:1f86:9ec4",
  "outer_header_removal": 0,
  "far_id": 1,
  "qer_id": 1,
  "trace": true
}
```

### - FAR map

Method | URL                    | Description                                                                   | Example
------ | ---------------------- | ----------------------------------------------------------------------------- | -------------------
`GET`  | `/api/v1/far_map/{id}` | Get FAR map element by id. Returns an object of `FarMapElement`               | `/api/v1/far_map/1`
`PUT`  | `/api/v1/far_map/{id}` | Set values for FAR map element by id. Returns a new object of `FarMapElement` | `/api/v1/far_map/1`

[PUT] Example request body:

```json
{
  "action": 0,
  "outer_header_creation": 0,
  "teid": 0,
  "remote_ip": 0,
  "transport_level_marking": 0
}
```

### - PFCP associations

Method | URL                              | Description                                                                 | Example
------ | -------------------------------- | --------------------------------------------------------------------------- | --------------------------------
`GET`  | `/api/v1/pfcp_associations`      | Lists the PFCP associations. Returns a list of `NodeAssociationDescription` | `/api/v1/pfcp_associations`
`GET`  | `/api/v1/pfcp_associations/full` | Lists the full PFCP associations. Returns objects of `NodeAssociation`      | `/api/v1/pfcp_associations/full`
`GET`  | `/api/v1/pfcp_associations/release` | Calls for PFCP Association Release Request for all established associations. Returns nothing      | `/api/v1/pfcp_associations/release`

### - PFCP sessions

Method   | URL                     | Description                                                                                                                                                | Example
-------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | -----------------------------------------------
`GET`    | `/api/v1/pfcp_sessions` | Lists the PFCP sessions. If no parameters are given, returns all sessions. If ip or teid is given, returns filtered sessions. Returns a list of `Session`. | `/api/v1/pfcp_sessions?ip=192.168.1.1&teid=123`
`DELETE` | `/api/v1/pfcp_sessions` | Deletes a PFCP session by IMSI, MSISDN or session ID. At least one parameter required. Returns status message.                                             | `/api/v1/pfcp_sessions?imsi=123456789012345`

### - Subscriber tracing

Method   | URL                        | Description                                                                                                                     | Example
-------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | -----------------------------------------------
`GET`    | `/api/v1/subscriber_trace` | Lists active subscribers for trace filtered by IMSI or MSISDN. Returns a list of `TraceRecord`                                  | `/api/v1/subscriber_trace`
`POST`   | `/api/v1/subscriber_trace` | Set subscriber for tracing by its IMSI or MSISDN. Repeated request for already existed subscriber is not considered as an error | `/api/v1/subscriber_trace?imsi=255018600005299`
`DELETE` | `/api/v1/subscriber_trace` | Delete subscriber from tracing by its IMSI or MSISDN. Delete all subscribers if not params are provided                         | `/api/v1/subscriber_trace?msisdn=88095983595`

## Swagger API Documentation

To explore and test the API, you can use the Swagger API documentation. To access the Swagger UI, navigate to the following endpoint in your browser:

- GET /swagger/index.html

## API docs generation

[Reference documentation](https://github.com/swaggo/gin-swagger)

```bash
go install github.com/swaggo/swag/cmd/swag@v1.8.12
cd {repo_root}/cmd/eupf
swag init --parseDependency
```
