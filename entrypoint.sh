#!/bin/env sh

# Mount bpffs if not shared already
if [[ $(/bin/mount | /bin/grep /sys/fs/bpf -c) -eq 0 ]]; then
    /bin/mount bpffs /sys/fs/bpf -t bpf;
fi
# Run app replacing current shell (to preserve signal handling) and forward cmd arguments
exec /app/bin/eupf "$@"
