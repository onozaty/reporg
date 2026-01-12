# reporg 仕様書

## 1. 概要

**reporg** は、Git リポジトリを ripgrep で検索し、検索結果を共有・参照しやすい形式（TSV）で出力する CLI ツールです。

検索結果には以下を含めます：

* ローカルのファイルパス（リポジトリルートからの相対パス）
* 一致した行の内容
* GitHub 上の該当行 URL

---

## 2. 基本コンセプト

* **検索は必ずローカル**
* **Git 管理されているリポジトリのみが対象**
* GitHub API やリモート検索は行わない
* 検索結果を「調査・共有・レビュー」に使える形で出力する

---

## 3. 対象範囲（Git リポジトリ限定）

* **指定する path は Git リポジトリのルートであることを前提**とする
* 各 path に対して以下を実行

  ```bash
  git -C <path> rev-parse --show-toplevel
  ```
* 上記の結果が **path 自身と一致する場合のみ有効**とする

  * 一致しない場合（サブディレクトリ指定など）は **エラーとして終了**
* Git リポジトリでない path も **エラーとして終了**
* 同一リポジトリが複数指定された場合は **1回のみ処理**

この仕様により、検索対象は常に **明示的に指定された Git リポジトリ単位**となる。

---

## 4. コマンド仕様

```bash
reporg [options] <pattern> <repoRoot1> [repoRoot2] ...
```

### 引数

* `pattern`（必須）

  * ripgrep に渡す検索パターン
* `repoRoot...`（必須）

  * **Git リポジトリのルートディレクトリ**
  * 複数指定可

※ サブディレクトリ指定は **エラー** とする

---

## 5. 検索処理の流れ

1. 指定された各 path を **Git リポジトリのルート候補**として扱う
2. 各 path について

   ```bash
   git -C <path> rev-parse --show-toplevel
   ```

   を実行
3. コマンド結果が **指定 path と一致する場合のみ**対象リポジトリとして採用
4. 採用された Git リポジトリ root を重複排除
5. 各リポジトリに対して

   ```bash
   rg --json [options] <pattern> <repoRoot>
   ```

   を実行

   * オプションには `-i`, `--glob`, `--hidden`, `-F` などが含まれる
   * ripgrep はデフォルトで `.gitignore` に記載されたファイルを自動的にスキップ
   * `.ignore` や `.rgignore` ファイルにも対応
   * `--hidden` を指定しない限り、隠しファイル・ディレクトリはスキップされる
6. `rg` の JSON 出力を解析し、`match` イベントのみ処理
7. 以下の情報を生成

   * ローカルパス（repo root からの相対）
   * 一致行のテキスト
   * GitHub 上の該当行 URL

---

## 6. 出力仕様（TSV）

### 出力形式

* **1行 = 1ヒット**
* **TSV（タブ区切り）**

### 列構成

1. `repository`

   * `owner/repo` 形式のリポジトリ識別子
2. `local_path`

   * `path/to/file:LINE` 形式
3. `matched_line`

   * 一致した行の内容
4. `github_url`

   * GitHub 上の該当行 URL

### 出力例

```text
owner/repo	src/main.go:12	// TODO: refactor	https://github.com/owner/repo/blob/main/src/main.go#L12
```

---

## 7. 出力先オプション

### 標準出力

```bash
reporg "TODO" .
```

### ファイル出力

```bash
reporg "TODO" . -o result.tsv
```

#### オプション

* `-o <path>`：出力先ファイルを指定（未指定時は stdout）

---

## 8. GitHub URL 生成方針

* **GitHub のみ対応**（将来拡張前提）
* 対応する remote URL
  * `https://github.com/owner/repo.git`
  * `git@github.com:owner/repo.git`
* URL 形式

```text
https://github.com/{owner}/{repo}/blob/{branch}/{path}#L{line}
```

* branch は以下の優先順位で決定

  1. `git branch --show-current`
  2. fallback: `main`

---

## 9. オプション仕様

### 出力先

* `-o <path>`、`--output <path>`：出力先ファイル指定（未指定時は stdout）

### 検索オプション

以下のオプションは ripgrep に渡され、検索動作をカスタマイズします:

#### 大文字小文字の区別

* `-i`、`--ignore-case`：大文字小文字を区別しない検索

**使用例:**
```bash
# "TODO", "todo", "Todo" すべてにマッチ
reporg -i "todo" /path/to/repo
```

#### ファイルフィルタリング（Glob パターン）

* `-g <pattern>`、`--glob <pattern>`：指定したパターンにマッチするファイルのみ検索（複数指定可能）
  * `*.ext` 形式でファイルタイプを指定
  * `!pattern` で除外パターンを指定

**使用例:**
```bash
# Go ファイルのみ検索
reporg "package" /repo -g "*.go"

# Go ファイルを検索、ただしテストファイルを除外
reporg "package" /repo -g "*.go" -g "!*_test.go"

# 特定ディレクトリのみ検索
reporg "TODO" /repo -g "src/**"

# .git ディレクトリを除外
reporg "pattern" /repo --hidden -g "!.git/**"
```

#### 隠しファイル検索

* `--hidden`：隠しファイル・ディレクトリも検索対象に含める
  * デフォルトでは `.` で始まるファイル・ディレクトリはスキップされる
  * このオプションを使用すると `.env`, `.git/config` なども検索される

**使用例:**
```bash
# 隠しファイルも含めて検索
reporg "secret" /repo --hidden

# .git ディレクトリを除外しつつ他の隠しファイルを検索
reporg "config" /repo --hidden -g "!.git/**"
```

#### 固定文字列検索

* `-F`、`--fixed-strings`：パターンを正規表現ではなく固定文字列として扱う
  * `()`, `[]`, `.`, `*` などの正規表現メタ文字をエスケープ不要で検索可能

**使用例:**
```bash
# "main()" を正規表現ではなく文字列として検索
reporg -F "main()" /repo

# 括弧を含むパターンを検索
reporg -F "if (x > 0) {" /repo

# ドットやアスタリスクを含むパターンを検索
reporg -F "*.txt" /repo
```

### オプションの組み合わせ

複数のオプションを同時に使用可能:

```bash
# Go ファイルから大文字小文字を区別せずに "TODO" を検索し、結果をファイルに保存
reporg -i "todo" /repo -g "*.go" -o results.tsv

# 隠しファイルを含めて固定文字列として検索
reporg -F "config.value" /repo --hidden -o config-usage.tsv

# 複数のリポジトリに対して複数条件で検索
reporg -i "error" /repo1 /repo2 -g "*.go" -g "!vendor/**" --hidden
```
