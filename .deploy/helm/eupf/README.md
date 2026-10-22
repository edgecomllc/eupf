# eUPF Helm Chart

Deploys [eUPF](https://github.com/edgecomllc/eupf) — an open-source 5G User Plane Function built on eBPF.

## Quick start

```bash
helm install eupf .deploy/helm/eupf
```

The default values produce a working pod: eUPF is configured via `UPF_*` environment variables (see below), probes and resource requests/limits are set, and the API is served on port 8080.

## Cluster prerequisites

- **Linux kernel >= 5.15** on the nodes (eBPF/XDP requirement).
- **Unsafe sysctl allowed**: the chart sets `net.ipv4.ip_forward=1` in the pod security context, which nodes must allow:

  ```
  kubelet --allowed-unsafe-sysctls 'net.ipv4.ip_forward'
  ```

- **Privileges**: eUPF loads eBPF programs, mounts bpffs/debugfs (via the image entrypoint) and raises `RLIMIT_MEMLOCK`. By default the chart gives the container the capabilities this takes and nothing more:

  ```yaml
  securityContext:
    capabilities:
      add: [NET_ADMIN, SYS_ADMIN, SYS_RESOURCE, BPF, PERFMON]
  ```

  The container keeps the device cgroup, the masked `/proc` and `/sys` paths and the runtime's seccomp and AppArmor profiles. Some environments deny the entrypoint's mounts even with `SYS_ADMIN`: nodes whose AppArmor profile forbids `mount` (the runtime's default profile on Ubuntu and Debian does), a custom seccomp profile that blocks `mount` or `bpf`, or a runtime that rejects `BPF` and `PERFMON` on a kernel before 5.8. There, either allow the mounts (for AppArmor: `securityContext.appArmorProfile: {type: Unconfined}` on Kubernetes 1.30+, the `container.apparmor.security.beta.kubernetes.io/eupf: unconfined` pod annotation before) or fall back to a privileged container:

  ```yaml
  securityContext:
    privileged: true
  ```

- **Non-root** (optional): the image carries `NET_ADMIN`, `SYS_RESOURCE`, `BPF` and `PERFMON` as file capabilities on the `eupf` binary, plus `DAC_OVERRIDE` so it can create its map pin directory under `/sys/fs/bpf`, which root owns. With them the container runs as a non-root user, with less privilege than root plus capabilities: `SYS_ADMIN` is not needed, since a non-root entrypoint mounts nothing and the kernels the chart supports have `CAP_BPF`:

  ```yaml
  securityContext:
    runAsUser: 1000
    runAsGroup: 1000
    runAsNonRoot: true
    capabilities:
      add: [NET_ADMIN, SYS_RESOURCE, BPF, PERFMON, DAC_OVERRIDE]
  ```

  Grant exactly that set, as root or not: a binary with file capabilities fails to exec (`Operation not permitted`) when one of them is missing from the container's bounding set. The chart's defaults do; `docker run` needs `--privileged` or `--cap-add` for each. Leave `allowPrivilegeEscalation` unset: file capabilities need it. A non-root process cannot mount bpffs itself, so the **node must have bpffs mounted at `/sys/fs/bpf`**, as systemd hosts do (`sys-fs-bpf.mount`). kind nodes don't; bind the host's in with `extraMounts: [{hostPath: /sys/fs/bpf, containerPath: /sys/fs/bpf}]` on every node.

- The `/sys` hostPath is mounted read-only into the pod: as root the entrypoint mounts bpffs/debugfs under it, as non-root eUPF reaches the node's bpffs through it.

## Configuration

eUPF supports configuration via CLI flags, environment variables and a YAML config file (see [docs/Configuration.md](https://github.com/edgecomllc/eupf/blob/main/docs/Configuration.md)). Precedence: CLI > **env** > config file > defaults.

The chart supports two modes:

### ENV mode (default, `config.enabled: false`)

The chart injects a working default set of `UPF_*` variables (see `defaultEnv` in `values.yaml`), where `UPF_PFCP_NODE_ID` and `UPF_N3_ADDRESS` default to the pod IP. Override any of them via the `env` map:

```yaml
env:
  UPF_INTERFACE_NAME: eth0,n6   # comma-separated, no brackets
  UPF_LOGGING_LEVEL: debug
```

List-valued parameters (e.g. `UPF_INTERFACE_NAME`) are comma-separated strings.

### Sessions and memory

eUPF sizes its eBPF maps from `max_sessions`: two PDR, FAR and URR entries and one QER entry per session, allocated at startup and charged to the pod's memory. eUPF's own default is 65535 sessions, whose maps need about 115 MiB (a working set of about 140 MiB with the process, measured on a kind cluster); the chart's default memory limit is 128 Mi, so `defaultEnv` sets `UPF_MAX_SESSIONS: "1024"`, a working set of about 30 MiB. For production, raise `max_sessions` and `resources.limits.memory` together: about 1.8 MiB per 1024 sessions on top of the process's ~30 MiB, plus headroom; in ConfigMap mode set `max_sessions` in `config.data`, because the default `UPF_*` variables are not injected there.

### ConfigMap mode (`config.enabled: true`)

```yaml
config:
  enabled: true
  data:
    config.yml: |
      interface_name: eth0
      api_address: :8080
      pfcp_address: :8805
      metrics_address: :9090
```

The chart creates a ConfigMap from `config.data`, mounts it at `/app/conf` and starts the container with `--config /app/conf/config.yml`. In this mode the default `UPF_*` env injection is **disabled**, because environment variables take precedence over the config file. You can still add env vars as targeted overrides via `env`.

### Ports

| Port | Name | Purpose |
|------|------|---------|
| 8080 | http | REST API (`service.port`, must match `UPF_API_ADDRESS` / `api_address`) |
| 9090 | metrics | Prometheus metrics |

Probes are TCP checks against the `http` port; keep `service.port` in sync with the API address if you change it.
