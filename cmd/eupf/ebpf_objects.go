package main

import (
	"io"
	"log"
	"os"

	"github.com/cilium/ebpf"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf ip_entrypoint 	xdp/n3n6_entrypoint.c -- -I. -O2 -Wall -g
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf n3_entrypoint 	xdp/n3_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf n6_entrypoint 	xdp/n6_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf qer_program 		xdp/qer_program.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf far_program 		xdp/far_program.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf upf_xdp 			xdp/upf_program.c -- -I. -O2 -Wall

type BpfObjects struct {
	upf_xdpObjects
	far_programObjects
	qer_programObjects
	ip_entrypointObjects
}

func (bpfObjects *BpfObjects) Load() error {

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
		Loader{loadUpf_xdpObjects, &bpfObjects.upf_xdpObjects},
		Loader{loadFar_programObjects, &bpfObjects.far_programObjects},
		Loader{loadQer_programObjects, &bpfObjects.qer_programObjects},
		Loader{loadIp_entrypointObjects, &bpfObjects.ip_entrypointObjects})
}

func (bpfObjects *BpfObjects) Close() error {
	return CloseAllObjects(
		&bpfObjects.upf_xdpObjects,
		&bpfObjects.far_programObjects,
		&bpfObjects.qer_programObjects,
		&bpfObjects.ip_entrypointObjects,
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
