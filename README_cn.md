# Papertail

从 Papertrail API 下载和处理日志的 CLI 工具。

## 安装

```bash
go install github.com/wakuwaku-lab/papertail@latest
```

或从源码构建:

```bash
git clone https://github.com/wakuwaku-lab/papertail.git
cd papertail
go build -o papertail .
```

## 配置

支持两种方式配置 API Token:

**方式1: 环境变量**
```bash
export PAPERTRAIL_API_TOKEN="your_token_here"
```

**方式2: .env 文件**
```bash
# 在项目目录创建 .env 文件
echo "PAPERTRAIL_API_TOKEN=your_token_here" > .env
```

程序会自动加载当前目录下的 `.env` 文件。

## 使用方法

### 三种模式

**1. 按小时下载**
```bash
# 最近 24 小时
papertail logs --hours 24 --tsv output.tsv

# 指定小时范围 (1-6 小时前)
papertail logs --hours 1-6 --tsv output.tsv
```

**2. 按特定日期下载**
```bash
papertail logs --date 2024-03-01 --tsv output.tsv
```

**3. 按日期范围下载**
```bash
papertail logs --start 2024-03-01 --end 2024-03-07 --tsv output.tsv
```

### 参数说明

| 参数 | 简写 | 说明 |
|------|------|------|
| `--hours` | `-n` | 小时数，支持 `24` 或 `1-12` 格式 |
| `--date` | `-D` | 特定日期 (YYYY-MM-DD) |
| `--start` | `-s` | 日期范围起始日期 |
| `--end` | `-e` | 日期范围结束日期 |
| `--tsv` | `-t` | 输出 TSV 文件路径 |
| `--db` | `-d` | SQLite 数据库路径 (可选) |

### 导入 SQLite

```bash
papertail logs --hours 24 --db logs.db --tsv logs.tsv
```

## 依赖

- **7z**: 解压 `.tsv.gz` 文件 (`sudo apt-get install p7zip-full`)
- **sqlite-utils**: 导入 SQLite (`pip install sqlite-utils`)
