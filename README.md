# Goche

Goche - it's fast, simple, zero dependency and well-tested Go package for create in-memory cache with generics.

[![Go Report Card](https://goreportcard.com/badge/github.com/romsar/goche)](https://goreportcard.com/report/github.com/romsar/goche)

<img width="300" alt="image" src="https://user-images.githubusercontent.com/27378369/207821452-0c8750bd-e4da-408d-83cf-5ae1f9708b73.png" />

## Benchmarks

<img width="861" alt="image" src="https://user-images.githubusercontent.com/27378369/207714000-0716f854-27ed-4013-ad1a-23761b0b60f3.png">

## Usage

### Basic

```golang
ctx := context.Background()
c := goche.New[string, string](
    // goche.WithPollInterval[string, string](3 * time.Second),
    // goche.WithSize[string, string](5),
    // goche.WithValues[string, string](map[string]string{"foo":"bar"}),
    // goche.WithDefaultTTL[string, string](5 * time.Second),
)
go c.Run(ctx)

c.Set("foo", "bar")

val, ok := c.Get("foo") // val == "bar"
if !ok {
    fmt.Println("not found!")
}

c.Delete("foo")
```

### TTL
```golang
ctx := context.Background()
c := goche.New[string, string]()
go c.Run(ctx)

c.Set("foo", "bar", TTL[string](10 * time.Second))
```

### TTL with reset

```golang
ctx := context.Background()
c := goche.New[string, string]()
go c.Run(ctx)

c.Set("foo", "bar", TTLWithReset[string](10 * time.Second))

// now, everytime when we do c.Get - ttl is resetting.
```