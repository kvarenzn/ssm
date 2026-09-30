# ssm — AGENTS.md

## Build

Go 1.25 (`go.mod`); CI currently pins `go-version: ^1.24`. Module path: `github.com/kvarenzn/ssm`.
cgo is required.

```bash
go build -ldflags "-X main.SSM_VERSION=$VERSION"
go build ./...   # also compiles packages not reachable from main (e.g. o)
```

System deps: `libusb-1.0-dev` (via `github.com/google/gousb`) and the FFmpeg libs
`libavformat`/`libavcodec`/`libavutil` (used in `decoders/av`), plus gcc and pkgconf.
On Windows: MSYS2 + MinGW with mingw-w64-x86_64-{gcc,pkgconf,libusb,ffmpeg,ntldd}.
See `.github/workflows/release.yml` — the only CI/automation file.

No tests, no linter/formatter config in the repo.

## Entrypoints & architecture

- `main.go` — orchestrator, flag parsing, TUI (`tui` struct) and autoplay loop; calls `Extract()` in `extract.go` for `-e`.
- `controllers/` — touch backends: `hid` (libusb/AOA 2.0) and `scrcpy` (adb). `Controller` interface is in `hid.go`; both expose `Preprocess()` to map chart coords to device coords.
- `scores/` — chart parsers (`bms.go` for BanG, `sus.go` for PJSK), touch-event generation (`generate.go`), pointer allocation via DSATUR graph coloring (`dsatur.go`/`graph.go`) plus min-cost max-flow (`flow_graph.go`, SPFA + Dijkstra); also `colorize.go`, `sls.go`, `common.go`.
- `stage/` — judge line position calculators, one per game mode (`bang.go`, `pjsk.go`).
- `db/` — music title/jacket metadata: Bestdori (`bestdori.go`, BanG) or Sekai World (`sekai-world.go`, PJSK). Downloads `all.5.json`/`all.1.json`/`musics.json` into cwd on first use.
- `adb/` — ADB client that speaks the adb socket protocol directly (device/forward/sync); still launches the local `adb` binary to `start-server`.
- `uni/` + `decoders/` — Unity asset bundle extraction (TextAsset, Texture2D); `decoders/{astc,etc}` handle texture formats, `decoders/av` wraps FFmpeg.
- `term/` — cross-platform terminal UI (Unix/Windows), image display via half-block/Sixel/Kitty protocols.
- `common/` — shared touch-event types (`VirtualTouchEvent`, `ViscousEventItem`, …).
- `config/` — `./config.json` load/save and per-serial width/height prompt.
- `log/` — locale-aware logging; debug output gated by `-g`.
- `locale/` — i18n. Detects system locale (default `ja_JP` when detection fails); the message catalog only has English + Simplified Chinese, unmatched languages resolve to English.
- `utils/`, `optional/` — generic helpers (`Set`/`Queue`/`PriorityQueue`, `Optional[T]`).
- `k/` — Project Sekai `.bundle` file header unscrambling, used by `extract.go`.
- `o/` — OurNotes asset bundle + master-data decryption (AES-ECB/custom Rijndael, catalog name mapping). **Not imported by any package yet**; only compiled via `go build ./...`.
- `docs/` — `BUILD.md`, `FAQ.md`, `USAGE.md`. `dump_svg.py` — debug helper rendering notes/graphs to SVG.

## CLI flags (non-obvious)

| Flag | Detail |
|------|--------|
| `-b hid\|adb` | Backend. Default `hid`. `adb` requires `scrcpy-server-v3.3.1` in cwd (auto-downloaded, SHA256 verified). |
| `-k` | **Project Sekai (PJSK) mode.** Switches the BanG defaults to PJSK: chart parser BMS→SUS, DB Bestdori→Sekai World, judge line BanG→PJSK. |
| `-e <path>` | Extract assets from a Unity bundle dir. Detects BanG vs PJSK by dir structure (only paths containing `startapp` + a score/jacket/ingameskin marker are kept). Writes decoded files under `./assets/...` and an `./extract.json` manifest. |
| `-p <path>` | Custom chart file path. Skips song ID/difficulty lookup. |
| `-r left\|right` | Device orientation. `right` only changes behavior in the `hid` backend; the scrcpy/adb path accepts but ignores it. |
| `-s <serial>` | Target device serial. Omit to auto-select first. |
| `-n <id>` / `-d <diff>` | Song ID / difficulty. Both required unless `-p` is given. |
| `-g` | Debug logging. |
| `-v` | Print version and exit (database is loaded before this check). |

## Key gotchas

- **No `main` package tests exist.** Do not look for `_test.go` files.
- **Chart asset paths differ by game mode:** `assets/star/forassetbundle/startapp/musicscore/musicscore*/NNN/*_<diff>.txt` (BanG) vs `assets/sekai/assetbundle/resources/startapp/music/music_score/NNNN_01/<diff>.txt` (PJSK).
- **Config is `./config.json`** (gitignored, mode 0600). `config.Load` creates an empty file when missing and then fails to `json.Unmarshal` it, so a truly fresh run dies with `unexpected end of JSON input` — write `{}` (or a valid config) first. The interactive width/height prompt only runs from `Config.Get(serial)`.
- **`scrcpy-server` is NOT committed.** `.gitignore` has `/scrcpy-server-*`. It is auto-downloaded at runtime to `./scrcpy-server-v3.3.1`; the expected SHA256 is `SERVER_FILE_SHA256` in `main.go` and both the filename and version string are hardcoded in `adbBackend` (`controller.Open("./scrcpy-server-v3.3.1", "3.3.1")`).
- **Locale** auto-detected per OS. English messages may be incomplete; source strings are in `locale/messages.go`.
- **Gitignore** covers: `*.json`, `/*.svg`, `/assets/`, `/data/`, `/ssm`, `/ssm.exe`, `/*.txt`, `/scrcpy-server-*`, `/*.dll`.
- **License**: GPL-3.0-or-later. Every source file has a copyright+SPDX header.
- **CI** triggers on `v*` tags only. Cross-platform builds (linux amd64/arm64, windows amd64, macos amd64/arm64).

## 交互要求
** 请以中文思考和回复 **
