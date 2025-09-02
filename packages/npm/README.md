# Aivis Cloud CLI

Aivis Cloud API を使用して音声合成と音声再生を行うコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## このパッケージでできること

- **音声合成（TTS）**: テキストを自然な音声に変換
- **履歴管理・Resume機能**: 合成履歴の自動保存と連番IDでの再生機能
- **柔軟な音声再生**: 同期・非同期再生、キュー管理による複数音声制御
- **MCP（Model Context Protocol）対応**: Claude などの AI アシスタントが直接音声再生可能
- **音声モデル検索**: 人気・最新・作者別など多様な検索とモデル管理
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
# 基本的な音声合成（デフォルトモデルを使用、出力ファイル名自動生成）
npx @kajidog/aivis-cloud-cli tts synthesize --text "こんにちは世界"
# → Output: tts_20240101_120000.wav（タイムスタンプ付き）
# → History saved with ID: 1

# 位置引数を使用した音声ファイル保存
npx @kajidog/aivis-cloud-cli tts synthesize "こんにちは" "output.wav"

# 出力ファイル名を明示的に指定
npx @kajidog/aivis-cloud-cli tts synthesize --text "こんにちは世界" --output "output.wav"

# 特定のモデルを指定
npx @kajidog/aivis-cloud-cli tts synthesize --text "こんにちは世界" --output "output.wav" --model-uuid "model-id"

# SSML マークアップを使用
npx @kajidog/aivis-cloud-cli tts synthesize --text '<speak>こんにちは<break time="1s"/>世界</speak>' --output "output.wav" --ssml

# 高度なTTSパラメータを使用
npx @kajidog/aivis-cloud-cli tts synthesize --text "感情豊かに話します" --output "output.wav" --emotional-intensity 1.5 --tempo-dynamics 1.2
```

</details>

<details>
<summary>音声の即時再生</summary>

```bash
# テキストを音声に変換してすぐに再生（デフォルトモデルを使用、履歴は既定で保存）
npx @kajidog/aivis-cloud-cli tts play --text "こんにちは世界"

# 履歴保存を無効化したい場合
npx @kajidog/aivis-cloud-cli tts play --text "こんにちは世界" --save-history=false

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
<summary>ストリーミング合成（標準出力）</summary>

```bash
# ストリーミング合成（リアルタイム出力、標準出力に音声データを出力）
npx @kajidog/aivis-cloud-cli tts stream --text "こんにちは世界" > output.wav
```

</details>

### TTS履歴管理機能

<details>
<summary>履歴の自動保存とResume機能</summary>

TTS合成実行時に履歴が自動保存され、連番IDで管理されます。

```bash
# 音声合成（履歴自動保存）
npx @kajidog/aivis-cloud-cli tts synthesize "こんにちは世界"
# → History saved with ID: 1

# 履歴一覧表示
npx @kajidog/aivis-cloud-cli tts history list
# ID  Text          Model     Format  Size    Created
# 1   こんにちは世界  a59cb...  wav     45KB    01/01 12:00

# 履歴詳細表示（リクエスト内容、ファイル情報など）
npx @kajidog/aivis-cloud-cli tts history show 1
# Text: こんにちは世界
# Model UUID: a59cb814-0083-4369-8542-f51a29e72af7
# Created: 2025-01-01 12:00:00
# File Path: tts_20250101_120000.wav
# File Format: wav
# File Size: 45.2 KB
# Credits Used: 0.0050
# 
# Request Details:
# ----------------
# Speaking Rate: 1.20
# Pitch: 0.10
# Volume: 0.80
# Output Format: mp3
# Audio Channels: stereo
# Leading Silence: 0.50 seconds
# Trailing Silence: 0.30 seconds
# Sampling Rate: 44100 Hz
# Bitrate: 128 kbps
# SSML: Enabled

# 履歴から再生（Resume機能）
npx @kajidog/aivis-cloud-cli tts history play 1

# 履歴統計
npx @kajidog/aivis-cloud-cli tts history stats

# 履歴削除
npx @kajidog/aivis-cloud-cli tts history delete 1 --force

# 古い履歴のクリーンアップ（30日以上前）
npx @kajidog/aivis-cloud-cli tts history clean --older-than 30

# 全履歴削除
npx @kajidog/aivis-cloud-cli tts history clean --all --force
```

**履歴設定:**
```bash
# 履歴機能を無効化
npx @kajidog/aivis-cloud-cli config set history_enabled false

# 最大保存件数を変更（デフォルト: 100）
npx @kajidog/aivis-cloud-cli config set history_max_count 50

# 履歴保存パスを変更
npx @kajidog/aivis-cloud-cli config set history_store_path "/custom/path"
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
  - **ストリーミング音声合成**: 音声生成をリアルタイムで実行、履歴ファイルに並行保存
  - **プログレッシブ再生**: MP3形式では音声生成と同時に再生開始、その他形式は合成完了後再生
  - パラメータ: `text` (必須), `model_uuid`, `format`, `volume`, `rate`, `pitch`, `playback_mode`, `wait_for_end`
  - 音声フォーマット: `wav`, `mp3`, `flac`, `aac`, `opus`
  - 再生モード: `immediate` (即座再生), `queue` (キュー追加, **デフォルト**), `no_queue` (同時再生)

- **play_text**: デフォルト設定でテキストを音声再生（簡易版）
  - パラメータ: `text` (必須), `playback_mode`, `wait_for_end`
  - 注意: `default_model_uuid` と `use_simplified_tts_tools: true` が設定されている場合のみ利用可能

**TTS履歴管理（Resume機能）:**

- **list_tts_history**: TTS履歴一覧表示・検索
  - パラメータ: `limit`, `offset`, `model_uuid`, `text_contains`, `sort_by`, `sort_order`
  - 連番IDで管理された履歴レコードをページネーション・フィルタリング表示

- **get_tts_history**: 特定履歴の詳細情報取得
  - パラメータ: `id` (必須)
  - テキスト、モデル、ファイル情報、使用クレジット、リクエスト詳細を表示

- **play_tts_history**: **履歴から音声再生（Resume機能）**
  - パラメータ: `id` (必須), `volume`, `playback_mode`, `wait_for_end`
  - **メインのResume機能**: 過去の音声合成をIDで即座再生

- **delete_tts_history**: 特定履歴削除
  - パラメータ: `id` (必須)
  - 履歴レコードと関連音声ファイルを削除

- **get_tts_history_stats**: 履歴統計情報取得
  - パラメータ: なし
  - 総レコード数、ストレージ使用量、使用クレジットの統計を表示

**設定管理関連:**

- **get_mcp_settings**: 現在のMCP設定を取得
  - パラメータ: なし
  - 戻り値: 現在の設定値（APIキーは除外）
  - セキュリティのため、API キーとシステム設定（ログ設定、簡易TTS設定）は表示されません

- **update_mcp_settings**: MCP設定を安全に更新
  - **基本パラメータ**: `base_url`, `default_model_uuid`, `default_playback_mode`, `default_volume`, `default_rate`, `default_pitch`, `default_format`
  - **高度なTTSパラメータ**: `default_ssml`, `default_emotional_intensity`, `default_tempo_dynamics`, `default_leading_silence`, `default_trailing_silence`, `default_channels`
  - **制限**: APIキー、ログ設定、`use_simplified_tts_tools` は変更不可
  - **設定値のバリデーション機能付き**（例：音量は0.0-2.0の範囲、無音時間は0.0-10.0秒の範囲）

### 🎯 **推奨設定**

**AIアシスタント用途**: 全ての音声を順序通り再生
```javascript
{
  "playback_mode": "queue",        // デフォルト - 全音声が順番に再生
  "wait_for_end": false           // MCPがブロックされずスムーズ
}
```

**リアルタイム会話**: 最新の音声を優先
```javascript
{
  "playback_mode": "immediate",    // 前の音声を停止して即座再生
  "wait_for_end": false
}
```

**並行効果音**: 複数音声の同時再生
```javascript
{
  "playback_mode": "no_queue",     // キュー無視で同時再生
  "wait_for_end": false
}
```

**使用例:**
```javascript
// 現在の設定を確認
get_mcp_settings({})

// 基本的な音声合成（最小限のパラメータ）
synthesize_speech({
  "text": "こんにちは世界"
})

// SSMLを使った高度な音声合成
synthesize_speech({
  "text": "<speak><prosody rate='slow'>ゆっくりと</prosody><break time='1s'/>話します</speak>",
  "ssml": true,
  "emotional_intensity": 1.5,
  "tempo_dynamics": 0.8,
  "leading_silence": 0.2,
  "trailing_silence": 0.5,
  "channels": "stereo",
  "format": "mp3"
})

// TTS履歴の管理・Resume機能
list_tts_history({"limit": 10, "sort_by": "created_at"})  // 最新10件を表示

get_tts_history({"id": 3})  // ID=3の履歴詳細を取得

play_tts_history({"id": 3, "volume": 0.8})  // ID=3を音量0.8で再生（Resume）

delete_tts_history({"id": 1})  // ID=1の履歴を削除

get_tts_history_stats({})  // 履歴統計を表示

// 設定を更新（高度なTTSパラメータを含む）
update_mcp_settings({
  "default_volume": 0.8,
  "default_playback_mode": "queue",
  "default_format": "mp3",
  "default_ssml": true,
  "default_emotional_intensity": 1.2,
  "default_tempo_dynamics": 1.1,
  "default_leading_silence": 0.1,
  "default_trailing_silence": 0.3,
  "default_channels": "stereo"
})
```

</details>

## 再生モードと動作の詳細

- **immediate**: 現在の再生を停止し、即座に新規音声を再生（キューもクリア）
- **queue**: 現在の再生を維持し、キューに追加して順次再生（`wait_for_end=true` で完了待機）
- **no_queue**: キューを使わず独立プレイヤーで並列再生（`wait_for_end=true` で同期待機）

補足:
- MCP/stdio 実行時は子プロセスの標準出力を抑止し、標準エラーにログを出力します（プロトコル保護）
- `AIVIS_KEEP_PLAYBACK_FILES=1` で再生用の一時ファイルを削除せず残せます（デバッグ用途）

## ストリーミング再生とプログレッシブ再生のポリシー

- ffplay が利用可能な環境では、標準入力（stdin）ストリーミング再生を優先します
  - 例: `ffplay -loglevel error -nodisp -autoexit -volume <0-100> -i -`
  - 履歴保存が必要な場合は tee で並行保存（単一合成で再生と保存を同時実行）
- ffplay がない Windows では、成長中ファイルの先行再生（プログレッシブ）は無効化し、生成完了後に再生します（途中停止回避のため）
- 低遅延での即時出音を重視する場合は、フォーマットに `mp3` または `opus` を推奨します

## 履歴保存の挙動

- `tts synthesize` は常に履歴を保存します（IDが付与されます）
- `tts play` は既定で履歴保存します（`--save-history=false` で無効化可能）
- MCP の `synthesize_speech` も「再生と同時保存」を単一合成で行い、原則 `wait_for_end=false` でも ID が返ります（内部で短時間ファイル生成を待機）

## FFplay の導入（任意・推奨）

ffplay は FFmpeg に同梱される小型プレイヤーで、標準入力からの再生に対応します。導入済みの環境では、低遅延で安定したストリーミング再生を自動的に使用します。

- 推奨: Windows では導入を推奨（未導入時は生成完了後の再生にフォールバック）
- 必須事項: `ffplay` に PATH が通っている必要があります。導入後はターミナル（やアプリ）を再起動して PATH を反映してください。

<details>
<summary>FFplay の導入手順と PATH 反映</summary>

インストール例:

- Windows（いずれか）
  - Winget: `winget install --id=Gyan.FFmpeg -e`
  - Chocolatey: `choco install ffmpeg`
  - Scoop: `scoop install ffmpeg`
  - 公式ビルド（例）: https://www.gyan.dev/ffmpeg/builds/ または https://github.com/BtbN/FFmpeg-Builds から zip を取得し、`bin` フォルダを PATH に追加

- macOS
  - Homebrew: `brew install ffmpeg`

- Linux
  - Debian/Ubuntu: `sudo apt-get update && sudo apt-get install -y ffmpeg`
  - Fedora: `sudo dnf install -y ffmpeg`
  - Arch: `sudo pacman -S ffmpeg`

PATH の反映:

- Windows: 環境変数に `...\ffmpeg\bin` を追加後、PowerShell/端末・エディタ（Claude/VS Code 等）を再起動。
  - 反映確認: `powershell -c "$env:Path"` に ffmpeg のパスが含まれること
- macOS/Linux: 通常は自動反映。必要に応じて `echo $PATH` で確認し、シェルを再起動。
- MCP クライアント（Claude Desktop/Code）: アプリ側のプロセス再起動で PATH を再読込します。

動作確認:

```bash
ffplay -version
```

バージョン情報が表示されれば導入完了です。CLI/MCP は自動的に ffplay を検出して標準入力ストリーミング再生を使用します。

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
<summary>利用可能なパラメータ一覧（クリックで展開）</summary>

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
| `default_ssml`             | bool    | `false`                         | デフォルトSSML有効化                       |
| `default_emotional_intensity` | float64 | `0.0`                        | デフォルト感情強度（0.0-2.0）              |
| `default_tempo_dynamics`   | float64 | `0.0`                           | デフォルトテンポダイナミクス（0.0-2.0）    |
| `default_leading_silence`  | float64 | `0.0`                           | デフォルト開始無音時間（0.0-10.0秒）       |
| `default_trailing_silence` | float64 | `0.0`                           | デフォルト終了無音時間（0.0-10.0秒）       |
| `default_channels`         | string  | `stereo`                        | デフォルトチャンネル設定（mono/stereo）    |
| `default_wait_for_end`     | bool    | `false`                         | デフォルト再生完了待機                     |
| `use_simplified_tts_tools` | bool    | `false`                         | MCP で簡略化された TTS ツールを使用        |
| `history_enabled`          | bool    | `true`                          | TTS履歴管理機能の有効/無効                 |
| `history_max_count`        | int     | `100`                           | 履歴最大保存件数（自動削除の閾値）         |
| `history_store_path`       | string  | `~/.aivis-cli/history/`         | 履歴ファイル保存ディレクトリ               |
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

#### 設定例

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
default_ssml: false
default_emotional_intensity: 0.0
default_tempo_dynamics: 0.0
default_leading_silence: 0.0
default_trailing_silence: 0.0
default_channels: "stereo"
default_wait_for_end: false
use_simplified_tts_tools: false
history_enabled: true
history_max_count: 100
history_store_path: "~/.aivis-cli/history/"
log_level: "INFO"
log_output: "stdout"
log_format: "text"
```

## 環境変数

<details>
<summary>`AIVIS_` プレフィックスの環境変数一覧（クリックで展開）</summary>

- `AIVIS_API_KEY`: API キー
- `AIVIS_BASE_URL`: ベース URL
- `AIVIS_TIMEOUT`: HTTP タイムアウト

</details>

## APIエラーコード（参考）

<details>
<summary>主なAPIエラー（クリックで展開）</summary>

以下は Aivis Cloud API 側から返る一般的なエラーです。CLI/MCP はこれらを適切に伝播します。

- 401 Unauthorized: API キーを確認してください
- 402 Payment Required: クレジット不足です
- 404 Not Found: モデル UUID が無効です
- 422 Unprocessable Entity: パラメータが無効です
- 429 Too Many Requests: レート制限に達しました

</details>

## ライセンス

MIT
