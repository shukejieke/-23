# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

率土之滨 (Three Kingdoms) game traffic analyzer. Passively captures TCP packets from the game server (port 8001), decodes proprietary protocol frames (zlib/XOR), and stores alliance data (members, battle reports, leaderboards) into MySQL. Serves a Vue 3 frontend on port 9527.

## Build & Run

**Backend requires CGO (pcap bindings):**
```bash
# macOS arm64
CGO_ENABLED=1 go build -tags="nomsgpack" -ldflags="-s -w" -o dist/stzbHelper-darwin-arm64 .

# Windows amd64
go build -tags="nomsgpack" -ldflags="-s -w" -o dist/stzbHelper-windows-amd64.exe .

# Dev run
CGO_ENABLED=1 go run .
```

**Frontend:**
```bash
cd web && npm install && npm run build   # produces web/dist/
cd web && npm run dev                    # Vite dev server
```

**Environment:**
- `STZB_MYSQL_DSN` — MySQL DSN (default: `root:123456@tcp(127.0.0.1:3306)/stzb?charset=utf8mb4&parseTime=True&loc=Local`)
- No tests exist yet.

## Data Flow

```
TCP :8001 packets
  └─ internal/capture/capture.go     — BPF filter, TCP reassembly, frame extraction
       └─ internal/parser/parse.go   — protocol dispatch, JSON decode, GORM save
            └─ model/                — GORM models + AutoMigrate
                 └─ internal/repo/   — parameterized queries (new code uses this)
                      └─ internal/service/ — business logic
                           └─ http/handle/api/ — thin JSON handlers
                                └─ :9527 (Gin)
```

## Protocol Encoding

Two wire encodings, identified at `frame[12]`:
- **type 3** — zlib compressed starting at `frame[17:]`, magic bytes `0x78 0x9C`
- **type 5** — XOR with key `152`, starting at `frame[12:]` (first byte is the type indicator)

Frame structure: `[4-byte big-endian length][4-byte cmd_id][4 bytes misc][1-byte dataType][4 bytes misc][payload]`
Total frame size = `length + 4`.

Key cmd IDs:
| ID | Meaning |
|----|---------|
| 3686 | 主公簿 (activation packet — triggers DB init + IP binding) |
| 103 | Alliance members (direct) |
| 100 | Alliance overview (member data embedded recursively) |
| 92 | Battle reports |
| 142 | Group metadata |
| 700 | Alliance leaderboard |
| 514 | Personal leaderboard snapshots |
| 949 | Member status delta (captured but not stored to DB) |

## Critical Architecture Notes

**Startup sequence matters:** `model.InitDB()` is called eagerly at startup (before any packet arrives). DB name is overridden when cmd 3686 is seen, re-initializing to `roleName_serverName`. Until then, all data goes to the default `stzb` database.

**IP binding:** After cmd 3686, `runtime.BindIPs()` filters subsequent packets to only the game server ↔ client pair. Controlled by `global.ExVar.BindIpInfo`.

**Dual battle report modes:** cmd 92 always triggers simple `report` table. Detailed `battle_report` table only populated when `global.ExVar.NeedGetBattleData = true` (toggled via HTTP API).

**Packet loss recovery (type 3 only):** For cmd 103 and 92, if `frame length ≠ bufsize`, the frame is held in `flowState.lossBytes` waiting for a subsequent high-cmd-ID frame to complete it. TTL: 5 seconds.

**Repo vs model.Conn:** Mid-refactor. Old handlers call `model.Conn` directly. New code flows through `repo.SetDB()` / `repo.DB()` (which falls back to `model.Conn`). New code should use the repo layer.

**TCP reassembly:** `flowState.fullbuf` accumulates PSH=false segments; PSH=true triggers processing. After the frame-extraction loop, leftover bytes that look like a partial frame header (`bufsize > 0`, `totalLen > len(work)`) should be carried forward — this is where `invalid-frame` log entries come from when a large frame spans multiple PSH boundaries.

## Runtime Diagnostics

The `/stzb/runtime/*` endpoints are key for debugging capture issues:
- `/stzb/runtime/status` — capture stats, parse counts, IP binding state
- `/stzb/runtime/cmd/list` — all seen cmd IDs with counts
- `/stzb/runtime/cmd/analysis` — auto-analyzed field schemas
- `/stzb/runtime/raw-capture/list` — raw hex dumps for cmd 100/103/92
- `/stzb/runtime/member/diagnosis` — member sync troubleshooting
- `/stzb/runtime/rank/trace` — leaderboard trigger timeline

## Extra Data

`skill_extra.json`, `hero_extra.json`, `gear_extra.json` are imported into MySQL asynchronously at startup via `model.ImportExtraJSONToMySQL()`. Failures are non-fatal. These power the battle report enrichment (hero names, gear names, skill names).
