interface SparklineProps {
  data: number[]
  color?: string
}

export function Sparkline({ data, color = '#10b981' }: SparklineProps) {
  const max = Math.max(...data, 1)

  return (
    <div className="sparkline">
      {data.map((value, i) => (
        <div
          key={i}
          className="sparkline-bar"
          style={{
            height: `${(value / max) * 100}%`,
            backgroundColor: color,
            minHeight: '2px',
          }}
        />
      ))}
    </div>
  )
}
