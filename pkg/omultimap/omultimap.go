package omultimap

import (
	"github.com/matheusoliveira/go-ordered-map/pkg/omap"
)

type OMultiMap[K comparable, V any] interface {
	Put(key K, value V)
	GetValuesOf(key K) omap.OMapIterator[K, V]
	DeleteAll(key K)
	DeleteAt(omap.OMapIterator[K, V]) error
	MustDeleteAt(omap.OMapIterator[K, V])
	Iterator() omap.OMapIterator[K, V]
}
