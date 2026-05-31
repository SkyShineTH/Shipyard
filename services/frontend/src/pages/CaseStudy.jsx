import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getPlatformStatus } from '../api/client'
import { useScrollReveal } from '../hooks/useScrollReveal'
import ArchitectureDiagram from '../components/ArchitectureDiagram'

const decisions = [
  {
    icon: 'savings',
    title: 'Cost-conscious DOKS',
    body: 'The demo runs on a small one-node pool with tuned requests so it can stay affordable while still proving Kubernetes delivery.',
  },
  {
    icon: 'sync_alt',
    title: 'GitOps delivery',
    body: 'Application state is declared through Helm charts and Argo CD Applications instead of manual cluster edits.',
  },
  {
    icon: 'rocket_launch',
    title: 'Progressive rollout',
    body: 'todo-service uses Argo Rollouts so deployments can pause, promote, and roll back from Kubernetes-native workflow.',
  },
  {
    icon: 'lock',
    title: 'Public-safe evidence',
    body: 'The live snapshot is read-only and intentionally strips secrets, tokens, pod IPs, node names, and database details.',
  },
  {
    icon: 'query_stats',
    title: 'On-demand observability',
    body: 'Prometheus and Grafana are wired for demo evidence through private ClusterIP services and port-forward access.',
  },
]

const evidenceItems = [
  'DOKS cluster serving the public demo through a DigitalOcean Load Balancer',
  'Argo CD Applications synced from the repository',
  'Argo Rollouts managing todo-service',
  'Kubernetes Secrets used for database and JWT credentials',
  'Cloudflare Full Strict TLS with an origin certificate',
  'Prometheus ServiceMonitors and a Grafana dashboard for service metrics',
]

function MaterialIcon({ name, className = '' }) {
  return (
    <span className={`material-symbols-outlined ${className}`} aria-hidden>
      {name}
    </span>
  )
}

function formatCheckedAt(value) {
  if (!value) return 'Not checked yet'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Not checked yet'
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

function statusClass(value) {
  const normalized = String(value || '').toLowerCase()
  if (
    normalized === 'healthy' ||
    normalized === 'synced' ||
    normalized === 'bound'
  ) {
    return 'is-ok'
  }
  if (normalized === 'progressing' || normalized === 'scaleddown') {
    return 'is-warn'
  }
  return 'is-muted'
}

function StatusBadge({ value }) {
  const label = value || 'Unknown'
  return <span className={`status-badge ${statusClass(label)}`}>{label}</span>
}

function SnapshotTable({ title, emptyText, children }) {
  return (
    <section className="case-snapshot__group" aria-label={title}>
      <h3>{title}</h3>
      {children || <p className="muted case-empty">{emptyText}</p>}
    </section>
  )
}

export default function CaseStudy() {
  const revealRoot = useScrollReveal()
  const [snapshot, setSnapshot] = useState(null)
  const [snapshotError, setSnapshotError] = useState('')
  const [loading, setLoading] = useState(true)

  const loadSnapshot = useCallback(async () => {
    setLoading(true)
    setSnapshotError('')
    try {
      const data = await getPlatformStatus()
      setSnapshot(data)
    } catch (err) {
      setSnapshot(null)
      setSnapshotError(
        err instanceof Error
          ? err.message
          : 'Live infrastructure snapshot unavailable',
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSnapshot()
  }, [loadSnapshot])

  return (
    <main className="page case-study-page" ref={revealRoot}>
      <section className="case-hero" data-reveal>
        <p className="page-eyebrow">Evidence-based DevOps case study</p>
        <h1>Shipyard on DigitalOcean Kubernetes</h1>
        <p className="muted case-hero__lead">
          A portfolio demo showing a React frontend, Go microservices,
          PostgreSQL, Helm, Argo CD, Argo Rollouts, Cloudflare TLS, and a
          cost-conscious DOKS deployment.
        </p>
        <div className="case-hero__actions">
          <Link to="/todos" className="btn primary">
            Open live app
          </Link>
          <a
            className="btn secondary"
            href="https://github.com/SkyShineTH/Shipyard"
            target="_blank"
            rel="noreferrer"
          >
            View repository
          </a>
        </div>
      </section>

      <section className="case-summary" aria-label="Deployment summary">
        <article className="case-metric reveal" data-reveal>
          <span className="case-metric__label">Cluster</span>
          <strong>{snapshot?.cluster?.provider || 'DigitalOcean Kubernetes'}</strong>
          <span>{snapshot?.cluster?.region || 'sgp1'}</span>
        </article>
        <article className="case-metric reveal" data-reveal>
          <span className="case-metric__label">Delivery</span>
          <strong>GitOps</strong>
          <span>GitHub Actions to Argo CD</span>
        </article>
        <article className="case-metric reveal" data-reveal>
          <span className="case-metric__label">Demo Mode</span>
          <strong>Cost-controlled</strong>
          <span>{snapshot?.cluster?.mode || 'one-node demo setup'}</span>
        </article>
      </section>

      <section className="case-section" aria-labelledby="architecture-heading">
        <div className="case-section__head" data-reveal>
          <p className="page-eyebrow">Architecture</p>
          <h2 id="architecture-heading">Request flow and delivery flow</h2>
        </div>
        <div className="card card--panel arch-card" data-reveal>
          <ArchitectureDiagram />
        </div>
      </section>

      <section className="case-section" aria-labelledby="snapshot-heading">
        <div className="case-section__head" data-reveal>
          <p className="page-eyebrow">Live infrastructure snapshot</p>
          <h2 id="snapshot-heading">Read-only Kubernetes evidence</h2>
          <p className="muted">
            This panel is served by a dedicated service account with read-only
            RBAC and sanitized output for public viewing.
          </p>
        </div>

        <div className="card card--panel case-snapshot" data-reveal>
          <div className="case-snapshot__toolbar">
            <div>
              <span className="case-snapshot__label">Last checked</span>
              <strong>{formatCheckedAt(snapshot?.checkedAt)}</strong>
            </div>
            <button
              type="button"
              className="btn secondary"
              onClick={loadSnapshot}
              disabled={loading}
            >
              {loading ? 'Refreshing' : 'Refresh'}
            </button>
          </div>

          {snapshotError ? (
            <p className="notice">
              Live snapshot unavailable. The case study is still available
              because the core page is static and public-safe.
            </p>
          ) : null}

          <div className="case-snapshot__grid">
            <SnapshotTable title="Workloads" emptyText="No workloads returned.">
              {snapshot?.workloads?.length ? (
                <div className="case-table" role="table">
                  {snapshot.workloads.map((item) => (
                    <div className="case-row" role="row" key={`${item.kind}-${item.name}`}>
                      <span>{item.name}</span>
                      <span>{item.kind}</span>
                      <span>{item.ready}</span>
                      <StatusBadge value={item.status} />
                    </div>
                  ))}
                </div>
              ) : null}
            </SnapshotTable>

            <SnapshotTable title="GitOps" emptyText="No Argo CD data returned.">
              {snapshot?.gitops?.length ? (
                <div className="case-table" role="table">
                  {snapshot.gitops.map((item) => (
                    <div className="case-row" role="row" key={item.name}>
                      <span>{item.name}</span>
                      <StatusBadge value={item.sync} />
                      <StatusBadge value={item.health} />
                    </div>
                  ))}
                </div>
              ) : null}
            </SnapshotTable>

            <SnapshotTable title="Services" emptyText="No services returned.">
              {snapshot?.services?.length ? (
                <div className="case-table" role="table">
                  {snapshot.services.map((item) => (
                    <div className="case-row" role="row" key={item.name}>
                      <span>{item.name}</span>
                      <span>{item.type}</span>
                      <span>{item.ports?.join(', ')}</span>
                    </div>
                  ))}
                </div>
              ) : null}
            </SnapshotTable>

            <SnapshotTable title="Storage" emptyText="No storage returned.">
              {snapshot?.storage?.length ? (
                <div className="case-table" role="table">
                  {snapshot.storage.map((item) => (
                    <div className="case-row" role="row" key={item.name}>
                      <span>{item.name}</span>
                      <StatusBadge value={item.status} />
                      <span>{item.size}</span>
                    </div>
                  ))}
                </div>
              ) : null}
            </SnapshotTable>
          </div>
        </div>
      </section>

      <section className="case-section" aria-labelledby="decisions-heading">
        <div className="case-section__head" data-reveal>
          <p className="page-eyebrow">Engineering decisions</p>
          <h2 id="decisions-heading">What this demo proves</h2>
        </div>
        <div className="case-decision-grid">
          {decisions.map((item) => (
            <article className="card card--panel case-decision" data-reveal key={item.title}>
              <div className="case-decision__head">
                <MaterialIcon name={item.icon} className="case-icon" />
                <h3>{item.title}</h3>
              </div>
              <p>{item.body}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="card card--panel case-evidence" data-reveal>
        <div>
          <p className="page-eyebrow">Evidence</p>
          <h2>Safe proof for reviewers</h2>
          <p className="muted">
            The page shows public-safe proof points without exposing tokens,
            passwords, kubeconfig, pod IPs, or node details.
          </p>
        </div>
        <ul>
          {evidenceItems.map((item) => (
            <li key={item}>
              <MaterialIcon name="check_circle" className="case-check" />
              <span>{item}</span>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}
