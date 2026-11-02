package cams

import (
	"log"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// copySeed writes the buffered past the recording starts from into its first
// segments, before the live capture goes on from there.
func (s *Service) copySeed(recording *capture) {
	if recording.seed == nil || s.services.Buffer == nil {
		return
	}
	s.copySeedTrack(recording, recording.video, recording.seed.Video)
	if recording.seed.Audio == nil {
		return
	}
	s.copySeedTrack(recording, recording.audioTrack(), recording.seed.Audio)
}

// copySeedTrack copies the segments of a buffered window into a track. A
// segment that cannot be read any more (the buffer trimmed it meanwhile) leaves
// a break instead.
func (s *Service) copySeedTrack(recording *capture, stored *track, window *livebuffer.Window) {
	for _, buffered := range window.Segments {
		if recording.context.Err() != nil {
			return
		}
		if err := s.copySeedSegment(recording, stored, buffered); err != nil {
			log.Printf("cams: record %s: seed segment %s: %v", recording.camModel.Username, buffered.Identity, err)
			recording.markBreak(stored)
			continue
		}
		recording.hasSeen(stored, buffered.Identity)
	}
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	stored.continuesAfter = lastIdentity(window)
}

// copySeedSegment copies one buffered segment, and its init section, into the track.
func (s *Service) copySeedSegment(recording *capture, stored *track, buffered livebuffer.WindowSegment) error {
	initName, err := s.copySeedInit(recording, stored, buffered.InitName)
	if err != nil {
		return err
	}
	data, err := s.services.Buffer.Read(recording.context, recording.camModel.Id, buffered.Name)
	if err != nil {
		return err
	}
	segment := livebuffer.Segment{Identity: buffered.Identity, Duration: buffered.Duration, Discontinuity: buffered.Discontinuity}
	return s.appendSegment(recording, stored, segment, data, initName)
}

// copySeedInit copies a buffered init section once and returns its name in the recording.
func (s *Service) copySeedInit(recording *capture, stored *track, bufferedName string) (string, error) {
	if bufferedName == "" {
		return "", nil
	}
	if name, known := recording.initName(stored, bufferedName); known {
		return name, nil
	}
	data, err := s.services.Buffer.Read(recording.context, recording.camModel.Id, bufferedName)
	if err != nil {
		return "", err
	}
	return s.saveInitSection(recording, stored, bufferedName, data)
}

// lastIdentity is the identity of the window's newest segment.
func lastIdentity(window *livebuffer.Window) string {
	if len(window.Segments) == 0 {
		return ""
	}
	return window.Segments[len(window.Segments)-1].Identity
}
