# Aivis Cloud CLI

Aivis Cloud API を使用して音声合成と音声再生を行うコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## このパッケージでできること

- **音声合成（TTS）**: テキストを自然な音声に変換
- **柔軟な音声再生**: 同期・非同期再生、キュー管理による複数音声制御
- **MCP（Model Context Protocol）対応**: Claude などの AI アシスタントが直接音声再生可能
- **豊富な音声モデル**: 人気・最新・作者別など多様な検索とモデル管理
- **設定管理**: API キーや各種設定の保存・管理

## インストール

### npx を使用（推奨）

インストール不要で直接実行：

```bash
npx @kajidog/aivis-cloud-cli --help
```

### グローバルインストール

```bash
npm install -g @kajidog/aivis-cloud-cli
aivis-cloud-cli --help
```

## 事前準備

Aivis Cloud API キーが必要です。

### API キーの取得

**ダッシュボード**: https://hub.aivis-project.com/cloud-api/dashboard

ダッシュボードにアクセスして API キーを取得してください。

### API キーの設定

取得した API キーを以下のいずれかの方法で設定してください：

1. 環境変数: `export AIVIS_API_KEY="your-api-key"`
2. コマンドフラグ: `--api-key "your-api-key"`
3. 設定ファイル: `aivis-cloud-cli config set api_key "your-api-key"`

## 基本的な使い方

### 音声合成（TTS）

<details>
<summary>テキストを音声に変換</summary>

```bash
# 基本的な音声合成（デフォルトモデルを使用、出力ファイル必須）
npx @kajidog/aivis-cloud-cli tts synthesize --text "こんにちは世界" --output "output.wav"

# 位置引数を使用した音声ファイル保存
npx @kajidog/aivis-cloud-cli tts synthesize "こんにちは" "output.wav"

# 特定のモデルを指定
npx @kajidog/aivis-cloud-cli tts synthesize --text "こんにちは世界" --output "output.wav" --model-uuid "model-id"

# SSML マークアップを使用
npx @kajidog/aivis-cloud-cli tts synthesize --text '<speak>こんにちは<break time="1s"/>世界</speak>' --output "output.wav" --ssml
```

</details>

<details>
<summary>音声の即時再生</summary>

```bash
# テキストを音声に変換してすぐに再生（デフォルトモデルを使用）
npx @kajidog/aivis-cloud-cli tts play --text "こんにちは世界"

# 特定のモデルを指定して再生
npx @kajidog/aivis-cloud-cli tts play --text "こんにちは世界" --model-uuid "model-id"
```

</details>

<details>
<summary>詳細なパラメータ設定</summary>

```bash
npx @kajidog/aivis-cloud-cli tts synthesize \
  --text "こんにちは世界" \
  --output "output.mp3" \
  --format mp3 \
  --channels stereo \
  --rate 1.2 \
  --pitch 0.8 \
  --volume 0.9 \
  --leading-silence 0.1 \
  --trailing-silence 0.2 \
  --sampling-rate 44100 \
  --bitrate 128
```

**注意**: `--model-uuid` を指定しない場合、システムはデフォルトモデル（`a59cb814-0083-4369-8542-f51a29e72af7`）を使用します。

</details>

<details>
<summary>ストリーミング合成</summary>

```bash
# ストリーミング合成（リアルタイム出力、標準出力に音声データを出力）
npx @kajidog/aivis-cloud-cli tts stream --text "こんにちは世界" > output.wav
```

</details>

### 音声モデル管理

<details>
<summary>モデル検索</summary>

```bash
# 日本語モデルを検索
npx @kajidog/aivis-cloud-cli models search --query "japanese"

# 人気のモデルを表示（ダウンロード数順）
npx @kajidog/aivis-cloud-cli models search --sort "downloads" --limit 10

# 最新のモデルを表示
npx @kajidog/aivis-cloud-cli models search --sort "created_at" --limit 5

# 特定の作者のモデルを検索
npx @kajidog/aivis-cloud-cli models search --author "作者名"

# 全モデルを表示
npx @kajidog/aivis-cloud-cli models search

# 詳細情報を表示
npx @kajidog/aivis-cloud-cli models search --verbose
```

</details>

<details>
<summary>特定モデルの詳細取得</summary>

```bash
npx @kajidog/aivis-cloud-cli models get --uuid "model-id"
```

</details>

### 設定管理

<details>
<summary>基本設定</summary>

```bash
# APIキーの設定
npx @kajidog/aivis-cloud-cli config set api_key "your-api-key"

# カスタムエンドポイントの設定
npx @kajidog/aivis-cloud-cli config set base_url "https://api.example.com"

# 現在の設定を表示
npx @kajidog/aivis-cloud-cli config show
```

</details>

## MCP サーバー機能

この CLI は MCP（Model Context Protocol）サーバーとして動作し、AI アシスタント（Claude など）に AivisCloud の音声合成機能を提供します。

<details>
<summary>MCPサーバーの起動</summary>

事前に API キーを設定してください：

```bash
# 設定ファイルにAPIキーを保存（推奨）
npx @kajidog/aivis-cloud-cli config set api_key "your-api-key"

# MCPサーバーを起動（stdio デフォルト）
npx @kajidog/aivis-cloud-cli mcp

# HTTPモードで起動（デフォルトポート8080）
npx @kajidog/aivis-cloud-cli mcp --transport http

# HTTPモードでカスタムポート
npx @kajidog/aivis-cloud-cli mcp --transport http --port 3000
```

</details>

<details>
<summary>Claude Desktop / Claude Code との連携</summary>

### Claude Desktop

Claude Desktop の設定ファイル（`~/Library/Application Support/Claude/claude_desktop_config.json`）に以下を追加：

**stdio モード（推奨）:**
```json
{
  "mcpServers": {
    "aivis-cloud-api": {
      "command": "npx",
      "args": ["@kajidog/aivis-cloud-cli", "mcp"],
      "env": {
        "AIVIS_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

- **Claude Desktop が自動的にMCPサーバーを起動・管理**
- **API キーが設定済みの場合**: `env` セクションは省略可能（設定ファイルまたは環境変数から読み込み）
- **プロセス管理不要**: Claude Desktop終了時に自動停止

**HTTP モード（リモートアクセス・デバッグ用）:**

まず、MCPサーバーを別途起動しておく必要があります：
```bash
# ターミナルでMCPサーバーを起動（常時実行）
npx @kajidog/aivis-cloud-cli mcp --transport http --port 8080
```

次に、Claude Desktop設定：
```json
{
  "mcpServers": {
    "aivis-cloud-api": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:8080"]
    }
  }
}
```

- **事前にサーバー起動が必要**: 上記のコマンドを実行し続ける必要があります
- **デバッグやリモート接続に有用**: 複数のクライアントから接続可能

### Claude Code CLI

Claude Code CLI を使用している場合は、以下のコマンドで追加できます：

**stdio モード（推奨）:**
```bash
# MCP サーバーを追加（stdio）
claude mcp add aivis npx @kajidog/aivis-cloud-cli mcp
```

- **Claude Code が自動的にMCPサーバーを起動・管理**
- **API キーが設定済みの場合**: 環境変数 `AIVIS_API_KEY` または設定ファイルから自動読み込み
- **プロセス管理不要**: Claude Code終了時に自動停止

**HTTP モード（リモートアクセス・デバッグ用）:**

まず、MCPサーバーを別途起動しておく必要があります：
```bash
# ターミナルでMCPサーバーを起動（常時実行）
npx @kajidog/aivis-cloud-cli mcp --transport http --port 8080
```

次に、Claude Code に追加：
```bash
# MCP サーバーを追加（デフォルトポート8080）
claude mcp add --transport http aivis http://localhost:8080

# カスタムポートの場合
claude mcp add --transport http aivis http://localhost:3000
```

- **事前にサーバー起動が必要**: 上記のコマンドでサーバーを実行し続ける必要があります  
- **デバッグやリモート接続に有用**: 複数のClaude Codeセッションから同じサーバーに接続可能

</details>

<details>
<summary>利用可能なMCPツール</summary>

MCP サーバーは以下のツールを AI アシスタントに提供します：

**音声モデル関連:**

- **search_models**: 音声モデルの検索（デフォルト 5 件）

  - パラメータ: `query`, `author`, `tags`, `limit`, `sort`, `public_only`

- **get_model**: モデルの基本情報取得

  - パラメータ: `uuid` (省略時は設定ファイルの `default_model_uuid` またはフォールバックモデルを使用)

- **get_model_speakers**: モデルのスピーカー情報取得
  - パラメータ: `uuid` (省略時は設定ファイルの `default_model_uuid` またはフォールバックモデルを使用)

**音声合成・再生関連:**

- **synthesize_speech**: テキストを音声に変換してサーバー上で再生（フル機能版）

  - パラメータ: `text` (必須), `model_uuid`, `format`, `volume`, `rate`, `pitch`, `playback_mode`, `wait_for_end`
  - 音声フォーマット: `wav`, `mp3`, `flac`, `aac`, `opus`
  - 再生モード: `immediate` (即座再生), `queue` (キュー追加), `no_queue` (同時再生)

- **play_text**: デフォルト設定でテキストを音声再生（簡易版）
  - パラメータ: `text` (必須), `playback_mode`, `wait_for_end`
  - 注意: `default_model_uuid` と `use_simplified_tts_tools: true` が設定されている場合のみ利用可能

**設定管理関連:**

- **get_mcp_settings**: 現在のMCP設定を取得
  - パラメータ: なし
  - 戻り値: 現在の設定値（APIキーは除外）
  - セキュリティのため、API キーとシステム設定（ログ設定、簡易TTS設定）は表示されません

- **update_mcp_settings**: MCP設定を安全に更新
  - パラメータ: `base_url`, `default_model_uuid`, `default_playback_mode`, `default_volume`, `default_rate`, `default_pitch`, `default_format`
  - 制限: APIキー、ログ設定、`use_simplified_tts_tools` は変更不可
  - 設定値のバリデーション機能付き（例：音量は0.0-2.0の範囲）

**使用例:**
```javascript
// 現在の設定を確認
get_mcp_settings({})

// 設定を更新
update_mcp_settings({
  "default_volume": 0.8,
  "default_playback_mode": "queue",
  "default_format": "mp3"
})
```

</details>

## 対応プラットフォーム

以下のプラットフォーム用のバイナリが含まれています：

- **Linux**: x64, arm64
- **macOS**: x64 (Intel), arm64 (Apple Silicon)
- **Windows**: x64, arm64

インストール時に適切なバイナリが自動選択されます。

## 設定ファイル

設定ファイルの場所：

- デフォルト: `~/.aivis-cli.yaml`
- カスタム: `--config` フラグで指定

<details>
<summary>利用可能なパラメータ一覧</summary>

| パラメータ                 | 型      | デフォルト値                    | 説明                                       |
| -------------------------- | ------- | ------------------------------- | ------------------------------------------ |
| `api_key`                  | string  | -                               | Aivis Cloud API キー（必須）               |
| `base_url`                 | string  | `https://api.aivis-project.com` | API のベース URL                           |
| `timeout`                  | string  | `60s`                           | HTTP リクエストのタイムアウト              |
| `default_playback_mode`    | string  | `immediate`                     | デフォルトの音声再生モード                 |
| `default_model_uuid`       | string  | -                               | デフォルト音声モデル UUID                  |
| `default_format`           | string  | `wav`                           | デフォルト音声フォーマット                 |
| `default_volume`           | float64 | `1.0`                           | デフォルト音量（0.0-2.0）                  |
| `default_rate`             | float64 | `1.0`                           | デフォルト再生速度（0.5-2.0）              |
| `default_pitch`            | float64 | `0.0`                           | デフォルトピッチ（-1.0 から 1.0）          |
| `default_wait_for_end`     | bool    | `false`                         | デフォルト再生完了待機                     |
| `use_simplified_tts_tools` | bool    | `false`                         | MCP で簡略化された TTS ツールを使用        |
| `log_level`                | string  | `INFO`                          | ログレベル（DEBUG, INFO, WARN, ERROR）     |
| `log_output`               | string  | `stdout`                        | ログ出力先（stdout, stderr, ファイルパス） |
| `log_format`               | string  | `text`                          | ログ形式（text, json）                     |

### 設定の優先度

設定値は以下の優先順位で適用されます（上位が優先）:

1. **コマンドラインフラグ** - `--api-key`, `--log-level` など
2. **環境変数** - `AIVIS_API_KEY`, `AIVIS_LOG_LEVEL` など  
3. **設定ファイル** - `~/.aivis-cli.yaml` の記載値

```bash
# 例：ログレベルの優先順位
npx @kajidog/aivis-cloud-cli --log-level DEBUG mcp  # 1. フラグ（最優先）
export AIVIS_LOG_LEVEL=INFO                         # 2. 環境変数
# ~/.aivis-cli.yaml: log_level: WARN                # 3. 設定ファイル
```

**環境変数の命名規則**: 設定名の前に `AIVIS_` を付け、大文字に変換します
- `api_key` → `AIVIS_API_KEY`
- `log_level` → `AIVIS_LOG_LEVEL`
- `default_model_uuid` → `AIVIS_DEFAULT_MODEL_UUID`

### ⚠️ MCP サーバー使用時の重要な注意点

#### stdio モード使用時のログ出力

**stdio モード**（デフォルト）では、標準入出力がMCPプロトコル通信に使用されるため、ログ出力が自動的に`stderr`にリダイレクトされます。

```bash
# stdio モード：ログ出力は自動的に stderr に変更されます
npx @kajidog/aivis-cloud-cli mcp
# → log_output が自動的に "stderr" に設定される

# HTTP モード：通常どおり stdout にログ出力
npx @kajidog/aivis-cloud-cli mcp --transport http
# → log_output の設定が適用される
```

これにより、Claude Desktop や他の MCP クライアントとの通信が正常に行われます。

</details>

設定例：

```yaml
api_key: "your-api-key"
base_url: "https://api.aivis-project.com"
timeout: "60s"
default_playback_mode: "immediate"
default_model_uuid: "a59cb814-0083-4369-8542-f51a29e72af7"
default_format: "wav"
default_volume: 1.0
default_rate: 1.0
default_pitch: 0.0
default_wait_for_end: false
use_simplified_tts_tools: false
log_level: "INFO"
log_output: "stdout"
log_format: "text"
```

## 環境変数

`AIVIS_` プレフィックスで設定可能：

- `AIVIS_API_KEY`: API キー
- `AIVIS_BASE_URL`: ベース URL
- `AIVIS_TIMEOUT`: HTTP タイムアウト

## エラーコードと対処法

- **401 Unauthorized**: API キーを確認してください
- **402 Payment Required**: クレジット不足です
- **404 Not Found**: モデル UUID が無効です
- **422 Unprocessable Entity**: パラメータが無効です
- **429 Too Many Requests**: レート制限に達しました

## ライセンス

MIT
