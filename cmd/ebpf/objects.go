package ebpf

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/RoaringBitmap/roaring"
	"github.com/edgecomllc/eupf/cmd/config"

	"github.com/cilium/ebpf"
)

//
// Supported BPF_CFLAGS:
// 	- ENABLE_LOG:
//		- enables debug output to tracepipe (`bpftool prog tracelog`)
// 	- ENABLE_ROUTE_CACHE
//		- enable routing decision cache
//

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cflags "$BPF_CFLAGS" -target bpf IpEntrypoint 	xdp/n3n6_entrypoint.c -- -I. -O2 -Wall -g
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf ZeroEntrypoint 	xdp/zero_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf N3Entrypoint 	xdp/n3_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf N6Entrypoint 	xdp/n6_entrypoint.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf QerProgram 		xdp/qer_program.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf FarProgram 		xdp/far_program.c -- -I. -O2 -Wall
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target bpf UpfXdp 			xdp/upf_program.c -- -I. -O2 -Wall

type BpfObjects struct {
	UpfXdpObjects
	FarProgramObjects
	QerProgramObjects
	IpEntrypointObjects

	FarIdTracker *IdTracker
	QerIdTracker *IdTracker
}

func NewBpfObjects() *BpfObjects {
	return &BpfObjects{
		FarIdTracker: NewIdTracker(config.Conf.FarMapSize),
		QerIdTracker: NewIdTracker(config.Conf.QerMapSize),
	}
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
			// program, so it can be re-used if it already exists or
			// create it if not
			PinPath: pinPath,
		},
	}

	return LoadAllObjects(&collectionOptions,
		Loader{LoadUpfXdpObjects, &bpfObjects.UpfXdpObjects},
		Loader{LoadFarProgramObjects, &bpfObjects.FarProgramObjects},
		Loader{LoadQerProgramObjects, &bpfObjects.QerProgramObjects},
		Loader{LoadIpEntrypointObjects, &bpfObjects.IpEntrypointObjects})
}

func (bpfObjects *BpfObjects) Close() error {
	return CloseAllObjects(
		&bpfObjects.UpfXdpObjects,
		&bpfObjects.FarProgramObjects,
		&bpfObjects.QerProgramObjects,
		&bpfObjects.IpEntrypointObjects,
	)
}

func (bpfObjects *BpfObjects) BuildPipeline() {
	upfPipeline := bpfObjects.UpfXdpObjects.UpfPipeline
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

type IdTracker struct {
	bitmap  *roaring.Bitmap
	maxSize uint32
}

func NewIdTracker(size uint32) *IdTracker {
	newBitmap := roaring.NewBitmap()
	newBitmap.Flip(0, uint64(size))

	return &IdTracker{
		bitmap:  newBitmap,
		maxSize: size,
	}
}

func (t *IdTracker) GetNext() (next uint32, err error) {

	i := t.bitmap.Iterator()
	if i.HasNext() {
		next := i.Next()
		t.bitmap.Remove(next)
		return next, nil
	}

	return 0, errors.New("pool is empty")
}

func (t *IdTracker) Release(id uint32) {
	if id >= t.maxSize {
		return
	}

	t.bitmap.Add(id)
}
