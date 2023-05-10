# eUPF API Documentation

In addition to prometheus metrics the eUPF API provides a set of endpoints for monitoring the User Plane Function (UPF). It includes endpoints for listing UPF pipeline, QER map content, PFCP associations, displaying configuration, and displaying XDP statistics. This API is built with the Gin Web Framework and provides a Swagger API documentation for easy exploration and testing.

## Endpoints

- **GET /api/v1/upf_pipeline**: Lists the UPF pipeline.
    - Returns contents of upf pipeline (loaded ebpf programs)
    - Example: `GET /api/v1/upf_pipeline`

- **GET /api/v1/qer_map**: Lists the QER map content.
    - Returns a list of `QerMapElement`.
    - Example: `GET /api/v1/qer_map`

- **GET /api/v1/pfcp_associations**: Lists the PFCP associations.
    - Returns an object of `NodeAssociationMap`.
    - Example: `GET /api/v1/pfcp_associations`

- **GET /api/v1/config**: Displays the configuration.
    - Returns an object of `UpfConfig`.
    - Example: `GET /api/v1/config`

- **GET /api/v1/xdp_stats**: Displays the XDP statistics.
    - Returns an object of `XdpStats`.
    - Example: `GET /api/v1/xdp_stats`

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