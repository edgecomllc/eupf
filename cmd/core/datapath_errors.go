package core

import (
	"errors"
	"strings"
	"syscall"
)

// isDataplaneKeyAbsent reports whether a datapath (eBPF map) delete error
// actually means "the entry is already absent" — which is the desired end
// state of a delete operation, not a failure.
//
// cilium/ebpf surfaces this as syscall.ENOENT or a "key does not exist"
// style error when the entry has already been removed (e.g. a delete that
// races with another cleanup path, or a retried deletion). Treating such
// errors as failures made the session-deletion handler return an error
// response — which, combined with the response-dropping bug in
// PfcpHandlerMap.Handle, led to PFCP Session Deletion Requests that were
// never answered at all (the CP function then times out its PFCP
// transaction; see issue #647).
func isDataplaneKeyAbsent(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.ENOENT) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "key does not exist") ||
		strings.Contains(msg, "no such file or directory")
}
