# eUPF API Documentation

In addition to prometheus metrics the eUPF API provides a set of endpoints for monitoring the User Plane Function (UPF). It includes endpoints for listing UPF pipeline, QER map content, PFCP associations, displaying configuration, and displaying XDP statistics. This API is built with the Gin Web Framework and provides a Swagger API documentation for easy exploration and testing.

## Endpoints

| Method | URL                              | Description                                                                     | Example                          |
|--------|----------------------------------|---------------------------------------------------------------------------------|----------------------------------|
| `GET`  | `/api/v1/upf_pipeline`           | Lists the UPF pipeline. Returns contents of upf pipeline (loaded ebpf programs) | `/api/v1/upf_pipeline`           |
| `GET`  | `/api/v1/xdp_stats`              | Displays the XDP statistics. Returns an object of `XdpStats`                    | `/api/v1/xdp_stats`              |
| `GET`  | `/api/v1/packet_stats`           | Displays the PACKET statistics. Returns an object of `PacketStats`              | `/api/v1/packet_stats`           |
| `GET`  | `/api/v1/qer_map`                | Lists the QER map content. Returns a list of `QerMapElement`                    | `/api/v1/qer_map`                |
| `GET`  | ` /api/v1/qer_map/:id`           | Get QER map element by id. Returns an object of `QerMapElement`                 | `/api/v1/qer_map/:id`            |
| `PUT`  | `/api/v1/qer_map/:id`            | Set values for QER map element by id. Returns a list of `QerMapElement`         | `/api/v1/qer_map/:id`            |
| `GET`  | `/api/v1/config`                 | Displays the configuration. Returns an object of `UpfConfig`                    | `GET /api/v1/config`             |
| `POST` | `/api/v1/config`                 | Set configuration values                                                        | `/api/v1/config`                 |
| `GET`  | `/api/v1/uplink_pdr_map/:id`     | Get Uplink PDR values by id. Returns an object of `PdrElement`                  | `/api/v1/uplink_pdr_map/:id`     |                       |
| `PUT`  | `/api/v1/uplink_pdr_map/:id`     | Set Uplink PDR values by id. Returns a new object of `PdrElement`               | `/api/v1/uplink_pdr_map/:id`     |
| `GET`  | ` /api/v1/far_map/:id`           | Get FAR map element by id. Returns an object of `FarMapElement`                 | `/api/v1/far_map/:id`            |
| `PUT`  | `/api/v1/far_map/:id`            | Set values for FAR map element by id. Returns a new object of `FarMapElement`   | `/api/v1/far_map/:id`            |
| `GET`  | ` /api/v1/pfcp_associations`     | Lists the PFCP associations. Returns an object of `NodeAssociationMap`          | `/api/v1/pfcp_associations`      |
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