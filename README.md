# AURor

An AI-powered CLI tool that performs deep security analysis of Arch Linux AUR `PKGBUILD` files. It fetches, parses, and evaluates PKGBUILD sources using a structured multi-step assessment methodology guided by an LLM, then generates a detailed markdown security report with a risk verdict (LOW / MEDIUM / HIGH / CRITICAL).

Currently in development. It may work, it may not work, it will surely produce reports that **MUST BE VERIFIED**.

## Features

- **Automated PKGBUILD security assessment** -- checksum verification, Shannon entropy analysis, obfuscation detection, network interaction auditing, privilege escalation checks, and supply chain checks.
- **Dual input support** -- analyze PKGBUILD files from local paths or remote URLs (AUR archive links).
- **AI-driven analysis** -- uses OpenAI-compatible LLMs with an embedded 9-step security assessment methodology to evaluate build scripts and generate structured reports.
- **Shell execution tools** -- the AI model can autonomously execute shell commands (e.g., running entropy-scanning scripts) during assessment.
- **Markdown report generation** -- detailed, human-readable security reports saved to the `reports/` directory.

## Tech Stack

| Component       | Technology                                      | Purpose                            |
|-----------------|-------------------------------------------------|------------------------------------|
| Language        | Go                                              | Core implementation                |
| CLI Framework   | [Cobra](https://github.com/spf13/cobra)         | Command-line interface             |
| Configuration   | [Viper](https://github.com/spf13/viper)         | Config file & environment loading  |
| AI Framework    | [Firebase Genkit](https://firebase.google.com/docs/genkit) (Go SDK) | Flow/agent orchestration      |
| LLM Provider    | OpenAI-compatible adapter                       | Flexible backend (Ollama, vLLM, OpenAI, etc.) |
| Logging         | [Logrus](https://github.com/sirupsen/logrus)    | Structured logging                 |

### Supported LLM Backends

AURor connects via the OpenAI API protocol, meaning any compatible endpoint works out of the box:

- **Llama.cpp** (default, local)
- **Ollama**
- **OpenAI**
- **vLLM**
- Any other OpenAI-compatible API

## Installation

### From source

```bash
git clone https://github.com/mcaimi/auror
cd auror
make
```

## Usage

### Analyze a remote AUR package

```bash
./auror -p https://aur.archlinux.org/cgit/aur.git/plain/PKGBUILD?h=some-package
```

### Analyze a local PKGBUILD file

```bash
./auror -p ./some-pkg/PKGBUILD
```

### Flags

| Flag            | Short | Required | Description                        |
|-----------------|-------|----------|------------------------------------|
| `--config`      | `-c`  | No       | Path to configuration file         |
| `--pkgbuild`    | `-p`  | Yes      | PKGBUILD URL or local file path    |

## Configuration

Config is loaded in this order (first found wins):

1. CLI flag `--config`
2. `auror.yaml` in the current directory
3. `$HOME/.auror/auror.yaml`
4. `/etc/auror/auror.yaml`

Environment variables with the `AUROR_` prefix override all config values. For example:

```bash
AUROR_BACKEND_MODEL=gpt-4 AUROR_BACKEND_APIKEY=sk-xxx ./auror -p <url>
```

### Default configuration example

```yaml
backend:
  provider: llama.cpp
  baseurl: http://localhost:11434/v1
  apikey: ""
  model: llama.cpp/unsloth/qwen3.5-9B-gguf:Q4_K_M
  temperature: 0.7
  maxtokens: 32768
  usagetracking: true

skills:
  skillspath: skill
  skillfile: pkgbuild-security-assessment.md
  systemprompt: ""

output:
  reports: reports
```

## Assessment Methodology

AURor applies a structured 9-step security assessment to every PKGBUILD:

1. **Metadata Analysis** -- Package name, version, maintainer, and source information.
2. **Checksum Integrity** -- Verification of md5/sha256 checksums for fetched sources.
3. **Shannon Entropy Scanning** -- Detects high-entropy (potentially base64-encoded or obfuscated) URLs and data.
4. **Obfuscation Detection** -- Identifies obfuscated shell code, inline evals, and encoded payloads.
5. **Network Interaction Audit** -- Analyzes all network calls in build/install functions.
6. **Privilege Escalation Checks** -- Flags dangerous `sudo`, `chmod 777`, or `chown root` usage.
7. **Supply Chain Assessment** -- Evaluates source repository provenance and third-party dependency risks.
8. **VCS-specific Checks** -- Additional checks for packages cloned from version control repositories.
9. **Risk Verdict** -- A final severity rating: LOW, MEDIUM, HIGH, or CRITICAL.

## Output

After analysis, a markdown report is saved to the `reports/` directory with the naming convention:

```
reports/report-<pkgname>-<date>.md
```

## Project Structure

```
auror/
  auror.yaml                       # Default configuration
  Containerfile                    # Docker container builder
  Makefile                         # Build, lint, format targets
  go.mod / go.sum                  # Dependencies
  cmd/auror/main.go                # CLI entry point
  internal/
    config/                        # Configuration loading
    backend/                       # OpenAI-compatible adapter
    flows/                         # Genkit analysis flow
    tools/                         # AI-executable tools
    utils/                         # PKGBUILD fetch & utility functions
  skill/
    pkgbuild-security-assessment.md  # Security assessment methodology
```

## Local model for on-device analysis

I use a local instance of Qwen-9B-GGUF from [unsloth](https://huggingface.co/unsloth/Qwen3.5-9B-GGUF), with llama.cpp on my Mac M1 machine.

```bash
llama serve --host 0.0.0.0 --port 11434 -hf unsloth/qwen3.5-9B-gguf:Q4_K_M --cont-batching -c 65535 -t 8 --device MTL0 --log-colors on
```

The endpoint is OpenAI compatible, so it works out of the box.

## Build the Container Image

Auror can run in a container image, with the provided Containerfile.

1- Set up the configuration file (`auror.yaml`)

2- Build the image

```bash
podman build -t auror:dev .
```

3- Run the container

```bash
podman run --rm --name auror -it auror:dev -p <PKGBUILD PATH/URL>
```

Using a custom `auror.yaml` is the preferred way of building and running the image but settings can also be overridden via ENV variables (look into `internal/config/default.go`):

```bash
podman run --rm --name auror -it -e AUROR_BACKEND.MODEL="MODEL NAME" -e AUROR_BACKEND.BASEURL="http://baseurl/v1" auror:dev -p <PKGBUILD URL/PATH>
```

## LICENSE

This software is licensed under the GPL v3. (see LICENSE file)
