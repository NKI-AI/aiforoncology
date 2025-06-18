# Linters

The AI for Oncology monorepo uses linters to maintain code quality. While linters are not required before commits, they run automatically on GitHub where [ReviewDog](https://github.com/reviewdog/reviewdog) provides feedback based on the results.

## Overview

- **Local linters** run in the [Bazel](https://bazel.build/) environment
- **Remote runners** use lightweight Docker containers on GitHub runners

## Running Linters

### Locally

Use the [Aspect CLI](https://aspect.build/cli) to run all linters:

```shell
bazelisk lint //...
```

Remote containers can be built and pushed with:

```shell
make linter_containers
```
