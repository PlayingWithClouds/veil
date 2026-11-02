/**
 * Face tracking: background scans of videos and performers' pictures, and
 * the guesses of who is in a video made from them.
 */
export {
	trackVideo,
	trackPerformer,
	cancelTracking,
	faceTracking,
	type FaceTracking,
	type TrackOutcome
} from './track';
export { fetchVideoFaces, type VideoFaceGuesses } from './api';
export { FACE_MODEL } from './engine';
