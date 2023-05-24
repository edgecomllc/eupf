package main

type mappingTableEntry struct {
	BucketId   uint64
	OriginalId uint32
}

type IdTranslator struct {
	GlobalMappingTable  map[uint32]mappingTableEntry
	BucketMappingTables map[uint64]map[uint32]uint32
}

func (t *IdTranslator) GetId(bucketId uint64, originalId uint32) uint32 {
	if val, exists := t.BucketMappingTables[bucketId][originalId]; exists {
		return val
	} else {
		newId := uint32(0)
		// Check if the bucket exists. If not, create a new one.
		if t.BucketMappingTables[bucketId] == nil {
			t.BucketMappingTables[bucketId] = make(map[uint32]uint32)
		}
		// We have relatively few IDs, so linear search is fine
		for ; ; newId++ {
			if _, exists := t.GlobalMappingTable[newId]; !exists {
				break
			}
		}
		t.GlobalMappingTable[newId] = mappingTableEntry{bucketId, originalId}
		t.BucketMappingTables[bucketId][originalId] = newId
		return newId
	}
}

func (t *IdTranslator) RemoveId(bucketId uint64, originalId uint32) uint32 {
	translatedId := t.GetId(bucketId, originalId)
	delete(t.GlobalMappingTable, translatedId)
	delete(t.BucketMappingTables[bucketId], originalId)
	return translatedId
}

func (t *IdTranslator) RemoveGlobalId(id uint32) (uint64, uint32) {
	entry := t.GlobalMappingTable[id]
	delete(t.GlobalMappingTable, id)
	delete(t.BucketMappingTables[entry.BucketId], entry.OriginalId)
	return entry.BucketId, entry.OriginalId
}

func NewIdTranslator() *IdTranslator {
	return &IdTranslator{
		GlobalMappingTable:  make(map[uint32]mappingTableEntry),
		BucketMappingTables: make(map[uint64]map[uint32]uint32),
	}
}
