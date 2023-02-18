package main

import (
	"io"
	"log"
	"os"

	"github.com/cilium/ebpf"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf ip_entrypoint 	xdp/ip_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf gtp_entrypoint 	xdp/gtp_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf qer_program 		xdp/qer_program.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf far_program 		xdp/far_program.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf upf_xdp 			xdp/upf_program.c -- -I. -O2 -Wall

type BpfObjects struct {
	upf_xdpObjects
	far_programObjects
	qer_programObjects
	ip_entrypointObjects
	gtp_entrypointObjects
}

func (o *BpfObjects) Load() error {

	pinPath := "/sys/fs/bpf/upf_pipeline"
	if err := os.MkdirAll(pinPath, os.ModePerm); err != nil {
		log.Printf("failed to create bpf fs subpath: %+v", err)
		return err
	}

	collectionOptions := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			// Pin the map to the BPF filesystem and configure the
			// library to automatically re-write it in the BPF
			// program so it can be re-used if it already exists or
			// create it if not
			PinPath: pinPath,
		},
	}

	return LoadAllObjects(&collectionOptions,
		Loader{loadUpf_xdpObjects, &o.upf_xdpObjects},
		Loader{loadFar_programObjects, &o.far_programObjects},
		Loader{loadQer_programObjects, &o.qer_programObjects},
		Loader{loadIp_entrypointObjects, &o.ip_entrypointObjects},
		Loader{loadGtp_entrypointObjects, &o.gtp_entrypointObjects})
}

func (o *BpfObjects) Close() error {
	return CloseAllObjects(
		&o.upf_xdpObjects,
		&o.far_programObjects,
		&o.qer_programObjects,
		&o.ip_entrypointObjects,
		&o.gtp_entrypointObjects,
	)
}

func (bpfObjects *BpfObjects) buildPipeline() {
	upfPipeline := bpfObjects.upf_xdpObjects.UpfPipeline
	upfMainProgram := bpfObjects.UpfFunc
	farProgram := bpfObjects.UpfFarProgramFunc
	qerProgram := bpfObjects.UpfQerProgramFunc

	if err := upfPipeline.Put(uint32(0), upfMainProgram); err != nil {
		panic(err)
	}

	if err := upfPipeline.Put(uint32(1), farProgram); err != nil {
		panic(err)
	}

	if err := upfPipeline.Put(uint32(2), qerProgram); err != nil {
		panic(err)
	}

}

type LoaderFunc func(obj interface{}, opts *ebpf.CollectionOptions) error
type Loader struct {
	LoaderFunc
	object interface{}
}

func LoadAllObjects(opts *ebpf.CollectionOptions, loaders ...Loader) error {
	for _, loader := range loaders {
		if err := loader.LoaderFunc(loader.object, opts); err != nil {
			return err
		}
	}
	return nil
}

func CloseAllObjects(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}
