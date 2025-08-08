# AivisCloud Go Client

AivisCloud APIのGolangクライアントライブラリです。音声合成とモデル検索機能を提供します。

## インストール

```bash
go get github.com/kajidog/aiviscloud-mcp/client
```

## 使用方法

### 基本的な使い方

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/kajidog/aiviscloud-mcp/client"
    "github.com/kajidog/aiviscloud-mcp/client/tts/domain"
)

func main() {
    // クライアントを初期化
    client, err := client.New("your-api-key")
    if err != nil {
        panic(err)
    }

    // 音声合成を実行
    request := client.NewTTSRequest("model-uuid", "こんにちは、世界！").
        WithSSML(true).
        WithOutputFormat(domain.OutputFormatMP3).
        Build()

    response, err := client.Synthesize(context.Background(), request)
    if err != nil {
        panic(err)
    }
    defer response.AudioData.Close()

    // ファイルに保存
    file, err := os.Create("output.mp3")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    err = client.SynthesizeToFile(context.Background(), request, file)
    if err != nil {
        panic(err)
    }

    fmt.Println("音声合成完了！")
}
```

### 音声合成

#### 基本的な音声合成

```go
request := client.NewTTSRequest("model-uuid", "読み上げテキスト").
    Build()

response, err := client.Synthesize(ctx, request)
```

#### 詳細パラメータ指定

```go
request := client.NewTTSRequest("model-uuid", "読み上げテキスト").
    WithSpeaker("speaker-uuid").
    WithStyleName("Happy").
    WithSSML(true).
    WithOutputFormat(domain.OutputFormatMP3).
    WithSpeakingRate(1.2).
    WithVolume(0.8).
    WithEmotionalIntensity(1.5).
    Build()

response, err := client.Synthesize(ctx, request)
```

#### ストリーミング音声合成

```go
type StreamHandler struct{}

func (h *StreamHandler) OnChunk(chunk *domain.TTSStreamChunk) error {
    // 音声チャンクを処理
    fmt.Printf("Received %d bytes\n", len(chunk.Data))
    return nil
}

func (h *StreamHandler) OnComplete() error {
    fmt.Println("Streaming complete")
    return nil
}

func (h *StreamHandler) OnError(err error) {
    fmt.Printf("Stream error: %v\n", err)
}

handler := &StreamHandler{}
err := client.SynthesizeStream(ctx, request, handler)
```

### モデル検索

#### 基本的な検索

```go
// 公開モデルを検索
models, err := client.SearchPublicModels(ctx, "音声合成")

// 作者で検索
models, err := client.SearchModelsByAuthor(ctx, "author-name")

// タグで検索
models, err := client.SearchModelsByTags(ctx, "tag1", "tag2")
```

#### 詳細な検索

```go
request := client.NewModelSearchRequest().
    WithQuery("音声合成").
    WithTags("japanese", "female").
    WithLanguage("ja").
    WithPublicOnly().
    WithPageSize(20).
    SortByDownloadCount().
    Descending().
    Build()

response, err := client.SearchModels(ctx, request)
```

#### 人気・最新モデル取得

```go
// 人気モデル
popular, err := client.GetPopularModels(ctx, 10)

// 最新モデル  
recent, err := client.GetRecentModels(ctx, 10)

// 高評価モデル
topRated, err := client.GetTopRatedModels(ctx, 10)
```

#### 特定モデルの詳細取得

```go
// モデル詳細
model, err := client.GetModel(ctx, "model-uuid")

// モデルの話者一覧
speakers, err := client.GetModelSpeakers(ctx, "model-uuid")
```

### 設定のカスタマイズ

```go
cfg := config.NewConfig("your-api-key").
    WithBaseURL("https://api.aivis-project.com").
    WithTimeout(30 * time.Second).
    WithUserAgent("my-app/1.0")

client, err := client.NewWithConfig(cfg)
```

### エラーハンドリング

```go
response, err := client.Synthesize(ctx, request)
if err != nil {
    if apiErr, ok := errors.IsAPIError(err); ok {
        switch apiErr.StatusCode {
        case 401:
            fmt.Println("APIキーが無効です")
        case 402:
            fmt.Println("クレジット残高が不足しています")
        case 404:
            fmt.Println("指定されたモデルが見つかりません")
        case 429:
            fmt.Println("レート制限に達しました")
        default:
            fmt.Printf("APIエラー: %v\n", apiErr)
        }
    } else {
        fmt.Printf("その他のエラー: %v\n", err)
    }
    return
}
```

## 機能

### 音声合成API
- 基本的な音声合成
- ストリーミング音声合成
- 多様な音声パラメータ対応（話速、ピッチ、音量など）
- 複数の出力形式サポート（WAV、FLAC、MP3、AAC、Opus）
- SSML対応
- ユーザー辞書対応

### モデル検索API
- キーワード検索
- フィルタリング（作者、タグ、言語など）
- ソート機能
- ページネーション
- モデル詳細取得
- 話者情報取得

### その他
- 設定可能なHTTPクライアント
- 統一されたエラーハンドリング
- レスポンスヘッダーから課金情報取得
- ビルダーパターンによる簡単なリクエスト構築

## ライセンス

MIT License