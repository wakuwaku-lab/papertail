# Papertail

A CLI tool for downloading and processing logs from the Papertrail API.

## Installation

```bash
go install github.com/wakuwaku-lab/papertail@latest
````

Or build from source:

```bash
git clone https://github.com/wakuwaku-lab/papertail.git
cd papertail
go build -o papertail .
```

## Configuration

Two methods are supported for configuring the API token:

**Method 1: Environment Variable**

```bash
export PAPERTRAIL_API_TOKEN="your_token_here"
```

**Method 2: .env File**

```bash
# Create a .env file in the project directory
echo "PAPERTRAIL_API_TOKEN=your_token_here" > .env
```

The program will automatically load the `.env` file from the current directory.

## Usage

### Three Modes

**1. Download by Hour**

```bash
# Last 24 hours
papertail logs --hours 24 --tsv output.tsv

# Specific hour range (1–6 hours ago)
papertail logs --hours 1-6 --tsv output.tsv
```

**2. Download by Specific Date**

```bash
papertail logs --date 2024-03-01 --tsv output.tsv
```

**3. Download by Date Range**

```bash
papertail logs --start 2024-03-01 --end 2024-03-07 --tsv output.tsv
```

### Parameters

| Parameter | Short | Description                                           |
| --------- | ----- | ----------------------------------------------------- |
| `--hours` | `-n`  | Number of hours, supports formats like `24` or `1-12` |
| `--date`  | `-D`  | Specific date (YYYY-MM-DD)                            |
| `--start` | `-s`  | Start date of the date range                          |
| `--end`   | `-e`  | End date of the date range                            |
| `--tsv`   | `-t`  | Output TSV file path                                  |
| `--db`    | `-d`  | SQLite database path (optional)                       |

### Import into SQLite

```bash
papertail logs --hours 24 --db logs.db --tsv logs.tsv
```

---

**Author:** [exia.huang](https://github.com/exiahuang)  
**Organization:** [wakuwaku-lab](https://github.com/wakuwaku-lab)
