# AivisCloud Go Client

Aivis Cloud API の Go クライアントライブラリです。音声合成と音声再生、モデル管理機能を提供します。

**公式サイト**: https://aivis-project.com/

詳細な機能説明・使用方法については [npm パッケージ README](../npm/README.md) を参照してください。

## 開発・ビルド

### 前提条件

- Go 1.21 以上

### インストール

```bash
go get github.com/kajidog/aivis-cloud-cli/client
```

### ビルド（開発用）

```bash
cd packages/client
go mod tidy
go test ./...           # テスト実行
go test -v              # 詳細なテスト実行
go test -cover          # カバレッジ付きテスト実行
go build -v ./...
```

## API使用例

```go
package main

import (
    "context"
    "github.com/kajidog/aivis-cloud-cli/client"
)

func main() {
    client, err := client.New("your-api-key")
    if err != nil {
        panic(err)
    }

    // 音声合成
    request := client.NewTTSRequest("model-uuid", "Hello").Build()
    response, err := client.SynthesizeToFile(context.Background(), request, "output.wav")
    
    // モデル検索
    models, err := client.SearchPublicModels(context.Background(), "Japanese")
}
```

## テスト

このライブラリには包括的なテストスイートが含まれています：

### テスト実行

```bash
# 全テスト実行
go test -v

# カバレッジ付き実行
go test -cover

# 特定のテスト実行
go test -v -run TestSearchPublicModels
```

### テストの特徴

- **Mock HTTP Server**: 実APIに依存せずテスト実行
- **Table-driven Tests**: 複数のシナリオを効率的にテスト
- **Error Handling**: 4xx/5xx エラーレスポンステスト
- **Builder Pattern**: TTS リクエストビルダーのテスト

### テスト例

```go
func TestSearchPublicModels(t *testing.T) {
    // Mock server setup
    handler := func(w http.ResponseWriter, r *http.Request) {
        response := `{"models": [{"uuid": "test-uuid", "name": "test-model"}], "total": 1}`
        w.Write([]byte(response))
    }
    
    client, teardown := setupTestClient(t, handler)
    defer teardown()
    
    resp, err := client.SearchPublicModels(context.Background(), "test")
    // テスト assertions...
}
```

## 詳細情報

- **GoDoc**: https://pkg.go.dev/github.com/kajidog/aivis-cloud-cli/client
- **サンプルコード**: [example/](./example/) ディレクトリ
- **詳細な使用方法**: [npm パッケージ README](../npm/README.md)
