# Aivis Cloud CLI

[![Test](https://github.com/kajidog/aivis-cloud-cli/actions/workflows/test.yml/badge.svg)](https://github.com/kajidog/aivis-cloud-cli/actions/workflows/test.yml) [![NPM Version](https://img.shields.io/npm/v/@kajidog/aivis-cloud-cli.svg)](https://www.npmjs.com/package/@kajidog/aivis-cloud-cli) [![License](https://img.shields.io/github/license/kajidog/aivis-cloud-cli.svg)](https://github.com/kajidog/aivis-cloud-cli/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/kajidog/aivis-cloud-cli/packages/client)](https://goreportcard.com/report/github.com/kajidog/aivis-cloud-cli/packages/client) [![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)](https://go.dev/dl/) [![GitHub release](https://img.shields.io/github/v/release/kajidog/aivis-cloud-cli)](https://github.com/kajidog/aivis-cloud-cli/releases/latest)

Aivis Cloud API を使用した音声合成・音声再生のコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## 主な機能

- **音声合成 (TTS)** - テキストから高品質音声ファイルを生成
- **音声再生** - 合成した音声をその場で再生（Windows/macOS/Linux対応）  
- **履歴管理** - TTS合成履歴の自動保存・管理・再生（Resume機能）
- **モデル検索** - 利用可能な音声モデルの検索・取得
- **MCP対応** - Claude CodeなどAIアシスタントからのストリーミング音声合成（stdio/http）
- **設定管理** - APIキー、デフォルト値、履歴設定の管理

## 使い方

**詳細な機能説明・使用例・MCP設定などは [npm パッケージのREADME](./packages/npm/README.md) をご確認ください。**

### インストール・セットアップ

```bash
# インストール不要で直接実行（推奨）
npx @kajidog/aivis-cloud-cli --help

# API キー設定
npx @kajidog/aivis-cloud-cli config set api_key "your-api-key"
```

## パッケージ構成

### [packages/npm/](./packages/npm/) - **メインパッケージ**
npm で配布される CLI ツール。**機能詳細・使用例・MCP設定はこちら**

使用例:
```bash
# 例1: テキストから音声ファイルを生成（履歴自動保存）
npx @kajidog/aivis-cloud-cli tts synthesize --text "こんにちは世界" --output "output.wav"

# 例2: TTS履歴の管理・再生
npx @kajidog/aivis-cloud-cli tts history list    # 履歴一覧表示
npx @kajidog/aivis-cloud-cli tts history play 1  # ID=1の履歴を再生

# 例3: Claude CodeにMCPを登録（AIアシスタントがストリーミング音声合成・即座再生可能）
claude mcp add aivis npx @kajidog/aivis-cloud-cli mcp
```

### [packages/cli/](./packages/cli/) - Go版CLI
開発者向け。Go での直接ビルド・実行用

**ビルド済みのバイナリ (Windows, macOS, Linux) は [GitHub Releases](https://github.com/kajidog/aivis-cloud-cli/releases) からダウンロードできます。**

### [packages/client/](./packages/client/) - Goライブラリ
他のアプリケーション組み込み用のGoクライアントライブラリ

## 開発者向け

**詳細なテスト実行・ビルド手順・開発ワークフローについては各パッケージのREADMEを参照してください：**

- **[packages/client/README.md](./packages/client/README.md)** - Go ライブラリのテスト・開発
- **[packages/cli/README.md](./packages/cli/README.md)** - CLI のビルド・開発  
- **CLAUDE.md** - プロジェクト全体のアーキテクチャ・設計思想

## API キー取得

[**ダッシュボード**](https://hub.aivis-project.com/cloud-api/dashboard) でAPI キーを取得してください。

## ライセンス

MIT License
