package handlers

// AudioHandler handles audio file management HTTP requests.
type AudioHandler struct {
	AudioFiles SystemAudioFileRepository
}

// NewAudioHandler creates a new AudioHandler.
func NewAudioHandler(audioFiles SystemAudioFileRepository) *AudioHandler {
	return &AudioHandler{AudioFiles: audioFiles}
}
