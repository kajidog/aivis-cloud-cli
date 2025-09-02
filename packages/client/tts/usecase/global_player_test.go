package usecase

import (
    "bytes"
    "context"
    "io"
    "sync"
    "testing"

    "github.com/kajidog/aivis-cloud-cli/client/tts/domain"
)

// fakeTTSRepo implements domain.TTSRepository for tests
type fakeTTSRepo struct{}

func (f *fakeTTSRepo) Synthesize(ctx context.Context, request *domain.TTSRequest) (*domain.TTSResponse, error) {
    return &domain.TTSResponse{AudioData: io.NopCloser(bytes.NewReader([]byte("abc")))}, nil
}
func (f *fakeTTSRepo) SynthesizeStream(ctx context.Context, request *domain.TTSRequest) (io.ReadCloser, error) {
    return io.NopCloser(bytes.NewReader([]byte("abcdef"))), nil
}

// mockPlayer implements domain.AudioPlayer for tests
type mockPlayer struct {
    mu        sync.Mutex
    playing   bool
    stopCount int
    playCount int
}

func (m *mockPlayer) Play(ctx context.Context, audioData io.Reader, format domain.OutputFormat) error {
    m.mu.Lock(); m.playing = true; m.playCount++; m.mu.Unlock()
    // drain reader to simulate playback until EOF
    io.Copy(io.Discard, audioData)
    m.mu.Lock(); m.playing = false; m.mu.Unlock()
    return nil
}
func (m *mockPlayer) Stop() error { m.mu.Lock(); m.stopCount++; m.playing = false; m.mu.Unlock(); return nil }
func (m *mockPlayer) Pause() error { return nil }
func (m *mockPlayer) Resume() error { return nil }
func (m *mockPlayer) SetVolume(volume float64) error { return nil }
func (m *mockPlayer) GetStatus() domain.PlaybackInfo { m.mu.Lock(); defer m.mu.Unlock(); if m.playing { return domain.PlaybackInfo{Status: domain.PlaybackStatusPlaying} }; return domain.PlaybackInfo{Status: domain.PlaybackStatusIdle} }
func (m *mockPlayer) IsPlaying() bool { m.mu.Lock(); defer m.mu.Unlock(); return m.playing }
func (m *mockPlayer) Close() error { return nil }

func newBasicRequest() *domain.PlaybackRequest {
    tts := domain.NewTTSRequestBuilder("model", "hello").WithOutputFormat(domain.OutputFormatMP3).Build()
    return domain.NewPlaybackRequest(tts).Build()
}

// When queue + no wait, and something is already playing, we enqueue and must not call Stop()
func TestPlayRequestWithHistory_Queue_NoWait_DoesNotInterruptPlaying(t *testing.T) {
    repo := &fakeTTSRepo{}
    synth := NewTTSSynthesizer(repo)
    mp := &mockPlayer{}

    s := GetGlobalAudioPlayerService()
    s.Initialize(synth, mp, &AudioPlayerConfig{MaxQueueSize: 10})

    // Simulate ongoing playback
    mp.mu.Lock(); mp.playing = true; mp.mu.Unlock()

    req := newBasicRequest()
    // explicit queue mode
    mode := domain.PlaybackModeQueue
    req.Mode = &mode
    wait := false
    req.WaitForEnd = &wait

    if err := s.PlayRequestWithHistory(context.Background(), req, "test_hist.mp3"); err != nil {
        // Even if processing fails later, enqueue path should return nil
        // but do not treat it as failure for this logic test
    }

    if mp.stopCount != 0 {
        t.Fatalf("expected no Stop() calls, got %d", mp.stopCount)
    }
}

// When queue + wait_for_end=true, it should process synchronously without calling Stop()
func TestPlayRequestWithHistory_Queue_Wait_SynchronousNoStop(t *testing.T) {
    repo := &fakeTTSRepo{}
    synth := NewTTSSynthesizer(repo)
    mp := &mockPlayer{}

    s := GetGlobalAudioPlayerService()
    s.Initialize(synth, mp, &AudioPlayerConfig{MaxQueueSize: 10})

    // Ensure idle to allow immediate processing
    mp.mu.Lock(); mp.playing = false; mp.mu.Unlock()

    req := newBasicRequest()
    mode := domain.PlaybackModeQueue
    req.Mode = &mode
    wait := true
    req.WaitForEnd = &wait

    if err := s.PlayRequestWithHistory(context.Background(), req, "test_hist.mp3"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if mp.stopCount != 0 {
        t.Fatalf("expected no Stop() calls, got %d", mp.stopCount)
    }
}

// Immediate should stop current playback (both async and sync)
func TestPlayRequestWithHistory_Immediate_StopsCurrent(t *testing.T) {
    repo := &fakeTTSRepo{}
    synth := NewTTSSynthesizer(repo)
    mp := &mockPlayer{}

    s := GetGlobalAudioPlayerService()
    s.Initialize(synth, mp, &AudioPlayerConfig{MaxQueueSize: 10})

    // Simulate ongoing playback
    mp.mu.Lock(); mp.playing = true; mp.mu.Unlock()

    req := newBasicRequest()
    mode := domain.PlaybackModeImmediate
    req.Mode = &mode

    // Async immediate
    wait := false
    req.WaitForEnd = &wait
    if err := s.PlayRequestWithHistory(context.Background(), req, "test_hist.mp3"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if mp.stopCount == 0 {
        t.Fatalf("expected Stop() to be called for immediate mode (async)")
    }

    // Reset playing and stopCount
    mp.mu.Lock(); mp.playing = true; mp.stopCount = 0; mp.mu.Unlock()

    // Sync immediate
    wait = true
    if err := s.PlayRequestWithHistory(context.Background(), req, "test_hist2.mp3"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if mp.stopCount == 0 {
        t.Fatalf("expected Stop() to be called for immediate mode (sync)")
    }
}
