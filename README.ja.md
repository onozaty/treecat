# treecat

[![GitHub license](https://img.shields.io/github/license/onozaty/treecat)](https://github.com/onozaty/treecat/blob/main/LICENSE)
[![Test](https://github.com/onozaty/treecat/actions/workflows/test.yaml/badge.svg)](https://github.com/onozaty/treecat/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/onozaty/treecat/graph/badge.svg?token=V114XLWDBR)](https://codecov.io/gh/onozaty/treecat)

[English](README.md) | 日本語


> ディレクトリツリーの可視化と複数ファイルの統合出力を提供する、LLM向けに最適化されたツール

**treecat**は、ディレクトリ内の複数ファイルを単一の出力に集約するコマンドラインツールです。上部にディレクトリツリー構造を表示し、その後に各ファイルの内容を続けて出力します。Claude Code、ChatGPT、GitHub Copilotなどの大規模言語モデル(LLM)に包括的なコードベースコンテキストを提供することを目的として設計されています。

AIアシスタントへのプロジェクト構造全体の提供、コードレビュー、ドキュメント生成、そして複数ファイルとそのディレクトリ階層の統一ビューが必要なあらゆるシナリオに最適です。

## 機能

- **ディレクトリツリーの可視化とファイル内容の集約** - ツリー構造とファイル内容を組み合わせた単一出力
- **LLMコンテキスト向けに最適化** - Claude Code、ChatGPT、GitHub Copilotなどのコーディングアシスタントに最適
- **自動.gitignoreパターン適用** - デフォルトでプロジェクトの.gitignoreルールを尊重
- **柔軟なglobパターンフィルタリング** - 強力な`--include`と`--exclude`パターンでファイルを含める/除外
- **文字エンコーディング変換** - UTF-8以外のエンコーディング(Shift_JIS、EUC-JP、GB2312など)をUTF-8に変換
- **UTF-8 BOM削除と改行正規化** - 一貫した出力を保証(CRLF → LF)
- **空ディレクトリの除外** - フィルタリング後の空ディレクトリを自動的に除外

## インストール

最新のバイナリを[GitHub Releases](https://github.com/onozaty/treecat/releases/latest)からダウンロードしてください。

## クイックスタート

```bash
# 現在のディレクトリをファイルに出力
treecat . > output.txt

# 特定のディレクトリを出力
treecat /path/to/project > output.txt

# ファイルパターンでフィルタリング
treecat . --include "**/*.go,**/*.md" > output.txt

# パターンで除外
treecat . --exclude "*.png,*.jpg,node_modules/**" > output.txt
```

## 使用方法

### 基本的な使い方

```bash
treecat [directory] [flags]
```

ディレクトリが指定されない場合、現在のディレクトリ(`.`)が使用されます。

### コマンドラインオプション

#### フィルタリングオプション

**`-i, --include <patterns>`**

指定されたglobパターン(カンマ区切り)に一致するファイルのみを含めます。指定された場合、これらのパターンに一致するファイルのみが出力に含まれます。

```bash
treecat . --include "**/*.go,**/*.md"
```

**`-e, --exclude <patterns>`**

指定されたglobパターン(カンマ区切り)に一致するファイルを除外します。これらのパターンは.gitignoreルールの後に適用されます。

```bash
treecat . --exclude "*.log,*.tmp,dist/**"
```

**`--no-gitignore`**

`.gitignore`ファイルのパターンを無視します。デフォルトでは、treecatは対象ディレクトリ内の`.gitignore`ルールを自動的に適用します。

```bash
treecat . --no-gitignore
```

#### エンコーディングオプション

**`--encoding-map <extension:encoding,...>`**

ファイル拡張子ごとに文字エンコーディングを指定します。指定された拡張子に一致するファイルは、元のエンコーディングから出力時にUTF-8に変換されます。

**サポートされるエンコーディング**(IANA標準名):
- **日本語**: `shift_jis`、`euc-jp`、`iso-2022-jp`
- **西ヨーロッパ**: `windows-1252`、`iso-8859-1`、`iso-8859-15`
- **中国語**: `gb2312`、`gbk`、`big5`
- **韓国語**: `euc-kr`
- その他多数(`golang.org/x/text/encoding/htmlindex`経由)

**注意事項:**
- エンコーディング名は大文字小文字を区別しません(`Shift_JIS`、`shift-jis`、`SHIFT_JIS`はすべて同等)
- アンダースコアとハイフンは互換性があります(`shift_jis` = `shift-jis`)
- 拡張子は先頭のドットなしで指定し、大文字小文字を区別しません(`txt`、`TXT`、`.txt`すべて動作)

```bash
# 単一の拡張子
treecat . --encoding-map "txt:shift_jis"

# 異なるエンコーディングを持つ複数の拡張子
treecat legacy/ --encoding-map "txt:shift_jis,log:euc-jp,csv:windows-1252"
```

**エンコーディングの動作:**
- **BOM(バイトオーダーマーク)削除**: すべてのファイルからUTF-8 BOMが自動的に削除されます
- **改行の正規化**: すべての改行(CRLF、CR、LF)がLF(`\n`)に変換されます
- **マップされていないファイル**: エンコーディングマップにないファイルはUTF-8として扱われます

### フィルタリングの優先順位

フィルタは以下の順序で適用されます:

1. **`.git/`ディレクトリの除外** - 常に適用(最高優先度)
2. **`.gitignore`パターン** - `--no-gitignore`が指定されない限り適用
3. **`--exclude`パターン** - ユーザー指定の除外パターン
4. **`--include`パターン** - 指定された場合、一致するファイルのみが含まれます(最低優先度)

### 例

#### ファイルタイプでフィルタリング

```bash
# Goソースファイルのみ
treecat . --include "**/*.go" > go-code.txt

# ソースコードとドキュメント
treecat . --include "**/*.go,**/*.md,**/*.txt" > project.txt
```

#### ファイルとディレクトリを除外

```bash
# 画像ファイルを除外
treecat . --exclude "*.png,*.jpg,*.gif,*.svg" > output.txt

# 一般的なビルド/依存関係ディレクトリを除外
treecat . --exclude "vendor/**,node_modules/**,dist/**,.next/**" > output.txt
```

#### includeとexcludeを組み合わせる

```bash
# テストを除くGoファイル
treecat . --include "**/*.go" --exclude "**/*_test.go" > main-code.txt

# 仮想環境を除くPythonソース
treecat . --include "**/*.py" --exclude "venv/**,**/__pycache__/**" > python-code.txt
```

#### 文字エンコーディング変換

```bash
# Shift_JISエンコードの.txtファイルをUTF-8に変換
treecat legacy-docs/ --encoding-map "txt:shift_jis" > output.txt

# 異なるファイルタイプに複数のエンコーディング
treecat old-project/ --encoding-map "txt:shift_jis,log:euc-jp,csv:windows-1252" > output.txt

# 混在エンコーディングを持つ日本語レガシーコードベース
treecat src/ --encoding-map "txt:shift_jis,md:euc-jp" > codebase.txt
```

## 出力形式

出力は2つの主要セクションで構成されます:

```
project-root/
├── file1.go
├── README.md
└── src/
    └── main.go

=== file1.go ===
package main

func main() {
    // ...
}

=== README.md ===
# Project

Documentation here...

=== src/main.go ===
package main

// Additional code...
```

**1. ディレクトリツリー**(上部セクション):
- ルートディレクトリ名を含む階層構造を表示
- Unicode罫線文字(`├──`、`└──`、`│`)を使用
- フィルタリング後の空ディレクトリは自動的に除外
- ディレクトリは同じレベルのファイルの前に表示
- 各レベル内で辞書順ソート

**2. ファイル内容**(ツリーの下):
- 各ファイルは`=== filepath ===`マーカーで区切られます
- ファイルパスは指定されたルートディレクトリからの相対パス
- ファイルは相対パスの辞書順で表示

## ライセンス

MIT License

詳細は[LICENSE](LICENSE)ファイルを参照してください。

## 作者

[onozaty](https://github.com/onozaty)
