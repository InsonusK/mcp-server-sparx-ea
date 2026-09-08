Task Specification for Agent
Context & Goal
We are developing a high-performance Model Context Protocol (MCP) server in Go: mcp-server-sparx-ea.
The primary objective of this server is to allow LLM agents to query Sparx Enterprise Architect project files (.eapx / MS Access JET 4.0 databases) directly under Linux via standard stdio.

Chosen Architecture Approach
We rejected external CLI sub-process execution (mdbtools binary wrapper via /tmp) and pure Go parsers (which lack SQL support and break CP1251 encodings).

Selected Approach: Go + CGO Direct Binding with libmdb (mdbtools core library)

Single Process Execution: The server runs as a single compiled Go binary.

In-Memory Operations: CGO bindings directly hook into libmdb functions in C. No /tmp files or child sub-processes are created at runtime.

Database & Encoding Compatibility: Native support for JET 4.0 structures, complex tables (t_object, t_connector, etc.), and proper text decoding (CP1251 / Windows-1251 to UTF-8).

Technical Requirements & Setup
Build Prerequisites:

Linux environment with gcc, CGO_ENABLED=1.

C dependencies installed: libmdb-dev, libglib2.0-dev.

Core Package Structure:

Plaintext
mcp-server-sparx-ea/
├── main.go               # MCP Server entry point using mark3labs/mcp-go
└── eapx/
    ├── cgo_mdb.go        # CGO headers, bindings, and C memory management
    └── reader.go         # Go wrappers, data structs, and CP1251 conversion
Key Integration Points:

CGO Binding (eapx/cgo_mdb.go): Utilize #cgo LDFLAGS: -lmdb -lglib-2.0 to import <mdbtools.h> and <glib.h>.

Resource Safety: Ensure all memory allocated in C space (C.CString, g_string_new, mdb_col_to_string) is explicitly freed via defer C.free(...) or g_string_free to prevent memory leaks during long-running MCP sessions.

Encoding Handling: Safely convert strings fetched from .eapx text/memo columns into valid UTF-8 Go strings.

MCP Tools to Implement:

ea_search_objects: Query elements from t_object table by name or type.

ea_get_connectors: Fetch relationships between objects from t_connector.

Definition of Done (DoD)
Project builds into a single standalone binary using CGO_ENABLED=1 go build.

Execution of MCP tools does not spawn child processes (exec.Command) or write temporary runtime files to disk.

Reading .eapx correctly yields JSON-formatted output for requested objects and relationships without breaking Cyrillic (CP1251) characters.