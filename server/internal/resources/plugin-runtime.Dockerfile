FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    openjdk-17-jre-headless \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /plugins /data

WORKDIR /plugins

CMD ["sleep", "infinity"]
