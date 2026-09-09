# mcp-server-sparx-ea — a stdio MCP server. Run it with `docker run -i`, or point
# an MCP client's "command" at a wrapper that does so. It links mdbtools through
# cgo, so the runtime image ships the mdbtools + glib shared libraries.

FROM golang:1.26 AS build
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential pkg-config libglib2.0-dev mdbtools-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=1 go build -trimpath -ldflags "-s -w" -o /out/mcp-server-sparx-ea .

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
        mdbtools libglib2.0-0 ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/mcp-server-sparx-ea /usr/local/bin/mcp-server-sparx-ea

# The server reads the model files you pass as tool arguments — mount them and
# pass paths under the mount point.
#   docker run -i --rm -v "$PWD:/work" ghcr.io/insonusk/mcp-server-sparx-ea \
#     # then the MCP client speaks JSON-RPC on stdio
VOLUME ["/work"]
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/mcp-server-sparx-ea"]
