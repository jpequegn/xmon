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
| `xmon digest` | Show activity summary |
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

### Phase 1 (MVP)
- [ ] Project setup
- [ ] init command
- [ ] add/remove commands
- [ ] accounts command
- [ ] X API client
- [ ] fetch command
- [ ] digest command
- [ ] show command

### Phase 2 (Intelligence)
- [ ] LLM integration
- [ ] Most Amplified detection
- [ ] Topic extraction
- [ ] Notable tweets ranking

### Phase 3 (Export & Polish)
- [ ] export command
- [ ] Daemon mode
- [ ] Scraping fallback
