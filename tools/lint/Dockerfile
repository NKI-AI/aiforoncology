# Stage 1: Extract clang-tidy
FROM debian:bullseye-slim AS clang_tidy_extractor

# This requires bazelisk build //tools/lint:clang_tidy_tar

COPY tools/lint/clang_tidy_tar.tar.gz /tmp/
WORKDIR /tmp
RUN apt-get update && apt-get install -y --no-install-recommends \
      tar \
      && tar -xzf clang_tidy_tar.tar.gz \
      && rm clang_tidy_tar.tar.gz \
      && rm -rf /var/lib/apt/lists/*

# Stage 2: Download bazel only
FROM alpine:3.21 AS bazel_downloader

ARG BAZEL_VERSION=7.6.1

RUN apk add --no-cache curl

RUN curl -fsSL -o /usr/local/bin/bazel https://github.com/bazelbuild/bazel/releases/download/${BAZEL_VERSION}/bazel-${BAZEL_VERSION}-linux-x86_64 && \
      chmod +x /usr/local/bin/bazel

# Stage 3: cpplint+ruff container (Alpine, minimal)
FROM python:3.12.10-alpine3.21 AS cpplint_ruff

RUN apk add --no-cache \
      git \
      libstdc++

RUN pip install --no-cache-dir \
      ruff==0.11.6 \
      cpplint==2.0.2

WORKDIR /workspace

# Stage 4: clang-tidy container (Debian slim for real glibc)
FROM python:3.12-slim AS clang_tidy_bazel

# zstd is needed for clang-tidy
RUN apt-get update && apt-get install -y --no-install-recommends \
      libc++-dev \
      libc++abi-dev \
      libstdc++-12-dev \
      zstd \
      curl \
      git \
      && rm -rf /var/lib/apt/lists/*

COPY --from=clang_tidy_extractor /tmp/clang-tidy /usr/local/bin/clang-tidy
COPY compile_commands.json /workspace/compile_commands.json

WORKDIR /workspace
