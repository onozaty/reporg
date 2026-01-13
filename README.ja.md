# reporg

[![Test](https://github.com/onozaty/reporg/actions/workflows/test.yaml/badge.svg)](https://github.com/onozaty/reporg/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/onozaty/reporg/branch/main/graph/badge.svg)](https://codecov.io/gh/onozaty/reporg)
[![GitHub release](https://img.shields.io/github/release/onozaty/reporg.svg)](https://github.com/onozaty/reporg/releases/latest)
[![License](https://img.shields.io/github/license/onozaty/reporg.svg)](LICENSE)

[English](README.md) | 日本語

**reporg** は、Git リポジトリを [ripgrep](https://github.com/BurntSushi/ripgrep) で検索し、検索結果を共有・参照しやすい形式(TSV)で出力する CLI ツールです。

## 特徴

- **ローカル検索**: ripgrep を使用した高速な全文検索
- **GitHub URL 生成**: 各検索結果に対応する GitHub 上の該当行 URL を自動生成
- **TSV 形式出力**: スプレッドシートや他のツールで簡単に処理可能
- **複数リポジトリ対応**: 一度に複数のリポジトリを検索可能
- **豊富な検索オプション**: 大文字小文字の区別、Glob パターン、隠しファイル検索、固定文字列検索など

## インストール

### Homebrew (macOS/Linux)

```bash
brew install onozaty/tap/reporg
```

### Scoop (Windows)

```bash
scoop bucket add onozaty https://github.com/onozaty/scoop-bucket
scoop install reporg
```

### Go install

```bash
go install github.com/onozaty/reporg@latest
```

### バイナリダウンロード

[Releases](https://github.com/onozaty/reporg/releases) ページから、お使いのプラットフォーム用のバイナリをダウンロードしてください。

## 前提条件

reporg を使用するには、[ripgrep](https://github.com/BurntSushi/ripgrep) がインストールされている必要があります。

```bash
# macOS
brew install ripgrep

# Ubuntu/Debian
sudo apt-get install ripgrep

# Windows (Scoop)
scoop install ripgrep
```

詳細は [ripgrep のインストールガイド](https://github.com/BurntSushi/ripgrep#installation)を参照してください。

## 検索の動作

reporg は ripgrep を使用して検索を行います。以下のファイル・ディレクトリは自動的にスキップされます:

- `.gitignore` に記載されたファイル・ディレクトリ
- `.ignore` や `.rgignore` に記載されたファイル・ディレクトリ
- 隠しファイル・ディレクトリ(`.` で始まるもの) - `--hidden` オプションで検索対象に含めることができます

## 使い方

### 基本的な使い方

```bash
reporg <pattern> <repoRoot1> [repoRoot2...]
```

- `pattern`: 検索パターン(正規表現)
- `repoRoot`: Git リポジトリのルートディレクトリ(複数指定可能)

**例:**

```bash
# カレントディレクトリで "TODO" を検索
reporg "TODO" .

# 複数のリポジトリを検索
reporg "TODO" /path/to/repo1 /path/to/repo2
```

### 出力形式

検索結果は TSV(タブ区切り)形式で出力されます。

```
owner/repo	src/main.go:12	// TODO: refactor	https://github.com/owner/repo/blob/main/src/main.go#L12
owner/repo	src/utils.go:25	// TODO: optimize	https://github.com/owner/repo/blob/main/src/utils.go#L25
```

**列の説明(タブ区切り):**

1. `repository`: `owner/repo` 形式のリポジトリ識別子
2. `local_path`: ファイルパスと行番号(`path/to/file:LINE` 形式)
3. `matched_line`: 一致した行の内容
4. `github_url`: GitHub 上の該当行 URL

### オプション

#### 出力先

```bash
# ファイルに出力
reporg "TODO" . -o result.tsv
```

#### 検索オプション

**大文字小文字を区別しない検索:**

```bash
# "TODO", "todo", "Todo" すべてにマッチ
reporg -i "todo" /path/to/repo
```

**Glob パターンでファイルをフィルタリング:**

```bash
# Go ファイルのみ検索
reporg "package" /repo -g "*.go"

# Go ファイルを検索、ただしテストファイルを除外
reporg "package" /repo -g "*.go" -g "!*_test.go"

# 特定ディレクトリのみ検索
reporg "TODO" /repo -g "src/**"
```

**隠しファイルも検索:**

```bash
# 隠しファイルも含めて検索
reporg "secret" /repo --hidden

# .git ディレクトリを除外しつつ他の隠しファイルを検索
reporg "config" /repo --hidden -g "!.git/**"
```

**固定文字列として検索(正規表現ではなく):**

```bash
# "main()" を正規表現ではなく文字列として検索
reporg -F "main()" /repo

# 括弧を含むパターンを検索
reporg -F "if (x > 0) {" /repo
```

**出力する行の長さを制限:**

```bash
# 行を 500 文字に制限（minified ファイルなどで有用）
reporg "pattern" /repo -m 500

# 1000 文字に制限してファイルに出力
reporg "TODO" /repo --max-line-length 1000 -o results.tsv
```

**オプションの組み合わせ:**

```bash
# Go ファイルから大文字小文字を区別せずに "TODO" を検索し、結果をファイルに保存
reporg -i "todo" /repo -g "*.go" -o results.tsv

# 隠しファイルを含めて固定文字列として検索
reporg -F "config.value" /repo --hidden -o config-usage.tsv

# 複数のリポジトリに対して複数条件で検索
reporg -i "error" /repo1 /repo2 -g "*.go" -g "!vendor/**" --hidden
```

### 全オプション一覧

```
  -o, --output string           出力先ファイルパス(未指定時は stdout)
  -i, --ignore-case             大文字小文字を区別しない検索
  -g, --glob pattern            Glob パターンでファイルをフィルタリング(複数指定可能)
      --hidden                  隠しファイル・ディレクトリも検索対象に含める
  -F, --fixed-strings           パターンを正規表現ではなく固定文字列として扱う
  -m, --max-line-length int     出力する行の最大文字数(0 = 制限なし)。指定した長さを超える行は '...' で切り詰められる
  -h, --help                    ヘルプを表示
  -v, --version                 バージョン情報を表示
```

## 制限事項

- **GitHub のみ対応**: 現在、GitHub リポジトリのみサポートしています
- **Git リポジトリルートが必須**: 指定するパスは Git リポジトリのルートディレクトリである必要があります(サブディレクトリ指定はエラー)

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 作者

[onozaty](https://github.com/onozaty)
