# Papertail

Papertrail API からログをダウンロードして処理するための CLI ツールです。

## インストール

```bash
go install github.com/wakuwaku-lab/papertail@latest
````

またはソースコードからビルド:

```bash
git clone https://github.com/wakuwaku-lab/papertail.git
cd papertail
go build -o papertail .
```

## 設定

API トークンは次の 2 つの方法で設定できます。

**方法1: 環境変数**

```bash
export PAPERTRAIL_API_TOKEN="your_token_here"
```

**方法2: .env ファイル**

```bash
# プロジェクトディレクトリに .env ファイルを作成
echo "PAPERTRAIL_API_TOKEN=your_token_here" > .env
```

プログラムはカレントディレクトリにある `.env` ファイルを自動的に読み込みます。

## 使用方法

### 3つのモード

**1. 時間単位でダウンロード**

```bash
# 直近 24 時間
papertail logs --hours 24 --tsv output.tsv

# 指定した時間範囲（1〜6時間前）
papertail logs --hours 1-6 --tsv output.tsv
```

**2. 特定の日付でダウンロード**

```bash
papertail logs --date 2024-03-01 --tsv output.tsv
```

**3. 日付範囲でダウンロード**

```bash
papertail logs --start 2024-03-01 --end 2024-03-07 --tsv output.tsv
```

### パラメータ

| パラメータ     | 短縮形  | 説明                           |
| --------- | ---- | ---------------------------- |
| `--hours` | `-n` | 時間数。`24` または `1-12` の形式をサポート |
| `--date`  | `-D` | 特定の日付 (YYYY-MM-DD)           |
| `--start` | `-s` | 日付範囲の開始日                     |
| `--end`   | `-e` | 日付範囲の終了日                     |
| `--tsv`   | `-t` | 出力する TSV ファイルのパス             |
| `--db`    | `-d` | SQLite データベースのパス（任意）         |

### SQLite へインポート

```bash
papertail logs --hours 24 --db logs.db --tsv logs.tsv
```

