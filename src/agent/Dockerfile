FROM ubuntu:22.04 AS base

RUN apt-get update && apt-get install -y \
    libssl-dev \
    libcurl4-openssl-dev \
    libxml2-dev \
    libzstd-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

FROM base AS builder

RUN apt-get update && apt-get install -y \
    cmake \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY src/agent/ .
RUN mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc)

FROM base AS final

COPY --from=builder /build/backup-agent /usr/local/bin/backup-agent

RUN useradd -m -s /bin/bash agent && \
    chown -R agent:agent /app

USER agent
ENTRYPOINT ["backup-agent"]
