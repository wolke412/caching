# Caching
## A thread-safe int-map for golang


### Installation

```bash
go get github.com/wolke412/caching
```


### Basic Usage
```go

const ExampleKey = 0x1 
const InvalidKey = 0x42 

//| Could be any type:         
//|                            v
var cache := caching.NewCache[int]{}

cache.Add( ExampleKey, 10 )

//| For now the cache looks like:
//+ ----------------------------------------
//| cache {
//|     1 -> 10
//| }


err := cache.Update( InvalidKey, 40 )
//|                 ^~~~~~~~~~~ 
//|                  not added first
//|                 
//| This will result in an error.
//| 'InvalidKey' (42) not found.

err := cache.Update( ExampleKey, 40 )
//| But this will work.

//| Now the cache looks like:
//+ ----------------------------------------
//| cache {
//|     1 -> 40
//| }
```

### Avoiding Race Condition
Although every call is protected:

```go
// Safely adds a key, or multiple keys
cache.Add(key int, value T)
cache.AddMultiple(entries map[int]T)

// Safely deletes one or all
cache.Delete(keys ...int) error
cache.Clear()

// Safely modify data
cache.Update(key int, value T) error
cache.UpdateWith(key int, cb cache.UpdateWithCallback[T]) error
cache.Upsert(key int, value T) error

// Safely access existing data 
cache.Exists(key int) bool
cache.Find(key int) (T, error)
cache.Get() *map[int]T
```

This is **NOT** enough to avoid race conditions
in all cases. 

For example:

```go
func addOne() {
    v, _ := caching.Find(ExampleKey) 

    // [ /!\ Danger Zone ]

    caching.Update(ExampleKey, v + 1 )
}
for range 1000 {
    go addOne() // reads 10, store 10 + 1
    go addOne() // reads 10, store 10 + 1
}
//| The expected result would be 2000.
//|
//| But you are a smart guy...
//|
//| You can see we first load the value, only then
//| its stored as value +1
//|
//| Since its composed of two actions, another thread
//| can surely sneak an update in between them. 
//| I even pointed the *Danger Zone*

```

So now you are wonder, well, why use this then...
For now, this is just a map with a mutex, right.
So the problem above is easily solvable thorough this:


```go
	cache.Lock()
	defer cache.Unlock() 

	v, _ := cache.UnsafeFind(Key_CachingDemo)
	        cache.UnsafeUpdate(Key_CachingDemo, v+1)
```

And now you wonder, **WHAT?** 

_To prevent race we use unsafe methods? Yeah great library man..._

The thing is, every call prefixed with `Unsafe`, is the 
true action... So no mutex locks inside there. Thats why,
when we are in a context which already holds the mutext, 
we should not call methods which will wait for them...

This would cause a...?
**You guessed it, nice!** A `DEADLOCK`.

Here is the list of unsafe methods:

```go
func (c *caching.Cache[T]) Lock()
func (c *caching.Cache[T]) Unlock()
func (c *caching.Cache[T]) UnsafeAdd(key int, value T)
func (c *caching.Cache[T]) UnsafeClear()
func (c *caching.Cache[T]) UnsafeDelete(keys ...int) error
func (c *caching.Cache[T]) UnsafeExists(key int) bool
func (c *caching.Cache[T]) UnsafeFind(key int) (T, error)
func (c *caching.Cache[T]) UnsafeGet() *map[int]T
func (c *caching.Cache[T]) UnsafeUpdate(key int, value T) error
func (c *caching.Cache[T]) UnsafeUpsert(key int, value T) error
```