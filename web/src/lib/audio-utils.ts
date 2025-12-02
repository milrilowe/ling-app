/**
 * Audio utility functions for recording and playback
 */

/**
 * Check if microphone permission is granted
 */
export async function checkMicrophonePermission(): Promise<boolean> {
  try {
    const result = await navigator.permissions.query({ name: 'microphone' as PermissionName })
    return result.state === 'granted'
  } catch (error) {
    // Permissions API might not be supported
    console.warn('Permissions API not supported:', error)
    return false
  }
}

/**
 * Request microphone access
 */
export async function requestMicrophoneAccess(): Promise<MediaStream> {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    return stream
  } catch (error) {
    console.error('Failed to access microphone:', error)
    throw new Error('Microphone access denied')
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
    navigator.mediaDevices.getUserMedia &&
    typeof MediaRecorder !== 'undefined'
  )
}
