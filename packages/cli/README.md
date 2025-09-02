# Aivis Cloud CLI

Aivis Cloud API を使用して音声合成と音声再生を行うコマンドラインツールです。

**公式サイト**: https://aivis-project.com/

## 開発・ビルド

機能・使用方法については [npm パッケージ README](../npm/README.md) を参照してください。

補足（CLI 実行時の挙動の要点）
- `tts play` は既定で履歴保存（Resume対応）。無効化は `--save-history=false`
- 再生モード: `immediate` は既存再生を停止、`queue` は順次、`no_queue` は並列
- Windows で ffplay がある場合はストリーミング再生。ない場合は生成完了後に再生（途中停止回避）
- MCP stdio 実行時は子プロセス stdout を抑止してプロトコル汚染を防止

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
