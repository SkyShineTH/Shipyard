import { useMemo, useState } from 'react'

/**
 * Interactive architecture diagram for the case study.
 * Hover or focus a node to highlight it together with everything it talks to;
 * the rest dims so each request/delivery path is easy to read on its own.
 */

// Visual container that groups the Go services. Edges may target it by id.
const SERVICES_BOX = { id: 'services', x: 540, y: 70, w: 232, h: 206 }

const NODES = [
  { id: 'clients', label: 'Clients', sub: 'web / mobile', group: 'client', x: 24, y: 150, w: 132, h: 64 },
  { id: 'cloudflare', label: 'Cloudflare', sub: 'Full-Strict TLS', group: 'edge', x: 196, y: 118, w: 140, h: 52 },
  { id: 'lb', label: 'DO Load Balancer', sub: '', group: 'edge', x: 196, y: 196, w: 140, h: 52 },
  { id: 'fe', label: 'frontend', sub: 'nginx · proxy /api/v1', group: 'frontend', x: 372, y: 150, w: 140, h: 64 },

  { id: 'auth', label: 'auth-service', sub: 'JWT', group: 'service', x: 556, y: 106, w: 200, h: 46 },
  { id: 'todo', label: 'todo-service', sub: 'Argo Rollout', group: 'service', x: 556, y: 162, w: 200, h: 46 },
  { id: 'platform', label: 'platform-status', sub: 'read-only', group: 'service', x: 556, y: 218, w: 200, h: 46 },

  { id: 'db', label: 'PostgreSQL', sub: '', group: 'data', x: 824, y: 94, w: 152, h: 52 },
  { id: 'k8s', label: 'Kubernetes API', sub: 'read-only', group: 'source', x: 824, y: 162, w: 152, h: 48 },
  { id: 'argoApps', label: 'Argo CD Apps', sub: 'read-only', group: 'source', x: 824, y: 220, w: 152, h: 48 },

  { id: 'prometheus', label: 'Prometheus', sub: 'private', group: 'observ', x: 540, y: 326, w: 130, h: 50 },
  { id: 'grafana', label: 'Grafana', sub: 'private', group: 'observ', x: 372, y: 326, w: 140, h: 50 },

  { id: 'gha', label: 'GitHub Actions', sub: '', group: 'ci', x: 24, y: 466, w: 150, h: 54 },
  { id: 'ghcr', label: 'GHCR', sub: 'images', group: 'ci', x: 214, y: 466, w: 130, h: 54 },
  { id: 'helm', label: 'Helm values', sub: 'image tag bump', group: 'ci', x: 384, y: 466, w: 150, h: 54 },
  { id: 'argocd', label: 'Argo CD', sub: 'sync', group: 'ci', x: 574, y: 466, w: 130, h: 54 },
  { id: 'rollouts', label: 'Argo Rollouts', sub: '', group: 'ci', x: 744, y: 466, w: 150, h: 54 },
]

// draw: rendered line. Synthetic edges (draw:false) only drive highlighting.
const EDGES = [
  { from: 'clients', to: 'cloudflare' },
  { from: 'cloudflare', to: 'lb' },
  { from: 'lb', to: 'fe' },
  { from: 'fe', to: 'auth' },
  { from: 'fe', to: 'todo' },
  { from: 'fe', to: 'platform' },
  { from: 'auth', to: 'db' },
  { from: 'todo', to: 'db' },
  { from: 'platform', to: 'k8s', kind: 'read' },
  { from: 'platform', to: 'argoApps', kind: 'read' },

  { from: 'grafana', to: 'prometheus' },
  { from: 'prometheus', to: 'services', kind: 'read' },

  { from: 'gha', to: 'ghcr' },
  { from: 'ghcr', to: 'helm' },
  { from: 'helm', to: 'argocd' },
  { from: 'argocd', to: 'rollouts' },
  { from: 'argocd', to: 'services', kind: 'sync' },
  { from: 'rollouts', to: 'todo', kind: 'sync' },

  // Highlight-only links so the services box and its members react together.
  { from: 'services', to: 'auth', draw: false },
  { from: 'services', to: 'todo', draw: false },
  { from: 'services', to: 'platform', draw: false },
]

const NODE_MAP = new Map([...NODES, SERVICES_BOX].map((n) => [n.id, n]))

function center(node) {
  return { x: node.x + node.w / 2, y: node.y + node.h / 2 }
}

/** Pick connection points on the facing sides of two boxes for a clean line. */
function connect(a, b) {
  const ca = center(a)
  const cb = center(b)
  const dx = cb.x - ca.x
  const dy = cb.y - ca.y
  if (Math.abs(dx) >= Math.abs(dy)) {
    const sign = dx >= 0 ? 1 : -1
    return {
      x1: ca.x + (sign * a.w) / 2,
      y1: ca.y,
      x2: cb.x - (sign * b.w) / 2,
      y2: cb.y,
    }
  }
  const sign = dy >= 0 ? 1 : -1
  return {
    x1: ca.x,
    y1: ca.y + (sign * a.h) / 2,
    x2: cb.x,
    y2: cb.y - (sign * b.h) / 2,
  }
}

// Undirected adjacency for highlighting.
const ADJACENCY = (() => {
  const map = new Map()
  for (const { from, to } of EDGES) {
    if (!map.has(from)) map.set(from, new Set())
    if (!map.has(to)) map.set(to, new Set())
    map.get(from).add(to)
    map.get(to).add(from)
  }
  return map
})()

export default function ArchitectureDiagram() {
  const [active, setActive] = useState(null)

  const related = useMemo(() => {
    if (!active) return null
    const set = new Set([active, ...(ADJACENCY.get(active) ?? [])])
    return set
  }, [active])

  const isDim = (id) => related !== null && !related.has(id)
  const isActive = (id) => related !== null && related.has(id)
  const edgeActive = (edge) =>
    related !== null && (edge.from === active || edge.to === active)

  function handleActivate(id) {
    setActive(id)
  }
  function handleClear() {
    setActive(null)
  }

  return (
    <figure
      className={`arch-diagram${active ? ' is-focused' : ''}`}
      onMouseLeave={handleClear}
    >
      <svg
        className="arch-svg"
        viewBox="0 0 1000 560"
        role="img"
        aria-label="Shipyard architecture: request flow and GitOps delivery"
        preserveAspectRatio="xMidYMid meet"
      >
        <defs>
          <marker
            id="arch-arrow"
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="7"
            markerHeight="7"
            orient="auto-start-reverse"
          >
            <path d="M 0 0 L 10 5 L 0 10 z" fill="context-stroke" />
          </marker>
        </defs>

        {/* Band divider between runtime and delivery */}
        <line className="arch-divider" x1="24" y1="412" x2="976" y2="412" />
        <text className="arch-band-label" x="24" y="400">
          Runtime request flow
        </text>
        <text className="arch-band-label" x="24" y="448">
          GitOps delivery
        </text>

        {/* Microservices container */}
        <g
          className={`arch-box${isDim('services') ? ' is-dim' : ''}${
            isActive('services') ? ' is-active' : ''
          }`}
        >
          <rect
            className="arch-box__rect"
            x={SERVICES_BOX.x}
            y={SERVICES_BOX.y}
            width={SERVICES_BOX.w}
            height={SERVICES_BOX.h}
            rx="14"
            tabIndex={0}
            role="button"
            aria-label="Microservices group: auth-service, todo-service, platform-status-service"
            onMouseEnter={() => handleActivate('services')}
            onFocus={() => handleActivate('services')}
            onBlur={handleClear}
          />
          <text className="arch-box__label" x={SERVICES_BOX.x + 14} y={SERVICES_BOX.y + 22}>
            Microservices · Go/Gin
          </text>
        </g>

        {/* Edges */}
        <g className="arch-edges">
          {EDGES.filter((e) => e.draw !== false).map((edge) => {
            const a = NODE_MAP.get(edge.from)
            const b = NODE_MAP.get(edge.to)
            const { x1, y1, x2, y2 } = connect(a, b)
            const cls = [
              'arch-edge',
              edge.kind ? `arch-edge--${edge.kind}` : '',
              related === null ? '' : edgeActive(edge) ? 'is-active' : 'is-dim',
            ]
              .filter(Boolean)
              .join(' ')
            return (
              <line
                key={`${edge.from}-${edge.to}`}
                className={cls}
                x1={x1}
                y1={y1}
                x2={x2}
                y2={y2}
                markerEnd="url(#arch-arrow)"
              />
            )
          })}
        </g>

        {/* Nodes */}
        <g className="arch-nodes">
          {NODES.map((node) => {
            const cls = [
              'arch-node',
              `arch-node--${node.group}`,
              isDim(node.id) ? 'is-dim' : '',
              isActive(node.id) ? 'is-active' : '',
            ]
              .filter(Boolean)
              .join(' ')
            const labelY = node.sub ? node.y + node.h / 2 - 4 : node.y + node.h / 2 + 4
            return (
              <g
                key={node.id}
                className={cls}
                tabIndex={0}
                role="button"
                aria-label={node.sub ? `${node.label}, ${node.sub}` : node.label}
                onMouseEnter={() => handleActivate(node.id)}
                onFocus={() => handleActivate(node.id)}
                onBlur={handleClear}
              >
                <rect
                  className="arch-node__rect"
                  x={node.x}
                  y={node.y}
                  width={node.w}
                  height={node.h}
                  rx="10"
                />
                <rect className="arch-node__stripe" x={node.x} y={node.y} width="4" height={node.h} rx="2" />
                <text className="arch-node__label" x={node.x + node.w / 2} y={labelY}>
                  {node.label}
                </text>
                {node.sub ? (
                  <text className="arch-node__sub" x={node.x + node.w / 2} y={node.y + node.h / 2 + 14}>
                    {node.sub}
                  </text>
                ) : null}
              </g>
            )
          })}
        </g>
      </svg>

      <figcaption className="arch-legend" aria-hidden="true">
        <span className="arch-legend__item arch-legend__item--edge">Edge / TLS</span>
        <span className="arch-legend__item arch-legend__item--service">Go services</span>
        <span className="arch-legend__item arch-legend__item--data">Data</span>
        <span className="arch-legend__item arch-legend__item--source">Read-only</span>
        <span className="arch-legend__item arch-legend__item--observ">Observability</span>
        <span className="arch-legend__item arch-legend__item--ci">GitOps / CI</span>
      </figcaption>

      <p className="arch-hint muted">Hover or focus a box to trace its connections.</p>
    </figure>
  )
}
