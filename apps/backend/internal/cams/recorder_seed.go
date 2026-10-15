package cams

import (
	"log"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// copySeed writes the buffered past the recording starts from into its first
// segments, before the live capture goes on from there.
func (r *Recorder) copySeed(recording *capture) {
	if recording.seed == nil || r.buffer == nil {
		return
	}
	r.copySeedTrack(recording, recording.video, recording.seed.Video)
	if recording.seed.Audio == nil {
		return
	}
	recording.mutex.Lock()
	recording.audio = newTrack("aseg", "ainit", "audio.m3u8")
	audio := recording.audio
	recording.mutex.Unlock()
	r.copySeedTrack(recording, audio, recording.seed.Audio)
}

// copySeedTrack copies the segments of a buffered window into a track. A
// segment that cannot be read any more (the buffer trimmed it meanwhile) leaves
// a break instead.
func (r *Recorder) copySeedTrack(recording *capture, stored *track, window *livebuffer.Window) {
	roomID := recording.camera.ID
	for _, buffered := range window.Segments {
		if recording.context.Err() != nil {
			return
		}
		initName, err := r.copySeedInit(recording, stored, roomID, buffered.InitName)
		if err != nil {
			recording.markBreak(stored)
			continue
		}
		data, err := r.buffer.Read(recording.context, roomID, buffered.Name)
		if err != nil {
			recording.markBreak(stored)
			continue
		}
		segment := playlistSegment{Identity: buffered.Identity, Duration: buffered.Duration, Discontinuity: buffered.Discontinuity}
		if err := r.appendSegment(recording, stored, segment, data, initName); err != nil {
			log.Printf("cams: record %s: seed segment %s: %v", recording.camera.Username, buffered.Identity, err)
			recording.markBreak(stored)
			continue
		}
		recording.hasSeen(stored, buffered.Identity)
	}
	recording.mutex.Lock()
	stored.continuesAfter = lastIdentity(window)
	recording.mutex.Unlock()
}

// copySeedInit copies a buffered init section once and returns its name in the recording.
func (r *Recorder) copySeedInit(recording *capture, stored *track, roomID, bufferedName string) (string, error) {
	if bufferedName == "" {
		return "", nil
	}
	if name, known := recording.initName(stored, bufferedName); known {
		return name, nil
	}
	data, err := r.buffer.Read(recording.context, roomID, bufferedName)
	if err != nil {
		return "", err
	}
	return r.saveInitSection(recording, stored, bufferedName, data)
}

// lastIdentity is the identity of the window's newest segment.
func lastIdentity(window *livebuffer.Window) string {
	if len(window.Segments) == 0 {
		return ""
	}
	return window.Segments[len(window.Segments)-1].Identity
}
