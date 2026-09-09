# Installation and access — sparx-ea-mcp

## Get the server

The server is a single pure-Go binary, no dependencies.

- **Release binary** — the repository's Releases page has archives for
  `linux` / `darwin` / `windows` × `amd64` / `arm64`.
- **From source** — Go 1.23+:

  ```bash
  go build -o mcp-server-sparx-ea .
  ```

Verify: `go test ./...`.

## Register with an MCP client

The server speaks MCP over stdio, no flags, no environment variables.

```json
{
  "mcpServers": {
    "sparx-ea": { "command": "/absolute/path/to/mcp-server-sparx-ea" }
  }
}
```

After registering, `tools/list` returns 21 tools:

```
ea_model_tree  ea_element  ea_package  ea_diagram  ea_archimate_types
ea_new_model  ea_create_root_package  ea_set_root_name
ea_create_element  ea_update_element  ea_delete_element
ea_create_relationship  ea_delete_relationship
ea_create_package  ea_update_package  ea_copy_package  ea_delete_package
ea_create_diagram  ea_place_on_diagram  ea_move_on_diagram  ea_remove_from_diagram
```

## Files the server reads

All tools take a `file` — a model the user exported from EA
(`File → Export → Package to XMI`, XMI 2.1). Editing tools also write to
`output` (a different, writable path).

The server never opens a path you did not pass in a tool argument, and never
makes a network call.
