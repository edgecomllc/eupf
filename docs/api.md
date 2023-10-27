# eUPF API Documentation

In addition to prometheus metrics the eUPF API provides a set of endpoints for monitoring the User Plane Function (UPF). It includes endpoints for listing UPF pipeline, QER map content, PFCP associations, displaying configuration, and displaying XDP statistics. This API is built with the Gin Web Framework and provides a Swagger API documentation for easy exploration and testing.

## Endpoints

| Method | URL                              | Description                                                                     | Example                          |
|--------|----------------------------------|---------------------------------------------------------------------------------|----------------------------------|
| `GET`  | `/api/v1/upf_pipeline`           | Lists the UPF pipeline. Returns contents of upf pipeline (loaded ebpf programs) | `/api/v1/upf_pipeline`           |
| `GET`  | `/api/v1/xdp_stats`              | Displays the XDP statistics. Returns an object of `XdpStats`                    | `/api/v1/xdp_stats`              |
| `GET`  | `/api/v1/packet_stats`           | Displays the PACKET statistics. Returns an object of `PacketStats`              | `/api/v1/packet_stats`           |


#### - QER map

| Method | URL                             | Description                                                                     | Example             |
|--------|---------------------------------|---------------------------------------------------------------------------------|---------------------|
| `GET`  | `/api/v1/qer_map`               | Lists the QER map content. Returns a list of `QerMapElement`                    | `/api/v1/qer_map`   |
| `GET`  | `/api/v1/qer_map/:id`           | Get QER map element by id. Returns an object of `QerMapElement`                 | `/api/v1/qer_map/1` |
| `PUT`  | `/api/v1/qer_map/:id`           | Set values for QER map element by id. Returns a list of `QerMapElement`         | `/api/v1/qer_map/1` |

 [PUT] Example request body:

    {  
      "gate_status_ul": 0,
      "gate_status_dl": 0,
      "qfi": 0,
      "max_bitrate_ul": 200000000,
      "max_bitrate_dl": 100000000
    }

#### - Config

| Method | URL                              | Description                                                                     | Example                |
|--------|----------------------------------|---------------------------------------------------------------------------------|------------------------|
| `GET`  | `/api/v1/config`                 | Displays the configuration. Returns an object of `UpfConfig`                    | `/api/v1/config`       |
| `POST` | `/api/v1/config`                 | Set configuration values                                                        | `/api/v1/config`       |

 [POST] Example request body:

    {
      "interface_name": [
        "eth0",
        "eth1"
      ],
      "xdp_attach_mode": "generic",
      "api_address": "8080",
      "pfcp_address": "10.100.200.14:8805",
      "pfcp_node_id": "10.100.200.14",
      "metrics_address": ":9090",
      "n3_address": "10.100.200.14",
      "qer_map_size": 1024,
      "far_map_size": 1024,
      "pdr_map_size": 1024,
      "resize_ebpf_maps": false,
      "heartbeat_retries": 3,
      "heartbeat_interval": 5,
      "heartbeat_timeout": 5,
      "logging_level": "info"
    }



#### - Uplink PDR

| Method | URL                              | Description                                                                     | Example                    |
|--------|----------------------------------|---------------------------------------------------------------------------------|----------------------------|
| `GET`  | `/api/v1/uplink_pdr_map/:id`     | Get Uplink PDR values by id. Returns an object of `PdrElement`                  | `/api/v1/uplink_pdr_map/1` |              
| `PUT`  | `/api/v1/uplink_pdr_map/:id`     | Set Uplink PDR values by id. Returns a new object of `PdrElement`               | `/api/v1/uplink_pdr_map/1` |

 [PUT] Example request body:

    {
      "outer_header_removal": 0,
      "far_id": 0,
      "qer_id": 0
    }

#### - FAR map

| Method | URL                             | Description                                                                     | Example             |
|--------|---------------------------------|---------------------------------------------------------------------------------|---------------------|
| `GET`  | `/api/v1/far_map/:id`           | Get FAR map element by id. Returns an object of `FarMapElement`                 | `/api/v1/far_map/1` |
| `PUT`  | `/api/v1/far_map/:id`           | Set values for FAR map element by id. Returns a new object of `FarMapElement`   | `/api/v1/far_map/1` |

 [PUT] Example request body:

    {
      "action": 0,
      "outer_header_creation": 0,
      "teid": 0,
      "remote_ip": 0,
      "local_ip": 0,
      "transport_level_marking": 0
    }


#### - PFCP associations

| Method | URL                              | Description                                                                     | Example                          |
|--------|----------------------------------|---------------------------------------------------------------------------------|----------------------------------|
| `GET`  | `/api/v1/pfcp_associations`      | Lists the PFCP associations. Returns an object of `NodeAssociationMap`          | `/api/v1/pfcp_associations`      |
| `GET`  | `/api/v1/pfcp_associations/full` | Lists the full PFCP associations. Returns an object of `NodeAssociationMap`     | `/api/v1/pfcp_associations/full` |
| `GET`  | `/api/v1/pfcp_sessions`          | Lists the PFCP sessions content. Returns a list of `Session`                    | `/api/v1/pfcp_sessions`          |

## Swagger API Documentation

To explore and test the API, you can use the Swagger API documentation. To access the Swagger UI, navigate to the following endpoint in your browser:

- GET /swagger/index.html

## API docs generation 
(Reference documentation)[https://github.com/swaggo/gin-swagger]
```bash
go install github.com/swaggo/swag/cmd/swag@v1.8.12
cd {repo_root}/cmd/eupf
swag init --parseDependency
```