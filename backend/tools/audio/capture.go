package tools

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/gen2brain/malgo"
)

func CaptureAudio() {
	// Initialize malgo context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("Malgo Log: %v\n", message)
	})
	if err != nil {
		log.Fatalf("Failed to initialize malgo context: %v", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// Create a buffer to store recording data
	// Use int16 type for 16-bit depth
	var pcmData []int16

	// Device callback function - called when audio data is available
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		// Convert byte slice to int16 slice
		sampleCount := framecount * uint32(CHANNELS)
		samples := unsafe.Slice((*int16)(unsafe.Pointer(&pInputSamples[0])), sampleCount)
		pcmData = append(pcmData, samples...)
	}

	// Configure capture device
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = CHANNELS
	deviceConfig.SampleRate = SAMPLE_RATE
	deviceConfig.Alsa.NoMMap = 1

	// Callback configuration
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	// Initialize capture device
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		log.Fatalf("Failed to initialize capture device: %v", err)
	}
	defer device.Uninit()

	fmt.Printf("Please start speaking after the countdown... Recording for %d seconds\n", SECONDS_TO_RECORD)
	for i := 3; i > 0; i-- {
		fmt.Printf("%d...\n", i)
		time.Sleep(1 * time.Second)
	}

	// Start recording
	fmt.Println("--> Recording...")
	if err := device.Start(); err != nil {
		log.Fatalf("Failed to start device: %v", err)
	}

	// Record for the specified duration
	time.Sleep(time.Duration(SECONDS_TO_RECORD) * time.Second)

	// Stop recording
	if err := device.Stop(); err != nil {
		log.Fatalf("Failed to stop device: %v", err)
	}
	fmt.Println("--> Recording finished.")

	// Encode raw audio data to WAV format (in memory)
	wavData, err := encodeToWAV(pcmData)
	if err != nil {
		log.Fatalf("Failed to encode WAV: %v", err)
	}

	// TODO: Use wavData for transcription or save to file
	fmt.Printf("Recorded %d samples, WAV size: %d bytes\n", len(pcmData), len(wavData))
	_ = wavData
}
