# OpenSlide Go (unofficial)

OpenSlide Go is a Golang interface to the [OpenSlide] library. It intends to wrap all OpenSlide C functions in idiomatic Golang.

[OpenSlide]: https://openslide.org/
[OpenSlide Python]: https://github.com/openslide/openslide-python

## Requirements

- Go &ge; 1.24
- golang.org/x/image &ge; v0.27.0
- OpenSlide &ge; 4.0.0

## Installation

OpenSlide Go requires [OpenSlide]. To run the tests, this requires that you
have downloaded the test image in `testdata`. To do so run

```shell
openslide/testdata
wget https://openslide.cs.cmu.edu/download/openslide-testdata/Generic-TIFF/CMU-1
```

Subsequently

```shell
bazelisk test //aifo/openslide_go/openslide:openslide_test --test_output=all --test_arg=-test.v --cache_test_results=no
```

## Command-line interface

## More Information

- [OpenSlide website][OpenSlide]
- [Sample data](https://openslide.cs.cmu.edu/download/openslide-testdata/)

## License

Licensed under the [Apache License 2.0](./LICENSE).
