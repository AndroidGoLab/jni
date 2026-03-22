package recorder

// Audio source constants from android.media.MediaRecorder.AudioSource.
const (
	AudioSourceDefault         int32 = 0
	AudioSourceMIC             int32 = 1
	AudioSourceVoiceUplink     int32 = 2
	AudioSourceVoiceDownlink   int32 = 3
	AudioSourceVoiceCall       int32 = 4
	AudioSourceCamcorder       int32 = 5
	AudioSourceVoiceRecognition int32 = 6
	AudioSourceVoiceCommunication int32 = 7
	AudioSourceUnprocessed     int32 = 9
	AudioSourceVoicePerformance int32 = 10
)

// Video source constants from android.media.MediaRecorder.VideoSource.
const (
	VideoSourceDefault int32 = 0
	VideoSourceCamera  int32 = 1
	VideoSourceSurface int32 = 2
)

// Output format constants from android.media.MediaRecorder.OutputFormat.
const (
	OutputFormatDefault      int32 = 0
	OutputFormatThreeGPP     int32 = 1
	OutputFormatMPEG4        int32 = 2
	OutputFormatAMRNB        int32 = 3
	OutputFormatAMRWB        int32 = 4
	OutputFormatAACADTS      int32 = 6
	OutputFormatWebM         int32 = 9
	OutputFormatOGG          int32 = 11
)

// Audio encoder constants from android.media.MediaRecorder.AudioEncoder.
const (
	AudioEncoderDefault    int32 = 0
	AudioEncoderAMRNB      int32 = 1
	AudioEncoderAMRWB      int32 = 2
	AudioEncoderAAC        int32 = 3
	AudioEncoderHEAAC      int32 = 4
	AudioEncoderAACELD     int32 = 5
	AudioEncoderVorbis     int32 = 6
	AudioEncoderOpus       int32 = 7
)

// Video encoder constants from android.media.MediaRecorder.VideoEncoder.
const (
	VideoEncoderDefault int32 = 0
	VideoEncoderH263    int32 = 1
	VideoEncoderH264    int32 = 2
	VideoEncoderMPEG4SP int32 = 3
	VideoEncoderVP8     int32 = 4
	VideoEncoderHEVC    int32 = 5
)
