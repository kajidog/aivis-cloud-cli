# Aivis Cloud CLI

Aivis Cloud API を使用して音声合成と音声再生を行うコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## 開発・ビルド

機能・使用方法については [npm パッケージ README](../npm/README.md) を参照してください。

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

## テスト

```bash
cd packages/cli
go test -v
```
