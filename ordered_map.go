package lox

// OMap is a map sorted by insertion order.
type OMap[K comparable, V any] struct {
	m     map[K]V
	order []K
}

type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

func MakeOMap[K comparable, V any]() *OMap[K, V] {
	return &OMap[K, V]{
		m: make(map[K]V),
	}
}

func MakeOMap1[K comparable, V any](k1 K, v1 V) *OMap[K, V] {
	return &OMap[K, V]{
		m:     map[K]V{k1: v1},
		order: []K{k1},
	}
}

func MakeOMap2[K comparable, V any](k1 K, v1 V, k2 K, v2 V) *OMap[K, V] {
	return &OMap[K, V]{
		m:     map[K]V{k1: v1, k2: v2},
		order: []K{k1, k2},
	}
}

func MakeOMap3[K comparable, V any](k1 K, v1 V, k2 K, v2 V, k3 K, v3 V) *OMap[K, V] {
	return &OMap[K, V]{
		m:     map[K]V{k1: v1, k2: v2, k3: v3},
		order: []K{k1, k2, k3},
	}
}

func (omap *OMap[K, V]) Get(key K) (V, bool) {
	v, ok := omap.m[key]
	return v, ok
}

func (omap *OMap[K, V]) Put(key K, value V) {
	if _, ok := omap.m[key]; !ok {
		omap.order = append(omap.order, key)
	}
	omap.m[key] = value
}

func (omap *OMap[K, V]) Keys() []K {
	keys := make([]K, len(omap.order))
	copy(keys, omap.order)
	return keys
}

func (omap *OMap[K, V]) Values() []V {
	values := make([]V, len(omap.order))
	for i, key := range omap.order {
		values[i] = omap.m[key]
	}
	return values
}

func (omap *OMap[K, V]) Entries() []Entry[K, V] {
	entries := make([]Entry[K, V], len(omap.order))
	for i, key := range omap.order {
		entries[i] = Entry[K, V]{key, omap.m[key]}
	}
	return entries
}
