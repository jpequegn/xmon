# xmon

Monitor X/Twitter accounts and get digests of their activity.

## Features

- Track tweets from accounts you care about
- Detect who influential people are amplifying
- Generate activity digests with trending topics
- Optional LLM-powered analysis of themes
- Export reports to markdown

## Installation

```bash
go install github.com/jpequegn/xmon@latest
```

## Quick Start

```bash
# Initialize with your X API bearer token
xmon init

# Add accounts to monitor
xmon add pmarca
xmon add naval

# Fetch recent tweets
xmon fetch

# View digest
xmon digest

# Get AI-powered insights
xmon digest --smart
```

## Commands

| Command | Description |
|---------|-------------|
| `xmon init` | Initialize config and database |
| `xmon add <user>` | Add an account to monitor |
| `xmon remove <user>` | Remove an account |
| `xmon accounts` | List monitored accounts |
| `xmon fetch` | Pull recent tweets |
| `xmon digest` | Show activity summary (--smart for AI insights) |
| `xmon show <user>` | Show user details |
| `xmon export` | Generate markdown report |

## Configuration

Config is stored in `~/.xmon/config.yaml`

```yaml
x:
  bearer_token: "AAAA..."

apis:
  llm_provider: "ollama"
  llm_model: "llama3.2"

digest:
  default_days: 7
```

## Development Status

### Phase 1 (MVP) - Complete
- [x] Project setup (Go, Cobra, SQLite)
- [x] init command (config, database, bearer token setup)
- [x] add / remove commands
- [x] accounts command
- [x] X API client (user lookup, timeline fetch)
- [x] fetch command (pull tweets, respect rate limits)
- [x] digest command (basic aggregation)
- [x] show command

### Phase 2 (Intelligence) - Complete
- [x] LLM integration (Ollama) for --smart summaries
- [x] "Most Amplified" detection (who multiple accounts RT)
- [x] Topic/keyword extraction
- [x] Notable tweets ranking
- [x] Monthly API usage tracking with warnings

### Phase 3 (Export & Polish) - Planned
- [ ] export command (markdown generation)
- [ ] Daemon mode (scheduled fetching)
- [ ] Scraping fallback (if API limits prove too restrictive)
