/**
 * Audio utility functions for recording and playback
 */

export type MicrophonePermissionState = 'granted' | 'denied' | 'prompt' | 'unknown'

/**
 * Check microphone permission state
 * - 'granted': permission already given
 * - 'denied': user blocked microphone access (must change in browser settings)
 * - 'prompt': user will be prompted when requesting access
 * - 'unknown': couldn't determine state (Permissions API not supported)
 */
export async function checkMicrophonePermission(): Promise<MicrophonePermissionState> {
  try {
    const result = await navigator.permissions.query({ name: 'microphone' as PermissionName })
    return result.state as MicrophonePermissionState
  } catch (error) {
    // Permissions API might not be supported
    console.warn('Permissions API not supported:', error)
    return 'unknown'
  }
}

export class MicrophoneAccessError extends Error {
  constructor(
    message: string,
    public readonly reason: 'denied' | 'not-supported' | 'not-found' | 'unknown'
  ) {
    super(message)
    this.name = 'MicrophoneAccessError'
  }
}

/**
 * Request microphone access with detailed error information
 */
export async function requestMicrophoneAccess(): Promise<MediaStream> {
  if (!isRecordingSupported()) {
    throw new MicrophoneAccessError(
      'Audio recording is not supported in this browser',
      'not-supported'
    )
  }

  // Always try getUserMedia first - it's the only way to trigger the permission prompt.
  // Don't pre-check with Permissions API as some browsers (Android Chrome) may return
  // 'denied' as the default state even when the user has never been asked.
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    return stream
  } catch (error) {
    console.error('Failed to access microphone:', error)

    if (error instanceof DOMException) {
      if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
        throw new MicrophoneAccessError(
          'Microphone access was blocked. Please enable it in your browser settings and try again.',
          'denied'
        )
      }
      if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
        throw new MicrophoneAccessError(
          'No microphone found. Please connect a microphone and try again.',
          'not-found'
        )
      }
    }

    throw new MicrophoneAccessError('Failed to access microphone', 'unknown')
  }
}

/**
 * Get audio duration from a blob
 */
export function getAudioDuration(blob: Blob): Promise<number> {
  return new Promise((resolve, reject) => {
    const audio = new Audio()
    audio.addEventListener('loadedmetadata', () => {
      resolve(audio.duration)
    })
    audio.addEventListener('error', () => {
      reject(new Error('Failed to load audio'))
    })
    audio.src = URL.createObjectURL(blob)
  })
}

/**
 * Format duration in seconds to mm:ss
 */
export function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

/**
 * Check if browser supports audio recording
 */
export function isRecordingSupported(): boolean {
  return !!(
    navigator.mediaDevices &&
    typeof navigator.mediaDevices.getUserMedia === 'function' &&
    typeof MediaRecorder !== 'undefined'
  )
}
