package audio

import (
	"context"
	"io"
	"os/exec"
)

type FFmpegProcess struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	cancel context.CancelFunc
}

func NewFFmpegProcess(ctx context.Context, url string, filterChain string) (*FFmpegProcess, error) {
	ctx, cancel := context.WithCancel(ctx)

	args := []string{
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", url,
		"-f", "s16le",  // signed 16-bit little endian PCM
		"-ar", "48000", // 48kHz sample rate (Discord requirement)
		"-ac", "2",     // stereo (Discord requirement)
	}

	if filterChain != "" {
		args = append(args, "-af", filterChain)
	}

	args = append(args, "-loglevel", "quiet", "pipe:1")

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	return &FFmpegProcess{
		cmd:    cmd,
		stdout: stdout,
		cancel: cancel,
	}, nil
}

func (f *FFmpegProcess) Read(p []byte) (int, error) {
	return f.stdout.Read(p)
}

func (f *FFmpegProcess) Stop() {
	f.cancel()
	f.cmd.Wait()
}
