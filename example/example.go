package main

import (
	"caching"
	"fmt"
	"sync"
)

type DeviceConnectionData struct {
	Name string
}

var connected caching.Cache[*DeviceConnectionData]

const (
	Key_CachingDemo     = 1
	NumIter_CachingDemo = 1_000
)

// Can be any type of value
var cache *caching.Cache[int]

// !
// !
// !
func addOne(wg *sync.WaitGroup) {
	//
	// [i] Note that each call to cache.*() waits the mutex
	// ------------------------------------------------------------
	// Waits for mutex
	//             v
	v, _ := cache.Find(Key_CachingDemo)

	// Waits for mutex again
	//     v
	cache.Update(Key_CachingDemo, v+1)

	wg.Done()
	// ------------------------------------------------------------
	// ! So, althought this code may seem ATOMIC at first glange
	// ! another thread could come in between Find and Update
	// :
	//
	//  [Thread 1]    [LOCK]    cache.Find(KEY)         [RELEASE]
	//  [Thread 2]    [LOCK]    cache.Update(KEY, v+1)  [RELEASE]   <-/ Got ourselves a race condition...
	//  [Thread 1]    [LOCK]    cache.Update(KEY, v+1)  [RELEASE]     \ ... And a very slow one.
}

// !
// ! This is truly atomic multiple operations
// ! caching code.
// !
// ! We first lock the cache, then execute
// ! a BATCH of ~Unsafe operation.
// !
// ! The "Unsafe" prefix, denote that the functions
// ! does not acquire the mutex lock.
func atomicAddOne(wg *sync.WaitGroup) {

	// Starts locking the cache for the whole function
	cache.Lock()
	defer cache.Unlock() // We can defer the release to the end.

	// These operations do their thing
	// without acquiring the lock.
	v, _ := cache.UnsafeFind(Key_CachingDemo)
	cache.UnsafeUpdate(Key_CachingDemo, v+1)

	wg.Done()

	// The cache.Unlock() will happen here.
	// So other threads can acquire it.
}

func main() {
	cache = caching.NewCache[int]()

	provokeRaceCondition()
	confirmTrueAtomicity()
}

func confirmTrueAtomicity() {

	// Resets key
	cache.Update(Key_CachingDemo, 0)

	wg := sync.WaitGroup{}
	wg.Add(NumIter_CachingDemo)

	// No race condition
	for range NumIter_CachingDemo {
		go atomicAddOne(&wg)
	}

	wg.Wait()

	// Could handle a possible error if necessary.
	result, _ := cache.Find(Key_CachingDemo)

	fmt.Printf("EXPECT ATOMIC: Expected %d, got %d\n", NumIter_CachingDemo, result)
}

func provokeRaceCondition() {

	cache.Add(Key_CachingDemo, 0)

	wg := sync.WaitGroup{}
	wg.Add(NumIter_CachingDemo)

	// Race condition
	for range NumIter_CachingDemo {
		go addOne(&wg)
	}

	wg.Wait()

	result, _ := cache.Find(Key_CachingDemo)

	fmt.Printf("PROVOKED RACE: Expected %d, got %d\n", NumIter_CachingDemo, result)

}
