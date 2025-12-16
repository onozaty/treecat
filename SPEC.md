# treecat 仕様書

## 概要

treecatは、複数のファイルを1つのファイルにまとめ、ディレクトリツリー構造を可視化するGoベースのCLIツールです。主な用途は、LLM（大規模言語モデル）にコードベース全体の情報を渡すことです。

## 基本機能

### 入力
- ディレクトリパスを指定

### 出力
- 標準出力に以下の形式で出力：
  1. ディレクトリツリー構造（上部）
     - 最初の行：指定されたディレクトリパス（引数で指定されたパス、デフォルトは`.`）
     - 以降の行：ツリー構造（`├──`、`└──`、`│`などのボックス描画文字を使用）
  2. 各ファイルの内容（`=== filepath ===`で区切り）

### 出力フォーマット例
```
project
├── file1.go
└── dir/
    └── file2.go

=== file1.go ===
package main

func main() {
    // ...
}

=== dir/file2.go ===
package dir

// ...
```

## 詳細仕様

### ファイルフィルタリング

#### デフォルトの動作
1. `.git/`ディレクトリを自動的に除外
2. `.gitignore`ファイルが存在する場合、そのパターンを自動的に適用
3. 上記以外のすべてのファイルを対象とする

#### カスタムフィルタリング
- `--exclude`: 除外するGlobパターンを指定（カンマ区切り）
- `--include`: 含めるGlobパターンを指定（カンマ区切り）

#### フィルタリングの優先順位
1. `.git/`ディレクトリの除外（最優先）
2. `.gitignore`パターンの適用
3. `--exclude`パターンの適用
4. `--include`パターンの適用（指定時はこれにマッチするファイルのみ対象）

### バイナリファイルの扱い

- バイナリファイルの検出は行わない
- すべてのファイルをそのまま読み込んで出力
- バイナリファイルは文字化けした内容で出力される可能性がある

**推奨**: バイナリファイルは`.gitignore`または`--exclude`で除外すること

### パターンマッチング

Glob形式のパターンをサポート：

- `*` - 任意の文字列（ディレクトリ区切り文字を除く）
- `?` - 任意の1文字
- `**` - 任意の階層のディレクトリ
- `[abc]` - 文字クラス

#### パターン例
```bash
*.png              # すべてのPNGファイル
src/**/*.go        # srcディレクトリ以下のすべてのGoファイル
test_*.go          # test_で始まるGoファイル
**/.env*           # すべての環境変数ファイル
```

### ファイルの並び順

ファイルは相対パスの辞書順（lexicographic order）でソートされます。

例：
```
.gitignore
README.md
cmd/treecat/main.go
internal/filter/filter.go
internal/filter/filter_test.go
internal/scanner/scanner.go
```

### ツリー構造の表示

#### 空ディレクトリの除外

フィルタリング後に出力対象となるファイルが存在しないディレクトリは、ツリー表示から自動的に除外されます。

例：`--include "**/*.go"` でフィルタリングした場合

**ディレクトリ構成：**
```
project/
├── src/
│   └── main.go
├── docs/
│   └── README.md    # .goファイルではない
└── tests/
    └── test.go
```

**ツリー出力：**
```
├── src/
│   └── main.go
└── tests/
    └── test.go
```

`docs/` ディレクトリは `.go` ファイルを含まないため、ツリー表示から除外されます。

#### ディレクトリの表示順

ツリー内では、ディレクトリがファイルより先に表示されます。同じ種類（ディレクトリまたはファイル）の中では、名前のアルファベット順でソートされます。

### エラーハンドリング

#### 基本方針
- データの整合性を保証するため、エラー発生時は処理を中断する
- すべてのエラーは致命的エラーとして扱い、プログラムを終了する

#### エラー種別と対応

| エラー種別 | 対応 |
|-----------|------|
| 無効なディレクトリパス | 致命的エラー、使用方法を表示して終了 |
| ディレクトリの読み取り権限なし | 致命的エラー、エラーメッセージを表示して終了 |
| ファイルの読み取り権限なし | 致命的エラー、エラーメッセージを表示して終了 |
| ファイル読み取り中のエラー | 致命的エラー、エラーメッセージを表示して終了 |
| 無効なGlobパターン | 致命的エラー、エラーメッセージを表示して終了 |
| .gitignoreの読み取りエラー | 致命的エラー、エラーメッセージを表示して終了 |
| 対象ファイルが見つからない | 空のツリーを出力、正常終了 |

#### 終了コード
- `0`: 正常終了
- `1`: エラー発生時

## コマンドライン インターフェース

### 基本的な使い方

```bash
# カレントディレクトリを処理
treecat .

# 指定したディレクトリを処理
treecat /path/to/project

# 標準出力をファイルにリダイレクト
treecat /path/to/project > output.txt
```

### オプション

#### `--exclude <patterns>`
除外するファイルパターンをカンマ区切りで指定

```bash
treecat . --exclude "*.png,*.jpg,node_modules/**"
```

#### `--include <patterns>`
含めるファイルパターンをカンマ区切りで指定（このパターンにマッチするファイルのみを対象）

```bash
treecat . --include "**/*.go,**/*.md"
```

#### `--no-gitignore`
.gitignoreファイルを無視（将来の拡張として検討）

```bash
treecat . --no-gitignore
```

#### `--encoding <encoding-name>`
入力ファイルのエンコーディングを指定し、UTF-8に変換して出力

サポートされるエンコーディング（IANA標準）:
- **日本語**: `shift_jis`, `euc-jp`, `iso-2022-jp`
- **西欧言語**: `windows-1252`, `iso-8859-1`, `iso-8859-15`
- **中国語**: `gb2312`, `gbk`, `big5`
- **韓国語**: `euc-kr`
- その他、`golang.org/x/text/encoding/htmlindex` がサポートする全エンコーディング

エンコーディング名は大文字小文字を区別せず、アンダースコアとハイフンも同じ扱い
（例: `Shift_JIS`, `shift-jis`, `SHIFT_JIS` はすべて同じ）

```bash
# Shift_JISで書かれたファイルをUTF-8で出力
treecat . --encoding shift_jis

# EUC-JPファイルの変換
treecat legacy_code/ --encoding euc-jp > output.txt

# Windows-1252エンコードされたファイルを処理
treecat docs/ --encoding windows-1252
```

**動作**:
- オプション未指定時: バイナリとして読み込み、そのまま出力（既存の動作）
- `--encoding` 指定時: 指定エンコーディングで読み込み、UTF-8に変換して出力
- 未対応エンコーディング指定時: エラーメッセージを表示して終了

### 使用例

```bash
# Goプロジェクトのソースコードのみを出力
treecat . --include "**/*.go" > go-code.txt

# 画像ファイルを除外
treecat . --exclude "*.png,*.jpg,*.gif,*.svg" > output.txt

# 特定のディレクトリを除外
treecat . --exclude "vendor/**,node_modules/**,.next/**" > output.txt

# includeとexcludeの組み合わせ
treecat . --include "**/*.go" --exclude "**/*_test.go" > main-code.txt
```

## プロジェクト構造

```
/workspaces/treecat/
├── cmd/
│   └── treecat/
│       └── main.go              # CLIのエントリーポイント
├── internal/
│   ├── encoding/
│   │   ├── encoding.go          # エンコーディング変換
│   │   └── encoding_test.go     # エンコーディングテスト
│   ├── filter/
│   │   ├── filter.go            # フィルタリングロジック
│   │   └── filter_test.go       # フィルタのテスト
│   ├── scanner/
│   │   ├── scanner.go           # ディレクトリ走査
│   │   └── scanner_test.go      # スキャナのテスト
│   ├── tree/
│   │   ├── tree.go              # ツリー構造の生成とレンダリング
│   │   └── tree_test.go         # ツリーのテスト
│   └── output/
│       ├── output.go            # 出力フォーマット（エンコーディング変換統合）
│       └── output_test.go       # フォーマッタのテスト
├── testdata/                    # テスト用フィクスチャ（現在は空）
├── go.mod                       # Goモジュール定義
├── go.sum                       # 依存関係のチェックサム
├── .gitignore                   # Git除外設定
├── README.md                    # プロジェクト説明
└── SPEC.md                      # 本仕様書
```

## 依存ライブラリ

### 必須ライブラリ

1. **github.com/spf13/cobra**
   - 用途: CLIフレームワーク
   - 理由: 業界標準、優れたドキュメント、フラグ解析が容易

2. **github.com/go-git/go-git/v5/plumbing/format/gitignore**
   - 用途: .gitignoreファイルの解析
   - 理由: 最も広く使われている（176+パッケージ）、活発にメンテナンス、完全なgitignore仕様サポート

3. **github.com/bmatcuk/doublestar/v4**
   - 用途: Globパターンマッチング
   - 理由: `**`（globstar）の完全サポート、.gitignoreパターンとの互換性

4. **golang.org/x/text/encoding**
   - 用途: 文字エンコーディング変換（Shift_JIS, EUC-JP, etc. → UTF-8）
   - 理由: Goの公式サブリポジトリ、IANA標準エンコーディングの包括的サポート、htmlindexパッケージによる簡単なエンコーディング名解決

### 標準ライブラリ

- `filepath.WalkDir` - ディレクトリ走査
- `os.ReadFile` - ファイル読み取り
- `io.Writer` - 出力インターフェース
- `sort` - ファイルのソート

## 実装の詳細

### 主要なデータ構造

#### Filter インターフェース
```go
type Filter interface {
    ShouldInclude(path string, isDir bool) bool
}
```

ファイルやディレクトリを含めるべきかどうかを判定するインターフェース。

**実装**:
- `GitignoreFilter`: .gitignoreパターンに基づくフィルタ
- `PatternFilter`: include/excludeパターンに基づくフィルタ
- `CompositeFilter`: 複数のフィルタを組み合わせる

#### FileEntry 構造体
```go
type FileEntry struct {
    Path    string  // 絶対パス
    RelPath string  // ルートからの相対パス
    IsDir   bool    // ディレクトリかどうか
    Size    int64   // ファイルサイズ
}
```

スキャンされたファイルの情報を保持。

#### Node 構造体
```go
type Node struct {
    Name     string    // ファイル/ディレクトリ名
    Path     string    // 相対パス
    IsDir    bool      // ディレクトリかどうか
    Children []*Node   // 子ノード
}
```

ツリー構造を表現。

### 主要なアルゴリズム

#### 1. フィルタの合成

```
入力: ファイルパス、ディレクトリフラグ

1. .git/ディレクトリか？ → 除外
2. .gitignoreパターンにマッチ？ → 除外
3. excludeパターンにマッチ？ → 除外
4. includeパターンが指定されている？
   - YES: includeパターンにマッチ？ → 含める : 除外
   - NO: 含める
```

#### 2. ツリーの構築

```
入力: ソート済みのFileEntryリスト

1. ルートノードを作成
2. 各FileEntryに対して：
   - パスをセパレータで分割
   - ルートから順にノードを辿る/作成する
   - 最後のノードにファイル情報を設定
3. すべてのノードの子をソート（ディレクトリ優先、名前順）
```

#### 3. ツリーのレンダリング

```
入力: ノード、プレフィックス、最後の子フラグ

1. 現在のノードを出力
   - 最後の子: └──
   - それ以外: ├──
   - ディレクトリには / を付与

2. 子ノードを再帰的に処理
   - 親が最後の子でない: │   をプレフィックスに追加
   - 親が最後の子: スペースをプレフィックスに追加
```

### エッジケース

| ケース | 対応 |
|--------|------|
| シンボリックリンク | デフォルトでスキップ |
| 権限エラー | エラーで終了 |
| バイナリファイル | そのまま読み込み（検出しない） |
| 空のディレクトリ | ツリーには表示、内容セクションなし |
| 巨大なファイル | サイズ制限なし（すべて読み込む） |
| 非UTF-8ファイル名 | 生バイトを使用（Goが自然に処理） |
| 隠しファイル | デフォルトで含める |
| .gitignoreなし | 通常通り継続 |
| 複数の.gitignore | ルートの.gitignoreのみ処理（v1） |
| excludeとincludeの競合 | excludeが先に評価され、その後includeをチェック |
| 空のディレクトリ引数 | カレントディレクトリを使用 |

## ビルドとテスト

### ビルド

```bash
# ビルド
go build -o bin/treecat ./cmd/treecat

# インストール
go install ./cmd/treecat
```

### テスト戦略

#### ユニットテスト
各パッケージごとにテストを作成：

- `filter_test.go`: パターンマッチング、.gitignore統合、フィルタ合成
- `scanner_test.go`: ディレクトリ走査、フィルタ適用、ソート
- `tree_test.go`: ツリー構築、レンダリング、ネスト構造
- `output_test.go`: 出力フォーマット、エラーハンドリング

#### 統合テスト
`main_test.go`にて実装：

- 基本的なディレクトリ出力
- include/excludeパターン
- .gitignoreサポートと--no-gitignoreフラグ
- .gitディレクトリの除外
- ネストされたディレクトリ構造

テストは`t.TempDir()`を使用して一時ディレクトリで実行

#### テストの実行
```bash
# すべてのテスト
go test ./...

# 詳細出力
go test -v ./...

# カバレッジ
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 実装フェーズ

### Phase 1: プロジェクトのブートストラップ
1. Goモジュールの初期化
2. ディレクトリ構造の作成
3. .gitignoreの追加
4. 依存ライブラリのインストール
5. 基本的なmain.goの作成
6. Makefileの作成

### Phase 2: filterパッケージ
1. Filterインターフェースの定義
2. GitignoreFilterの実装
3. PatternFilterの実装
4. CompositeFilterの実装
5. ユニットテストの作成

### Phase 3: scannerパッケージ
1. FileEntry構造体の定義
2. Scannerの実装
3. filepath.WalkDirを使用した走査
4. フィルタの適用
5. ファイルのソート
6. ユニットテストの作成

### Phase 4: treeパッケージ
1. Node構造体の定義（tree.go）
2. ツリー構築アルゴリズムの実装
3. ツリーレンダリングの実装
4. ユニットテストの作成（tree_test.go）

### Phase 5: outputパッケージ
1. Formatterの実装（output.go）
2. ツリーセクションの出力
3. ファイル内容セクションの出力
4. エラーハンドリング
5. ユニットテストの作成（output_test.go）

### Phase 6: CLI統合
1. cobraコマンドのセットアップ
2. フラグの定義（cmd.Flags()から取得する方式）
3. 各コンポーネントの統合
4. エラーハンドリング
5. 統合テストの作成（main_test.go）

### Phase 7: ドキュメント
1. README.mdの作成
2. 使用例の追加
3. コードコメントの追加

## 成功基準

- [x] ディレクトリをスキャンしてツリー+内容を出力できる
- [x] .gitignoreパターンを正しく適用できる
- [x] カスタムinclude/exclude Globパターンをサポートできる
- [x] .git/ディレクトリを常に除外できる
- [x] 標準出力に出力できる（パイプフレンドリー）
- [x] エラーを適切にハンドリングできる（エラー発生時は即座に終了）
- [x] 適切なテストカバレッジがある（ユニット+統合）
- [ ] 明確なドキュメントと使用例がある（README.md未完成）

## 将来の拡張（スコープ外）

以下の機能はv1では実装せず、将来のバージョンで検討：

- 複数の.gitignoreサポート（ネストされたディレクトリ）
- `--max-file-size`フラグ（巨大ファイルのスキップ）
- トークン数のカウント表示
- 複数の出力フォーマット（JSON、XMLなど）
- 並列ファイル読み取り
- プログレスバー表示
- 設定ファイルのサポート（.treecatrc）
- ファイル重要度による並び替え（yekのような機能）
- バイナリファイルの自動検出と除外
- シンボリックリンクのフォロー機能
