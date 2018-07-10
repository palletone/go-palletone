package rwset

// RWSetBuilder helps building the read-write set
type RWSetBuilder struct {
	pubRwBuilderMap map[string]*nsPubRwBuilder
}

type nsPubRwBuilder struct {
	namespace         string
	readMap           map[string]*KVRead
	writeMap          map[string]*KVWrite
}

func NewRWSetBuilder() *RWSetBuilder {
	return &RWSetBuilder{make(map[string]*nsPubRwBuilder)}
}

// AddToReadSet adds a key and corresponding version to the read-set
func (b *RWSetBuilder) AddToReadSet(ns string, key string, version *Version) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	nsPubRwBuilder.readMap[key] = NewKVRead(key, version)
}

// AddToWriteSet adds a key and value to the write-set
func (b *RWSetBuilder) AddToWriteSet(ns string, key string, value []byte) {
	nsPubRwBuilder := b.getOrCreateNsPubRwBuilder(ns)
	nsPubRwBuilder.writeMap[key] = newKVWrite(key, value)
}

func (b *RWSetBuilder) getOrCreateNsPubRwBuilder(ns string) *nsPubRwBuilder {
	nsPubRwBuilder, ok := b.pubRwBuilderMap[ns]
	if !ok {
		nsPubRwBuilder = newNsPubRwBuilder(ns)
		b.pubRwBuilderMap[ns] = nsPubRwBuilder
	}
	return nsPubRwBuilder
}

func newNsPubRwBuilder(namespace string) *nsPubRwBuilder {
	return &nsPubRwBuilder{
		namespace,
		make(map[string]*KVRead),
		make(map[string]*KVWrite),
	}
}