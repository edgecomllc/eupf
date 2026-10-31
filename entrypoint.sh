#!/bin/env sh

# Mount bpffs and debugfs if not present already. Only root can mount: as a
# non-root user the node has to provide bpffs at /sys/fs/bpf (the chart's
# README says how).
if [ "$(id -u)" -eq 0 ]; then
    if [[ $(/bin/mount | /bin/grep /sys/fs/bpf -c) -eq 0 ]]; then
        /bin/mount bpffs /sys/fs/bpf -t bpf;
    fi
    if [[ $(/bin/mount | /bin/grep /sys/kernel/debug -c) -eq 0 ]]; then
        /bin/mount debugfs /sys/kernel/debug -t debugfs;
    fi
fi

# Run app replacing current shell (to preserve signal handling) and forward cmd arguments
exec /app/bin/eupf "$@"
