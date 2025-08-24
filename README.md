# Aivis Cloud CLI

Aivis Cloud API を使用して音声合成と音声再生を行うコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## このプロジェクトでできること

- **音声合成（TTS）**: テキストを自然な音声に変換
- **柔軟な音声再生**: 同期・非同期再生、キュー管理による複数音声制御
- **MCP（Model Context Protocol）対応**: Claude などの AI アシスタントが直接音声再生可能
- **豊富な音声モデル**: 人気・最新・作者別など多様な検索とモデル管理
- **設定管理**: API キーや各種設定の保存・管理

## パッケージ構成

このリポジトリは以下のパッケージで構成されています：

### [packages/npm/](./packages/npm/)

**メインパッケージ** - npm で配布される CLI ツール

- npm でのインストール・実行
- 詳細な機能説明と使用例
- MCP サーバー設定方法

### [packages/cli/](./packages/cli/)

**Go 版 CLI** - 開発・ビルド用

- Go での直接ビルド・実行
- クロスプラットフォームビルド
- 開発者向け情報

### [packages/client/](./packages/client/)

**Go クライアントライブラリ** - 他のアプリケーションでの利用

- Go プロジェクトでの組み込み用ライブラリ
- API のプログラマティックな利用
- 基本的な使用例とリファレンス

## クイックスタート

### npm 版（推奨）

```bash
# インストール不要で直接実行
npx @kajidog/aivis-cloud-cli --help

# API キー設定
npx @kajidog/aivis-cloud-cli config set api_key "your-api-key"

# 音声合成
npx @kajidog/aivis-cloud-cli tts play --text "こんにちは世界"
```

### Go 版（開発用）

```bash
cd packages/cli
go build -o aivis-cli
./aivis-cli --help
```

## API キーの取得

**ダッシュボード**: https://hub.aivis-project.com/cloud-api/dashboard

ダッシュボードにアクセスして API キーを取得してください。

## ライセンス

MIT License
