interface AvatarVisualizerProps {
  audioLevel?: number
}

export function AvatarVisualizer({ audioLevel = 0.5 }: AvatarVisualizerProps) {
  const scale = 1 + audioLevel * 0.3

  return (
    <div
      className="absolute inset-0 rounded-full border-4 border-primary/30 animate-pulse"
      style={{
        transform: `scale(${scale})`,
        transition: 'transform 0.1s ease-out',
      }}
    />
  )
}
