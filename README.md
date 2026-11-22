# zoro

A `mechfight/1` bot for the Builders Buildathon **Mech Royale** track. It pilots one mech in the
16-mech battle royale: it reads observations as JSON lines on stdin and answers each one with a single
action line on stdout.

Go, standard library only.

## Attribution

zoro is a from-scratch rewrite based on the Mech Royale bot in
[andrew-zhouu/builders-mech-royale-2026](https://github.com/andrew-zhouu/builders-mech-royale-2026)
at commit [`f3d4e15`](https://github.com/andrew-zhouu/builders-mech-royale-2026/tree/f3d4e15).
That project is MIT licensed. The overall approach comes from it: the walk grid and roof detection,
cover-first fighting, a gun lead learned per opponent, and the range-aware fight odds. The code
here was rewritten and restructured, and the measured hit-rate tables and several tuning constants
come from the original. Thanks to its authors.

Differences from the original:

- New package layout and APIs (`wire`, `arena`, `grid`, `pilot`).
- One fixed tuning set. The original's parameter variants, seat-plan experiments, lobby-type
  detection, visibility-map hideouts and telemetry logging are gone.
- A single careful "brawl" stance, plus an all-in duel stance once two mechs are left.
- A synthetic-arena smoke test (`go test ./...`).

## Layout

| Path | What it is |
| --- | --- |
| `cmd/zoro` | Protocol loop: answers only the newest observation, never writes anything else to stdout, has a fallback brain if the map cannot be loaded |
| `internal/wire` | `mechfight/1` message types and the command encoder |
| `internal/arena` | Map loading (`/arena/v2`), spatial buckets, terrain, ray casts, line of sight, tree cover |
| `internal/grid` | 2 m walk grid, A*/Dijkstra, connected regions, rooftop regions, top-surface routes, openness map |
| `internal/pilot` | The agent: memory, ballistics and lead, fight odds, dodging, cover, roofs, loot, zone, decision |

## Build

```sh
docker build --platform linux/amd64 -t zoro .
```

To build without Docker (Go 1.25 or newer):

```sh
go build -o zoro ./cmd/zoro
go test ./...
```

To run the kit's conformance check against the image, from the `mechfight-kit` folder:

```sh
python3 check/check.py --map-dir /arena -- docker run --rm -i --network none --cpus 1 --memory 512m \
  --read-only --tmpfs /tmp:rw,size=64m --user 65534:65534 -v "$PWD/arena:/arena:ro" zoro
```

## Runtime

- All map analysis runs before answering `hello`. Decisions take a few milliseconds at most.
- No network access, no files written.
- A panic inside a decision is caught and answered by repeating the previous movement, so the mech
  never goes silent.

## License

MIT. See `LICENSE`, which also carries the original project's notice.
