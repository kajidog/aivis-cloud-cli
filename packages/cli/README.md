# Aivis Cloud CLI

Aivis Cloud API を使用して音声合成と音声再生を行うコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## 機能・使用方法

機能説明と使用例については、[npm パッケージの README](../npm/README.md) を参照してください。

コマンド例の `npx @kajidog/aivis-cloud-cli` を `./aivis-cli` に読み替えて使用してください。

## 開発・ビルド

### 前提条件

- Go 1.21 以上

### ビルド

```bash
cd packages/cli
go mod tidy
go build -o aivis-cli
```

### クロスプラットフォームビルド

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o aivis-cli-darwin-amd64

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o aivis-cli-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o aivis-cli-windows-amd64.exe

# Linux
GOOS=linux GOARCH=amd64 go build -o aivis-cli-linux-amd64
```

## 事前準備

Aivis Cloud API キーが必要です。

### API キーの取得

**ダッシュボード**: https://hub.aivis-project.com/cloud-api/dashboard

ダッシュボードにアクセスして API キーを取得してください。

### API キーの設定

```bash
# 設定ファイルの初期化
./aivis-cli config init

# API キーの設定
./aivis-cli config set api_key YOUR_API_KEY
```

### 設定オプション

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
