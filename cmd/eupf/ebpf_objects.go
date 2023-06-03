package main

import (
	"github.com/edgecomllc/eupf/cmd/eupf/config"
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

	FarIdTracker         *IdTracker
	QerIdTracker         *IdTracker
	UplinkPdrIdTracker   *IdTracker
	DownlinkPdrIdTracker *IdTracker
}

func (bpfObjects *BpfObjects) Load() error {
	bpfObjects.FarIdTracker = NewIdTracker(config.Conf.FarMapSize)
	bpfObjects.QerIdTracker = NewIdTracker(config.Conf.QerMapSize)
	bpfObjects.UplinkPdrIdTracker = NewIdTracker(config.Conf.PdrMapSize)
	bpfObjects.DownlinkPdrIdTracker = NewIdTracker(config.Conf.PdrMapSize)
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

func ResizeEbpfMap(eMap **ebpf.Map, eProg *ebpf.Program, newSize uint32) error {
	mapInfo, err := (*eMap).Info()
	if err != nil {
		log.Printf("Failed get ebpf map info: %s", err)
		return err
	}
	mapInfo.MaxEntries = newSize
	// Create a new MapSpec using the information from MapInfo
	mapSpec := &ebpf.MapSpec{
		Name:       mapInfo.Name,
		Type:       mapInfo.Type,
		KeySize:    mapInfo.KeySize,
		ValueSize:  mapInfo.ValueSize,
		MaxEntries: mapInfo.MaxEntries,
		Flags:      mapInfo.Flags,
	}
	if err != nil {
		log.Printf("Failed to close old ebpf map: %s, %+v", err, *eMap)
		return err
	}

	// Unpin the old map
	err = (*eMap).Unpin()
	if err != nil {
		log.Printf("Failed to unpin old ebpf map: %s, %+v", err, *eMap)
		return err
	}

	// Close the old map
	err = (*eMap).Close()
	if err != nil {
		log.Printf("Failed to close old ebpf map: %s, %+v", err, *eMap)
		return err
	}

	// Old map will be garbage collected sometime after this point

	*eMap, err = ebpf.NewMapWithOptions(mapSpec, ebpf.MapOptions{})
	if err != nil {
		log.Printf("Failed to create resized ebpf map: %s", err)
		return err
	}
	err = eProg.BindMap(*eMap)
	if err != nil {
		log.Printf("Failed to bind resized ebpf map: %s", err)
		return err
	}
	return nil
}

func (bpfObjects *BpfObjects) ResizeAllMaps(qerMapSize uint32, farMapSize uint32, pdrMapSize uint32) error {
	// QEQ
	if err := ResizeEbpfMap(&bpfObjects.QerMap, bpfObjects.UpfQerProgramFunc, qerMapSize); err != nil {
		log.Printf("Failed to resize qer map: %s", err)
		return err
	}
	// FAR
	if err := ResizeEbpfMap(&bpfObjects.FarMap, bpfObjects.UpfFarProgramFunc, farMapSize); err != nil {
		log.Printf("Failed to resize far map: %s", err)
		return err
	}
	// PDR
	if err := ResizeEbpfMap(&bpfObjects.PdrMapDownlinkIp4, bpfObjects.UpfIpEntrypointFunc, pdrMapSize); err != nil {
		log.Printf("Failed to resize qer map: %s", err)
		return err
	}
	if err := ResizeEbpfMap(&bpfObjects.PdrMapDownlinkIp6, bpfObjects.UpfIpEntrypointFunc, pdrMapSize); err != nil {
		log.Printf("Failed to resize qer map: %s", err)
		return err
	}
	if err := ResizeEbpfMap(&bpfObjects.PdrMapUplinkIp4, bpfObjects.UpfIpEntrypointFunc, pdrMapSize); err != nil {
		log.Printf("Failed to resize qer map: %s", err)
		return err
	}

	return nil
}
