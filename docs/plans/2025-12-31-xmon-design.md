# xmon Design Document

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:writing-plans to create implementation plans from this design.

**Goal:** Monitor X/Twitter accounts you care about, digest their tweets, and surface early signals about what influential people are focusing on.

**Use Cases:**
- Stay current with thought leaders (10-15 key accounts)
- Early detection of emerging topics and trends
- Understand who influential people are amplifying

---

## Architecture

Pipeline pattern (same as ghmon):

```
sync → fetch → analyze → digest → export
```

| Stage | Description |
|-------|-------------|
| **sync** | Import accounts from a list or add manually |
| **fetch** | Pull recent tweets via X API (respecting rate limits) |
| **analyze** | Aggregate stats, optionally generate LLM summaries |
| **digest** | Display activity summary in terminal |
| **export** | Generate markdown reports |

**Data Storage:** SQLite at `~/.xmon/xmon.db`

**Config:** `~/.xmon/config.yaml`

```yaml
x:
  bearer_token: "AAAA..."

apis:
  llm_provider: "ollama"
  llm_model: "llama3.2"

fetch:
  default_interval: 1440  # minutes (daily)

digest:
  default_days: 7
```

---

## Database Schema

```sql
-- X accounts being monitored
CREATE TABLE accounts (
    id INTEGER PRIMARY KEY,
    user_id TEXT UNIQUE NOT NULL,    -- X's numeric user ID
    username TEXT NOT NULL,           -- @handle
    name TEXT,
    bio TEXT,
    followers INTEGER,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_fetched DATETIME
);

-- Tweets (originals, retweets, quotes)
CREATE TABLE tweets (
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL,
    tweet_id TEXT UNIQUE NOT NULL,
    tweet_type TEXT NOT NULL,         -- 'original', 'retweet', 'quote'
    content TEXT,
    referenced_user TEXT,             -- who they RT'd or quoted
    referenced_tweet_id TEXT,
    likes INTEGER,
    retweets INTEGER,
    created_at DATETIME,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

-- API usage tracking
CREATE TABLE api_usage (
    id INTEGER PRIMARY KEY,
    month TEXT NOT NULL,              -- 'YYYY-MM'
    tweets_read INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(month)
);
```

---

## Commands

| Command | Description |
|---------|-------------|
| `xmon init` | Initialize config and database, prompt for bearer token |
| `xmon add <username>` | Add an X account to monitor |
| `xmon remove <username>` | Remove an account |
| `xmon accounts` | List monitored accounts |
| `xmon fetch` | Pull recent tweets for all accounts |
| `xmon digest` | Show activity summary (--days, --smart) |
| `xmon show <username>` | Show detailed activity for one account |
| `xmon export` | Generate markdown report |

**Key Flags:**
- `--days N` - Override time window (default: 7)
- `--smart` - Enable LLM analysis for richer summaries

**Example Workflow:**
```bash
xmon init                     # Setup with bearer token
xmon add pmarca               # Add accounts to monitor
xmon add naval
xmon add paulg
xmon fetch                    # Pull their tweets
xmon digest                   # See what's happening
xmon digest --smart           # Get LLM insights
xmon export > weekly.md       # Save report
```

---

## X API Integration

**Authentication:** Bearer Token (OAuth 2.0 App-Only) stored in config.

**Free Tier Limits:**
- 1,500 tweets/month read
- 1 app environment
- No access to full-archive search

**API Endpoints:**

| Data | Endpoint | Notes |
|------|----------|-------|
| User lookup | `GET /2/users/by/username/:username` | Get user ID from handle |
| User tweets | `GET /2/users/:id/tweets` | Timeline (includes RTs, quotes) |

**Request Strategy:**
- Fetch once daily by default (configurable)
- Store `since_id` per account to only fetch new tweets
- Cache everything in SQLite to minimize repeat calls
- Track rate limit headers, pause if near limit

**Tweet Type Detection:**
```
- referenced_tweets[].type == "retweeted" → Retweet
- referenced_tweets[].type == "quoted" → Quote tweet
- No referenced_tweets → Original tweet
```

**Rate Limit Handling:**
- Parse `x-rate-limit-remaining` and `x-rate-limit-reset` headers
- Warn user when approaching monthly limit
- Store monthly usage count in database

---

## Digest Output

**Standard digest (`xmon digest`):**

```
X DIGEST (Dec 24 - Dec 31, 2024)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Summary: 12 accounts · 87 tweets · 34 retweets · 15 quotes

🔥 Most Active
  @pmarca          23 tweets
  @naval           18 tweets
  @paulg           12 tweets

🔁 Most Amplified (who they're retweeting)
  @elonmusk        ★ by pmarca, naval, balajis
  @sama            ★ by paulg, naval
  @vikirai         ★ by pmarca, balajis

📢 Trending Topics (hashtags & keywords)
  AI agents · crypto · startups · regulation

💬 Notable Tweets (most engaged)
  @naval: "The best founders are missionaries, not mercenaries"
    ↳ 12.4K likes · 2.1K RTs
```

**Smart digest (`xmon digest --smart`):**

Adds LLM-generated insights:

```
💡 Key Themes (AI-generated)
  • AI agent discourse intensifying: 4 accounts discussing autonomous
    agents this week, up from 1 last week
  • Regulatory concern: pmarca and balajis both critical of proposed
    AI legislation, may signal VC sentiment shift
  • Crypto resurgence: increased mentions of BTC/ETH across 3 accounts
```

---

## Implementation Phases

### Phase 1 - MVP
- Project setup (Go, Cobra, SQLite)
- `init` command (config, database, bearer token setup)
- `add` / `remove` commands
- `accounts` command
- X API client (user lookup, timeline fetch)
- `fetch` command (pull tweets, respect rate limits)
- `digest` command (basic aggregation)
- `show <username>` command

### Phase 2 - Intelligence
- LLM integration (Ollama) for `--smart` summaries
- "Most Amplified" detection (who multiple accounts RT)
- Topic/keyword extraction
- Notable tweets ranking
- Monthly API usage tracking with warnings

### Phase 3 - Export & Polish
- `export` command (markdown generation)
- Daemon mode (scheduled fetching)
- Scraping fallback (if API limits prove too restrictive)

---

## Tech Stack

- **Language:** Go
- **CLI Framework:** Cobra
- **Database:** SQLite
- **API:** X API v2
- **LLM:** Ollama (optional)
- **Styling:** Lipgloss (terminal output)

Same stack as ghmon for consistency.
