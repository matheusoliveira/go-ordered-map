package main

import (
	"fmt"
	"sync"
	"runtime"

	"github.com/matheusoliveira/go-ordered-map/omap"
)

const nValues = 1000
const step = 200

func putSequence(m omap.OMap[int, int], from, to int, wg *sync.WaitGroup) {
	for i := from; i < to; i++ {
		m.Put(i, i * i)
	}
	wg.Done()
}

func play(m omap.OMap[int, int]) {
	var wg sync.WaitGroup
	wg.Add(nValues / step)
	for i := 0; i < nValues; i += step {
		go putSequence(m, i, i + step, &wg)
	}
	wg.Wait()
}

func validate(m omap.OMap[int, int]) {
	cnt := 0
	for it := m.Iterator(); it.Next(); {
		cnt++
	}
	if cnt != nValues {
		fmt.Printf("%T - expected %d values, found %d\n", m, nValues, cnt)
	} else {
		fmt.Printf("%T is OK!\n", m)
	}
}

func main() {
	if runtime.GOMAXPROCS(0) == 1 {
		runtime.GOMAXPROCS(2)
	}
	// sync
	fmt.Println("OMapSync")
	syncMap := omap.NewOMapSync[int, int]()
	play(syncMap)
	validate(syncMap)
	// non-sync
	// will (probably) throw an unrecoverable fatal error ("concurrent map read and map write" or "concurrent map map write")
	fmt.Println("OMapLinked")
	nonSyncMap := omap.NewOMapLinked[int, int]()
	play(nonSyncMap)
	validate(nonSyncMap)
}
