# Aivis Cloud CLI

Aivis Cloud API を使用した音声合成・音声再生のコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## 使い方

**詳細な機能説明・使用例・MCP設定などは [npm パッケージのREADME](./packages/npm/README.md) をご確認ください。**

### インストール・実行

```bash
# インストール不要で直接実行（推奨）
npx @kajidog/aivis-cloud-cli --help

# API キー設定
npx @kajidog/aivis-cloud-cli config set api_key "your-api-key"

# 音声合成・再生
npx @kajidog/aivis-cloud-cli tts play --text "こんにちは世界"
```

## パッケージ構成

### [packages/npm/](./packages/npm/) - **メインパッケージ**
npm で配布される CLI ツール。**機能詳細・使用例・MCP設定はこちら**

### [packages/cli/](./packages/cli/) - Go版CLI
開発者向け。Go での直接ビルド・実行用

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
