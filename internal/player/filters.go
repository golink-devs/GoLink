package player

import (
	"fmt"
	"strings"
)

type Filters struct {
	Volume     *float32          `json:"volume,omitempty"`
	Equalizer  []EqualizerBand   `json:"equalizer,omitempty"`
	Karaoke    *KaraokeFilter    `json:"karaoke,omitempty"`
	Timescale  *TimescaleFilter  `json:"timescale,omitempty"`
	Tremolo    *TremoloFilter    `json:"tremolo,omitempty"`
	Vibrato    *VibratoFilter    `json:"vibrato,omitempty"`
	Rotation   *RotationFilter   `json:"rotation,omitempty"`
	Distortion *DistortionFilter `json:"distortion,omitempty"`
	ChannelMix *ChannelMixFilter `json:"channelMix,omitempty"`
	LowPass    *LowPassFilter    `json:"lowPass,omitempty"`
}

type EqualizerBand struct {
	Band int     `json:"band"` // 0-14
	Gain float32 `json:"gain"` // -0.25 to 1.0
}

type KaraokeFilter struct {
	Level       float32 `json:"level"`
	MonoLevel   float32 `json:"monoLevel"`
	FilterBand  float32 `json:"filterBand"`
	FilterWidth float32 `json:"filterWidth"`
}

type TimescaleFilter struct {
	Speed float32 `json:"speed"` // default 1.0
	Pitch float32 `json:"pitch"` // default 1.0
	Rate  float32 `json:"rate"`  // default 1.0
}

type TremoloFilter struct {
	Frequency float32 `json:"frequency"` // > 0
	Depth     float32 `json:"depth"`     // 0-1
}

type VibratoFilter struct {
	Frequency float32 `json:"frequency"` // 0-14
	Depth     float32 `json:"depth"`     // 0-1
}

type RotationFilter struct {
	RotationHz float32 `json:"rotationHz"` // Hz
}

type DistortionFilter struct {
	SinOffset float32 `json:"sinOffset"`
	SinScale  float32 `json:"sinScale"`
	CosOffset float32 `json:"cosOffset"`
	CosScale  float32 `json:"cosScale"`
	TanOffset float32 `json:"tanOffset"`
	TanScale  float32 `json:"tanScale"`
	Offset    float32 `json:"offset"`
	Scale     float32 `json:"scale"`
}

type ChannelMixFilter struct {
	LeftToLeft   float32 `json:"leftToLeft"`
	LeftToRight  float32 `json:"leftToRight"`
	RightToLeft  float32 `json:"rightToLeft"`
	RightToRight float32 `json:"rightToRight"`
}

type LowPassFilter struct {
	Smoothing float32 `json:"smoothing"`
}

func (f Filters) BuildFilterChain() string {
	var parts []string

	if f.Volume != nil {
		parts = append(parts, fmt.Sprintf("volume=%.2f", *f.Volume))
	}
	if f.Timescale != nil {
		// atempo only accepts 0.5–2.0, chain multiple for wider range if needed
		// For simplicity, we just use it once.
		parts = append(parts, fmt.Sprintf("atempo=%.2f", f.Timescale.Speed))
	}
	if f.LowPass != nil {
		parts = append(parts, fmt.Sprintf("lowpass=f=%.0f", f.LowPass.Smoothing))
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ",")
}
