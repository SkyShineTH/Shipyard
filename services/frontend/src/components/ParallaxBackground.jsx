import { useEffect, useRef, useState } from 'react'

function useMouseParallax() {
  const [layers, setLayers] = useState([
    { x: 0, y: 0 },
    { x: 0, y: 0 },
    { x: 0, y: 0 },
    { x: 0, y: 0 },
  ])
  const rafRef = useRef(null)
  const targetRef = useRef({ mx: 0, my: 0 })
  const currentRef = useRef([
    { x: 0, y: 0 },
    { x: 0, y: 0 },
    { x: 0, y: 0 },
    { x: 0, y: 0 },
  ])

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      return undefined
    }

    function onMove(e) {
      const cx = window.innerWidth / 2
      const cy = window.innerHeight / 2
      targetRef.current.mx = (e.clientX - cx) / cx
      targetRef.current.my = (e.clientY - cy) / cy
    }

    window.addEventListener('mousemove', onMove)

    const speeds = [0.012, 0.024, 0.042, 0.064]

    function tick() {
      const { mx, my } = targetRef.current
      let changed = false
      const next = currentRef.current.map((c, i) => {
        const tx = mx * (i + 1) * 14
        const ty = my * (i + 1) * 10
        const nx = c.x + (tx - c.x) * speeds[i]
        const ny = c.y + (ty - c.y) * speeds[i]
        if (Math.abs(nx - c.x) > 0.01 || Math.abs(ny - c.y) > 0.01) changed = true
        return { x: nx, y: ny }
      })
      currentRef.current = next
      if (changed) setLayers([...next])
      rafRef.current = requestAnimationFrame(tick)
    }

    rafRef.current = requestAnimationFrame(tick)

    return () => {
      window.removeEventListener('mousemove', onMove)
      if (rafRef.current) cancelAnimationFrame(rafRef.current)
    }
  }, [])

  return layers
}

/** Subtle nautical / chart motifs — from Figma Make; motion respects prefers-reduced-motion. */
export default function ParallaxBackground() {
  const layers = useMouseParallax()

  const accent = 'var(--accent)'

  return (
    <div className="parallax-bg" aria-hidden="true">
      <svg
        className="parallax-bg__dots"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <pattern
            id="shipyard-dotgrid"
            x="0"
            y="0"
            width="32"
            height="32"
            patternUnits="userSpaceOnUse"
          >
            <circle cx="1" cy="1" r="1" className="parallax-bg__dot" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#shipyard-dotgrid)" />
      </svg>

      <div
        className="parallax-bg__layer parallax-bg__layer--compass"
        style={{
          transform: `translate(${layers[0].x}px, ${layers[0].y}px)`,
        }}
      >
        <svg width="420" height="420" viewBox="0 0 420 420" fill="none">
          <circle cx="210" cy="210" r="200" stroke={accent} strokeWidth="1.5" />
          <circle
            cx="210"
            cy="210"
            r="160"
            stroke={accent}
            strokeWidth="1"
            strokeDasharray="6 8"
          />
          <circle cx="210" cy="210" r="120" stroke={accent} strokeWidth="0.8" />
          <circle
            cx="210"
            cy="210"
            r="80"
            stroke={accent}
            strokeWidth="0.6"
            strokeDasharray="3 6"
          />
          <line x1="10" y1="210" x2="410" y2="210" stroke={accent} strokeWidth="0.6" />
          <line x1="210" y1="10" x2="210" y2="410" stroke={accent} strokeWidth="0.6" />
          <line x1="68" y1="68" x2="352" y2="352" stroke={accent} strokeWidth="0.4" />
          <line x1="352" y1="68" x2="68" y2="352" stroke={accent} strokeWidth="0.4" />
          {[0, 45, 90, 135, 180, 225, 270, 315].map((deg) => {
            const rad = (deg * Math.PI) / 180
            const x1 = 210 + 196 * Math.cos(rad)
            const y1 = 210 + 196 * Math.sin(rad)
            const x2 = 210 + 204 * Math.cos(rad)
            const y2 = 210 + 204 * Math.sin(rad)
            return (
              <line
                key={deg}
                x1={x1}
                y1={y1}
                x2={x2}
                y2={y2}
                stroke={accent}
                strokeWidth="1.5"
              />
            )
          })}
        </svg>
      </div>

      <div
        className="parallax-bg__layer parallax-bg__layer--chart"
        style={{
          transform: `translate(${layers[1].x}px, ${layers[1].y}px)`,
        }}
      >
        <svg width="260" height="200" viewBox="0 0 260 200" fill="none">
          {[0, 1, 2, 3, 4].map((i) => (
            <line
              key={`h${i}`}
              x1="0"
              y1={i * 40}
              x2="260"
              y2={i * 40}
              stroke={accent}
              strokeWidth="0.8"
            />
          ))}
          {[0, 1, 2, 3, 4, 5, 6].map((i) => (
            <line
              key={`v${i}`}
              x1={i * 40}
              y1="0"
              x2={i * 40}
              y2="200"
              stroke={accent}
              strokeWidth="0.8"
            />
          ))}
          <path
            d="M 0 160 C 40 140, 80 60, 120 80 S 200 40, 260 20"
            stroke={accent}
            strokeWidth="2"
            fill="none"
            strokeDasharray="5 4"
          />
          {[0, 60, 120, 180, 240].map((x, i) => {
            const ys = [160, 100, 80, 60, 30]
            return <circle key={x} cx={x} cy={ys[i]} r="3.5" fill={accent} />
          })}
        </svg>
      </div>

      <div
        className="parallax-bg__layer parallax-bg__layer--rings"
        style={{
          transform: `translate(${layers[2].x}px, ${layers[2].y}px)`,
        }}
      >
        <svg width="200" height="200" viewBox="0 0 200 200" fill="none">
          {[90, 70, 50, 30, 14].map((r, i) => (
            <circle
              key={r}
              cx="100"
              cy="100"
              r={r}
              stroke={accent}
              strokeWidth={i === 0 ? 1 : 0.7}
              strokeDasharray={i % 2 === 1 ? '4 5' : undefined}
            />
          ))}
          <line x1="100" y1="10" x2="100" y2="190" stroke={accent} strokeWidth="0.5" />
          <line x1="10" y1="100" x2="190" y2="100" stroke={accent} strokeWidth="0.5" />
        </svg>
      </div>

      <div
        className="parallax-bg__layer parallax-bg__layer--waves"
        style={{
          transform: `translate(${layers[3].x}px, ${layers[3].y}px)`,
        }}
      >
        <svg width="220" height="80" viewBox="0 0 220 80" fill="none">
          {[0, 20, 40].map((dy, i) => (
            <path
              key={dy}
              d={`M 0 ${30 + dy} Q 27 ${10 + dy} 55 ${30 + dy} Q 82 ${50 + dy} 110 ${30 + dy} Q 137 ${10 + dy} 165 ${30 + dy} Q 192 ${50 + dy} 220 ${30 + dy}`}
              stroke={accent}
              strokeWidth={i === 0 ? 1.5 : 0.8}
              fill="none"
            />
          ))}
        </svg>
      </div>
    </div>
  )
}
