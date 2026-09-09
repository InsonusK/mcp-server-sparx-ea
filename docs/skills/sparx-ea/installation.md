# Installation and access — sparx-ea-mcp

## Build the server

Needs Go 1.23+ and the `mdbtools` development headers (the server links `libmdb`
through cgo).

```bash
# Debian / Ubuntu
sudo apt-get install -y build-essential pkg-config libglib2.0-dev mdbtools-dev
# macOS
brew install mdbtools pkg-config glib

CGO_ENABLED=1 go build -o mcp-server-sparx-ea .
```

Verify: `CGO_ENABLED=1 go test ./...` — all 176 Cucumber scenarios pass.

## Register with an MCP client

The server speaks MCP over stdio, no flags, no environment variables.

```json
{
  "mcpServers": {
    "sparx-ea": { "command": "/absolute/path/to/mcp-server-sparx-ea" }
  }
}
```

After registering, `tools/list` returns 18 tools:

```
ea_query
ea_model_tree  ea_element  ea_package  ea_diagram  ea_archimate_types
ea_create_element  ea_update_element  ea_delete_element
ea_create_relationship  ea_delete_relationship
ea_create_package  ea_update_package  ea_copy_package  ea_delete_package
ea_place_on_diagram  ea_move_on_diagram  ea_remove_from_diagram
```

## Files the server reads

- `ea_query` — a Sparx EA project file: `.eapx` (JET4) or `.eap` (JET 3.5),
  passed as the `file` argument.
- The ArchiMate tools — a model the user exported from EA
  (`File → Export → Package to XMI`, XMI 2.1), passed as `file`. Editing tools
  also write to `output` (a different, writable path).

The server never opens a path you did not pass in a tool argument, and never
makes a network call.
