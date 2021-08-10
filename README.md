# StreamingFast Metrics Library
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/dmetrics)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

For now, this library contains simple wrapping structure around Prometheus
metrics. This improves usage and developer experience of defining metrics
for StreamingFast services.
It is part of **[StreamingFast](https://github.com/streamingfast/streamingfast)**.

This library should be kept as SMALL as possible, as it is a dependency
we want to sprinkle around widely.

## Usage

All metrics must be created from a metrics _Set_. The idea is that each
library and micro-services defines a single set of metrics.

All metrics to be published by Prometheus must be registered at some
point.

Example `metrics.go` in an example `bstream` project:

```go
// Exported, so it can be registered by a `metrics.go` in main packages.
var MetricsSet = dmetrics.NewSet()

// HitCount is exported if used by other packages elsewhere, use like `bstream.HitCount.Inc()`
var HitCount = MetricsSet.NewGauge("bstream_hit_count", "hello %s")

// myCount is not exported, because used only in here.  Use as: `myCount.Inc()`
var myCount = MetricsSet.NewGauge("bstream_my_count", "hello %s")
```

In a `main` package, in a `metrics.go` (similar to `logging.go`):

```go
func init() {
    dmetrics.Register(
        bstream.MetricsSet,
        dauth.MetricsSet,
        blockmeta.MetricsSet,
    )
}
```


## Background

Initially, we had defined our metrics directly as the Prometheus type giving
a definition of metrics in the form:

```go
var mapSize = newGauge(
	"map_size",
	"size of live blocks map",
)

func IncMapSize() {
	mapSize.Inc()
}

func DecMapSize() {
	mapSize.Dec()
}
```

The usage of this was then like this:

```go
    metrics.IncMapSize()

    ...

    metrics.DecMapSize()
```

This was repeated for all metrics then defined. This is problematic as when there is multiple
metrics, the source file for definitions becomes bloated with lots of repeated stuff and duplicated
stuff.

To overcome this, this library wraps different Prometheus metrics to clean down the
definitions file.
and offer a nicer
API around them , and also usage. The previous example
can now be turned into:

```go
var MapSize = dmetrics.NewGauge("map_size", "size of live blocks map")
```

And the usage is now like:

```go
    metrics.MapSize.Inc()

    ...

    metrics.MapSize.Dec()
```

An incredible improvement in the definitions of the metrics themselves.


## Contributing

**Issues and PR in this repo related strictly to the dmetrics library.**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.


## License

[Apache 2.0](LICENSE)
