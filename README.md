# Datastar Streaming + Fat Morph

A tiny demo that broadcasts HTML patches over SSE on each action.

## Run

```bash
air
```

Or:

```bash
go run ./src
```

Then open `http://localhost:8080` in your browser.

## How it works (step by step)

1. The server starts and initializes the SSE hub.
2. When a browser hits `/`, the server sets a client ID cookie and returns the full HTML page.
3. The page opens an SSE connection to `/stream`.
4. `/stream` registers the client with the hub and immediately sends an initial HTML patch for the `main` block.
5. When the user clicks a button (`/increment`, `/decrement`, `/reset`):
   - The server applies the action to in-memory state.
   - It appends the action to the in-memory history list.
   - It renders one HTML patch and broadcasts it to all connected SSE clients.
6. Every connected client receives the patch and Datastar morphs the DOM.
7. On restart, the app starts from a clean in-memory state.

## Files to know

- `src/main.go`: app startup and HTTP routes.
- `src/handlers.go`: HTTP handlers and HTML patch rendering/broadcast.
- `src/hub.go`: SSE client registry and broadcast fan-out.
- `src/state.go`: in-memory counter state and reducer.
