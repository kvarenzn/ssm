# ssm — AGENTS.md

## Build

```bash
go build -ldflags "-X main.SSM_VERSION=$VERSION"
```

System deps: `libusb-1.0-dev`, `libavformat-dev`, `libavcodec-dev`, `libavutil-dev`, gcc, pkgconf.
On Windows: MSYS2 + MinGW with mingw-w64-x86_64-{gcc,pkgconf,libusb,ffmpeg,ntldd}. See `.github/workflows/release.yml`.

No tests, no linter/formatter config in the repo.

## Entrypoints & architecture

- `main.go` — orchestrator, flag parsing, calls `Extract()` in `extract.go` for `-e`.
- `controllers/` — touch backends: `hid` (libusb/AOA 2.0) and `scrcpy` (adb).
- `scores/` — chart parsers (`bms.go` for BanG, `sus.go` for PJSK), touch event generation, graph coloring (DSATUR) for pointer allocation.
- `stage/` — judge line position calculators, differs per game mode.
- `db/` — music title/jacket metadata from Bestdori (BanG) or Sekai (PJSK).
- `adb/` — custom ADB client (not using `adb` binary).
- `uni/` + `decoders/` — Unity asset bundle extraction (TextAsset, Texture2D).
- `term/` — cross-platform terminal UI (Unix/Windows).
- `locale/` — i18n, auto-detects system language, falls back to `ja_JP`.

## CLI flags (non-obvious)

| Flag | Detail |
|------|--------|
| `-b hid\|adb` | Backend. Default `hid`. `adb` requires `scrcpy-server-v3.3.1` in cwd (auto-downloaded, SHA256 verified). |
| `-k` | **Project Sekai mode**. Changes chart format (SUS→BMS), DB (Sekai→Bestdori), judge lines. |
| `-e <path>` | Extract assets from Unity bundle dir. Detects BanG vs PJSK by dir structure. Writes `./extract.json`. |
| `-p <path>` | Custom chart file path. Skips song ID/difficulty lookup. |
| `-r left\|right` | Device orientation. Only matters for `hid` backend; ignored for `adb`. |
| `-s <serial>` | Target device serial. Omit to auto-select first. |
| `-g` | Debug logging. |
| `-v` | Print version and exit. |

## Key gotchas

- **No `main` package tests exist.** Do not look for `_test.go` files.
- **Chart asset paths differ by game mode:** `assets/star/forassetbundle/startapp/musicscore/musicscore*/NNN/*_<diff>.txt` (BanG) vs `assets/sekai/assetbundle/resources/startapp/music/music_score/NNNN_01/<diff>.txt` (PJSK).
- **Config is `./config.json`** — created on first run with interactive prompts for width/height per device serial.
- **`scrcpy-server`** is bundled in repo for CI, but auto-downloaded at runtime for adb backend. SHA256 must match `SERVER_FILE_SHA256` in `main.go`.
- **Locale** auto-detected per OS. English messages may be incomplete; source strings are in `locale/messages.go`.
- **Gitignore** covers: `*.json`, `/assets/`, `/data/`, `/ssm`, `/ssm.exe`, `/*.txt`, `/scrcpy-server-*`.
- **License**: GPL-3.0-or-later. Every source file has a copyright+SPDX header.
- **CI** triggers on `v*` tags only. Cross-platform builds (linux amd64/arm64, windows amd64, macos amd64/arm64).

## 交互要求
** 请以中文思考和回复 **
