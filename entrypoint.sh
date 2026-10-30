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

# The eupf binary relies on file capabilities (effective bit). exec fails
# with EPERM (shell exit status 126) when any of them is missing from the
# container's bounding set. Docker's default set lacks NET_ADMIN,
# SYS_RESOURCE, BPF and PERFMON. Probe with a cheap flag so the failure is
# explained instead of a bare "Operation not permitted".
/app/bin/eupf --help >/dev/null 2>&1
status=$?
if [ "$status" -eq 126 ]; then
    echo "ERROR: cannot exec /app/bin/eupf: a required file capability is missing from the container's bounding set." >&2
    echo "The binary needs NET_ADMIN, SYS_RESOURCE, BPF, PERFMON and DAC_OVERRIDE (set as file capabilities on the binary)." >&2
    echo "With docker run, add: --cap-add=NET_ADMIN --cap-add=SYS_RESOURCE --cap-add=BPF --cap-add=PERFMON (or --privileged)." >&2
    echo "With Kubernetes, set securityContext.capabilities.add on the container; the eUPF Helm chart does this by default." >&2
    exit 126
fi

# Run app replacing current shell (to preserve signal handling) and forward cmd arguments
exec /app/bin/eupf "$@"
