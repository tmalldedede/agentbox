import { type SVGProps } from 'react'

export function AgentBoxLogo(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      width="40"
      height="40"
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      {/* Main hexagon - bold */}
      <path
        d="M20 3 L35 11.5 L35 28.5 L20 37 L5 28.5 L5 11.5 Z"
        stroke="currentColor"
        strokeWidth="2.5"
        fill="none"
      />

      {/* Central hexagon core */}
      <path
        d="M20 13 L26 16.5 L26 23.5 L20 27 L14 23.5 L14 16.5 Z"
        fill="currentColor"
      />

      {/* Energy lines radiating out */}
      <line x1="20" y1="13" x2="20" y2="6" stroke="currentColor" strokeWidth="2" />
      <line x1="26" y1="16.5" x2="32" y2="13" stroke="currentColor" strokeWidth="2" />
      <line x1="26" y1="23.5" x2="32" y2="27" stroke="currentColor" strokeWidth="2" />
      <line x1="20" y1="27" x2="20" y2="34" stroke="currentColor" strokeWidth="2" />
      <line x1="14" y1="23.5" x2="8" y2="27" stroke="currentColor" strokeWidth="2" />
      <line x1="14" y1="16.5" x2="8" y2="13" stroke="currentColor" strokeWidth="2" />

      {/* Corner accents */}
      <circle cx="20" cy="5" r="2" fill="currentColor" />
      <circle cx="33" cy="12" r="2" fill="currentColor" />
      <circle cx="33" cy="28" r="2" fill="currentColor" />
      <circle cx="20" cy="35" r="2" fill="currentColor" />
      <circle cx="7" cy="28" r="2" fill="currentColor" />
      <circle cx="7" cy="12" r="2" fill="currentColor" />
    </svg>
  )
}
