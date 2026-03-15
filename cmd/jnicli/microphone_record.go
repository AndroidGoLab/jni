package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	pb "github.com/AndroidGoLab/jni/proto/jni_raw"
)

// Android MediaRecorder constants for audio-only recording.
const (
	micAudioSourceMIC       = 1      // MediaRecorder.AudioSource.MIC
	micOutputFormatMPEG4    = 2      // MediaRecorder.OutputFormat.MPEG_4
	micAudioEncoderAAC      = 3      // MediaRecorder.AudioEncoder.AAC
	micAudioEncodingBitRate = 128000 // 128 kbps
	micAudioSamplingRate    = 44100  // 44.1 kHz
)

var microphoneCmd = &cobra.Command{
	Use:   "microphone",
	Short: "Microphone operations (record)",
}

var microphoneRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record audio from the microphone via MediaRecorder (raw JNI)",
	Long: `Records audio using the Android MediaRecorder API,
driven entirely through raw JNI calls over gRPC.
The resulting M4A is written to stdout or a file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, _ := cmd.Flags().GetDuration("duration")
		if duration <= 0 {
			return fmt.Errorf("--duration must be positive")
		}
		output, _ := cmd.Flags().GetString("output")

		// Recording duration + 60s for setup/teardown.
		ctx, cancel := context.WithTimeout(cmd.Context(), duration+60*time.Second)
		defer cancel()

		client := pb.NewJNIServiceClient(grpcConn)
		j := &jniCaller{client: client, ctx: ctx}

		data, err := recordAudio(ctx, j, duration)
		if err != nil {
			return err
		}

		switch {
		case output == "-" || output == "":
			_, err := os.Stdout.Write(data)
			return err
		default:
			if err := os.WriteFile(output, data, 0644); err != nil {
				return fmt.Errorf("writing %s: %w", output, err)
			}
			fmt.Fprintf(os.Stderr, "Saved %d bytes to %s\n", len(data), output)
			return nil
		}
	},
}

func init() {
	microphoneRecordCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
	microphoneRecordCmd.Flags().DurationP("duration", "d", 10*time.Second, "recording duration")
	microphoneCmd.AddCommand(microphoneRecordCmd)
	rootCmd.AddCommand(microphoneCmd)
}

// recordAudio orchestrates audio-only recording via MediaRecorder over raw JNI.
func recordAudio(
	ctx context.Context,
	j *jniCaller,
	duration time.Duration,
) ([]byte, error) {
	appCtx, err := j.getAppContext()
	if err != nil {
		return nil, fmt.Errorf("getting app context: %w", err)
	}

	// Resolve output path on the device.
	outputPath, err := getAudioOutputPath(j, appCtx)
	if err != nil {
		return nil, fmt.Errorf("getting output path: %w", err)
	}

	// Create, configure, and prepare the MediaRecorder.
	recorder, err := setupAudioRecorder(j, appCtx, outputPath)
	if err != nil {
		return nil, fmt.Errorf("setting up MediaRecorder: %w", err)
	}
	defer func() {
		if releaseErr := releaseMediaRecorder(j, recorder); releaseErr != nil {
			fmt.Fprintf(os.Stderr, "warning: releasing MediaRecorder: %v\n", releaseErr)
		}
	}()

	// Start recording.
	mrCls, err := j.findClass("android/media/MediaRecorder")
	if err != nil {
		return nil, fmt.Errorf("finding MediaRecorder class: %w", err)
	}

	startMid, err := j.getMethodID(mrCls, "start", "()V")
	if err != nil {
		return nil, fmt.Errorf("getting MediaRecorder.start method: %w", err)
	}
	if err := j.callVoidMethod(recorder, startMid); err != nil {
		return nil, fmt.Errorf("calling MediaRecorder.start: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Recording audio for %v...\n", duration)
	select {
	case <-time.After(duration):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Stop recording.
	stopMid, err := j.getMethodID(mrCls, "stop", "()V")
	if err != nil {
		return nil, fmt.Errorf("getting MediaRecorder.stop method: %w", err)
	}
	if err := j.callVoidMethod(recorder, stopMid); err != nil {
		return nil, fmt.Errorf("calling MediaRecorder.stop: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Recording stopped.\n")

	// Read the recorded file back via JNI.
	fmt.Fprintf(os.Stderr, "Reading recorded file...\n")
	return readFileViaJNI(j, outputPath)
}

// getAudioOutputPath returns a path in the app's cache directory for the audio recording.
func getAudioOutputPath(j *jniCaller, appCtx int64) (string, error) {
	contextCls, err := j.findClass("android/content/Context")
	if err != nil {
		return "", fmt.Errorf("finding Context class: %w", err)
	}

	getCacheDirMid, err := j.getMethodID(contextCls, "getCacheDir", "()Ljava/io/File;")
	if err != nil {
		return "", fmt.Errorf("getting getCacheDir method: %w", err)
	}

	cacheDir, err := j.callObjectMethod(appCtx, getCacheDirMid)
	if err != nil {
		return "", fmt.Errorf("calling getCacheDir: %w", err)
	}

	fileCls, err := j.findClass("java/io/File")
	if err != nil {
		return "", fmt.Errorf("finding File class: %w", err)
	}

	getAbsolutePathMid, err := j.getMethodID(fileCls, "getAbsolutePath", "()Ljava/lang/String;")
	if err != nil {
		return "", fmt.Errorf("getting getAbsolutePath method: %w", err)
	}

	pathHandle, err := j.callObjectMethod(cacheDir, getAbsolutePathMid)
	if err != nil {
		return "", fmt.Errorf("calling getAbsolutePath: %w", err)
	}

	cachePath, err := j.getStringUTFChars(pathHandle)
	if err != nil {
		return "", fmt.Errorf("reading cache dir path: %w", err)
	}

	return cachePath + "/recording.m4a", nil
}

// setupAudioRecorder creates and configures a MediaRecorder for audio-only recording.
func setupAudioRecorder(
	j *jniCaller,
	appCtx int64,
	outputPath string,
) (int64, error) {
	mrCls, err := j.findClass("android/media/MediaRecorder")
	if err != nil {
		return 0, fmt.Errorf("finding MediaRecorder class: %w", err)
	}

	// MediaRecorder(Context) constructor (API 31+).
	// Fall back to no-arg constructor for older APIs.
	var recorder int64
	ctor, ctorErr := j.getMethodID(mrCls, "<init>", "(Landroid/content/Context;)V")
	if ctorErr == nil {
		recorder, err = j.newObject(mrCls, ctor, objVal(appCtx))
	}
	if ctorErr != nil || err != nil {
		ctor, err = j.getMethodID(mrCls, "<init>", "()V")
		if err != nil {
			return 0, fmt.Errorf("getting MediaRecorder constructor: %w", err)
		}
		recorder, err = j.newObject(mrCls, ctor)
		if err != nil {
			return 0, fmt.Errorf("creating MediaRecorder: %w", err)
		}
	}

	outputPathStr, err := j.newString(outputPath)
	if err != nil {
		return 0, fmt.Errorf("creating output path string: %w", err)
	}

	type methodCall struct {
		name string
		sig  string
		args []*pb.JValue
	}

	calls := []methodCall{
		{"setAudioSource", "(I)V", []*pb.JValue{intVal(micAudioSourceMIC)}},
		{"setOutputFormat", "(I)V", []*pb.JValue{intVal(micOutputFormatMPEG4)}},
		{"setAudioEncoder", "(I)V", []*pb.JValue{intVal(micAudioEncoderAAC)}},
		{"setAudioEncodingBitRate", "(I)V", []*pb.JValue{intVal(micAudioEncodingBitRate)}},
		{"setAudioSamplingRate", "(I)V", []*pb.JValue{intVal(micAudioSamplingRate)}},
		{"setOutputFile", "(Ljava/lang/String;)V", []*pb.JValue{objVal(outputPathStr)}},
		{"prepare", "()V", nil},
	}

	for _, c := range calls {
		mid, err := j.getMethodID(mrCls, c.name, c.sig)
		if err != nil {
			return 0, fmt.Errorf("getting MediaRecorder.%s method: %w", c.name, err)
		}
		if err := j.callVoidMethod(recorder, mid, c.args...); err != nil {
			return 0, fmt.Errorf("calling MediaRecorder.%s: %w", c.name, err)
		}
	}

	return recorder, nil
}
