# eUPF Command Line Interface Documentation

CLI tool is a separate application available to run as a carside to working eUPF service API. So, the CLI tool can be used remotely from anywhere there eUPF's API port accessible.
The tool has its built-in help 

```ruby
~# eupf -h
Usage of cli-flags:
      --backup_path string   Directory to store backups
      --eupf_addr string     Base URL for eupf API (e.g. http://localhost:8081/api/v1/)
      --log_level string     Log level (debug, info, warn, error)
eUPF CLI tool

Usage:
  eupf [command]

Available Commands:
  backup      Backup management
  completion  Generate shell completion scripts
  config      Configuration management
  help        Help about any command
  trace       Trace management

Flags:
  -h, --help   help for eupf

Use "eupf [command] --help" for more information about a command.
```

### Config

Configuration file `./config-cli.yaml` is possible to be created from scrutch using command <br>
`eupf config set-baseurl  --url http://127.0.0.1:8081/api/v1`

Default CLI tool config-cli.yaml is:

```yaml
backup_path: /var/lib/eupf/backups
eupf_addr: http://172.19.0.1:8082/api/v1/
log_level: debug
```

#### Configuration management

Usage:
  eupf config [command]

Available Commands:
-  set-baseurl  Set new base URL in config
-  show-baseurl Show current base URL from config

Use "eupf config [command] --help" for more information about a command.

#### Set a new base URL to be used as the default endpoint for eupf API.

This overrides the "eupf_addr" value in the config-cli.yaml file. <br>
Example:
-  eupf config set-baseurl --url http://localhost:8081/api/v1

Usage:
  eupf config set-baseurl [flags]

Flags:
-  -h, --help         help for set-baseurl
-    --url string   New base URL to save in config (e.g. http://localhost:8081/api/v1)

Example

```ruby
root@a2fc85d4d909:~# eupf config set-baseurl  --url http://172.19.0.1:8082/api/v1/
2025/05/22 06:49:31 INF CLI config file updated successfully: config-cli.yaml
Base URL updated successfully.
```

#### Display the currently configured default eupf API base URL,

which is stored in the config-cli.yaml file under "eupf_addr".

Usage:
  eupf config show-baseurl [flags]

Flags:
-  -h, --help   help for show-baseurl

Example

```ruby
root@a2fc85d4d909:~# eupf config show-baseurl
Current base URL: http://172.19.0.1:8082/api/v1/
```

### Backup/Restore current eUPF parameters

Usage:
-  eupf backup [command]

Available Commands:
-  create      Create backup
-  restore     Restore backup
-  show        Show backups files

#### Create backup configuration from current settings.

 Use --baseurl to override the eupf endpoint .

Usage:
-  eupf backup create [flags]

Flags:
-    --baseurl string   Optional base URL to override eupf API address (e.g. http://localhost:8081)
-  -h, --help             help for create

#### Retrieve list of backups.

Usage:
-  eupf backup show 

#### Restore backup configuration from given id.

 Use --baseurl to override the eupf endpoint.

ID of backup can be found from 'backup show' command. Example: 'eupf backup restore 1745307978'

Usage:
-  eupf backup restore <id of backup> [flags]

Flags:
-    --baseurl string   Optional base URL to override eupf API address (e.g. http://localhost:8081)

Example

```ruby
root@2ff8c95c6ecc:~# eupf backup create --baseurl http://172.19.0.1:8082/api/v1/
root@2ff8c95c6ecc:~# eupf backup show
#1: 1747753851 (20.05.25 15:10:51 UTC)
#2: 1746461057 (05.05.25 16:04:17 UTC)
#3: 1746460848 (05.05.25 16:00:48 UTC)
#4: 1746459629 (05.05.25 15:40:29 UTC)
root@2ff8c95c6ecc:~# 
root@2ff8c95c6ecc:~# ls -la /var/lib/eupf/backups/   
total 24
drwxrwxr-x 2 1000 1000 4096 May 20 15:10 .
drwxr-xr-x 3 root root 4096 May 20 05:28 ..
-rw-r--r-- 1 root root  422 May  5 15:40 1746459629.zip
-rw-r--r-- 1 root root  422 May  5 16:00 1746460848.zip
-rw-r--r-- 1 root root  545 May  5 16:04 1746461057.zip
-rw-r--r-- 1 root root  537 May 20 15:10 1747753851.zip
```

### Trace management

Packets dump pcap file content management.

Usage:
  eupf trace [command]

Available Commands:
-  set         Start trace for IMSI and/or MSISDN
-  show        Show current trace records
-  stop        Stop trace for IMSI and/or MSISDN

#### Start subscriber trace session. 

Use --imsi and/or --msisdn to filter, and --baseurl to override the eupf endpoint.

Usage:
  eupf trace set [flags]

Flags:
-  --baseurl string   Optional base URL to override eupf API
- -h, --help             help for set
-  --imsi string      IMSI to trace
-  --msisdn string    MSISDN to trace

#### Retrieve active trace records from the eupf API. 

Optionally use --baseurl to override target.

Usage:
  eupf trace show [flags]

Flags:
-    --baseurl string   Optional base URL to override eupf API address (e.g. http://localhost:8081)

Example: empty list returns error 404

```ruby
root@aeb8218ffaea:~# eupf trace show --baseurl http://172.19.0.1:8082/api/v1/
failed to list traces: trace list failed: 404 Not Found
```

#### Stop subscriber trace session. 

Use --imsi and/or --msisdn to match, and --baseurl to override the eupf endpoint.

Usage:
  eupf trace stop [flags]

Flags:
-    --baseurl string   Optional base URL to override eupf API
- -h, --help             help for stop
-    --imsi string      IMSI to stop trace
-    --msisdn string    MSISDN to stop trace

Example

```ruby
root@aeb8218ffaea:~# eupf trace set --imsi 250020000000016 --baseurl http://172.19.0.1:8082/api/v1/
Trace started (imsi=250020000000016, msisdn=)
root@aeb8218ffaea:~# eupf trace show --baseurl http://172.19.0.1:8082/api/v1/
{250020000000016  2025-05-22 20:22:02.686734276 +0300 +0300 }
{250020000000016  2025-05-22 20:22:02.686731014 +0300 +0300 }
{250020000000016  2025-05-22 20:22:02.686733069 +0300 +0300 }
root@aeb8218ffaea:~# eupf trace stop --baseurl http://172.19.0.1:8082/api/v1/
error: at least one of --imsi or --msisdn must be provided
root@aeb8218ffaea:~# eupf trace stop --imsi 250020000000016 --baseurl http://172.19.0.1:8082/api/v1/
Trace stopped (imsi=250020000000016, msisdn=)
root@aeb8218ffaea:~# eupf trace show --baseurl http://172.19.0.1:8082/api/v1/
failed to list traces: trace list failed: 404 Not Found
```

### Generate shell completion scripts

Usage:
  eupf completion 

```sh
mkdir -p /etc/bash_completion.d && \
    ./eupf completion > /etc/bash_completion.d/eupf

echo "source /etc/bash_completion" >> ~/.bashrc && \
    echo "source /etc/bash_completion.d/eupf" >> ~/.bashrc
```

Example: head of script

```ruby
root@aeb8218ffaea:~# eupf completion 
# bash completion for eupf                                 -*- shell-script -*-

__eupf_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE:-} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}
..........
```