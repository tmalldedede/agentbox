import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

interface DashboardChartProps {
  data: number[]
  color?: string
}

export function DashboardChart({ data, color = '#22c55e' }: DashboardChartProps) {
  const chartData = data.map((value, index) => ({
    name: index.toString(),
    value: value,
  }))

  return (
    <div className="h-full w-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
          <defs>
            <linearGradient id={`gradient-${color}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={color} stopOpacity={0.3} />
              <stop offset="95%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="rgba(255,255,255,0.05)" />
          <XAxis dataKey="name" hide />
          <YAxis tick={{ fill: '#6b7280', fontSize: 10 }} axisLine={false} tickLine={false} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1f2937',
              borderColor: 'rgba(255,255,255,0.1)',
              borderRadius: '8px',
              fontSize: '12px',
              color: '#f3f4f6',
            }}
            itemStyle={{ color: color }}
            cursor={{ stroke: 'rgba(255,255,255,0.1)', strokeWidth: 2 }}
          />
          <Area
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={2}
            fill={`url(#gradient-${color})`}
            animationDuration={1000}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

interface SparklineChartProps {
  data: number[]
  color?: string
}

export function SparklineChart({ data, color = '#22c55e' }: SparklineChartProps) {
  const chartData = data.map((value, index) => ({ i: index, v: value }))

  return (
    <div className="h-[40px] w-[80px]">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData}>
          <defs>
            <linearGradient id={`spark-${color}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.3} />
              <stop offset="100%" stopColor={color} stopOpacity={0} />
            </linearGradient>
          </defs>
          <Area
            type="monotone"
            dataKey="v"
            stroke={color}
            strokeWidth={2}
            fill={`url(#spark-${color})`}
            isAnimationActive={false}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}
