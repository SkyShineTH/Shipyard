import { Link } from 'react-router-dom'
import { useScrollReveal } from '../hooks/useScrollReveal'
import { useAuth } from '../context/authContext'

function MaterialIcon({ name }) {
  return (
    <span className="material-symbols-outlined landing-feature__icon" aria-hidden>
      {name}
    </span>
  )
}

export default function Landing() {
  const { isAuthenticated } = useAuth()
  const revealRoot = useScrollReveal()

  return (
    <div className="landing-page" ref={revealRoot}>
      <section className="landing-hero">
        <p className="page-eyebrow landing-hero__eyebrow" data-reveal>
          Full-stack GitOps platform
        </p>
        <h1 className="landing-hero__title" data-reveal>
          Shipyard
        </h1>
        <p className="muted landing-hero__lead" data-reveal>
          React, Go microservices, and Kubernetes-ready delivery — one cohesive
          stack for a modern DevOps portfolio.
        </p>
        <div className="landing-hero__actions" data-reveal>
          <Link to="/todos" className="btn primary">
            Open todos
          </Link>
          {!isAuthenticated ? (
            <Link to="/register" className="btn secondary">
              Create account
            </Link>
          ) : null}
        </div>
      </section>

      <section className="landing-features" aria-labelledby="features-heading">
        <h2 id="features-heading" className="sr-only">
          Platform capabilities
        </h2>
        <div className="landing-features__grid">
          <article className="card card--panel landing-feature reveal" data-reveal>
            <div className="landing-feature__head">
              <MaterialIcon name="schema" />
              <h3 className="landing-feature__title">Microservices architecture</h3>
            </div>
            <p className="landing-feature__body">
              Designed with isolation in mind. Decoupled services exposed over REST
              keep auth and tasks independent so you can scale and deploy each
              piece on its own cadence.
            </p>
          </article>

          <article className="card card--panel landing-feature reveal" data-reveal>
            <div className="landing-feature__head">
              <MaterialIcon name="terminal" />
              <h3 className="landing-feature__title">GitOps workflow</h3>
            </div>
            <p className="landing-feature__body">
              Infrastructure as code is first-class. Helm charts and Argo CD
              Application manifests ship in-repo for a tight loop from commit to
              cluster.
            </p>
          </article>

          <article className="card card--panel landing-feature reveal" data-reveal>
            <div className="landing-feature__head">
              <MaterialIcon name="monitoring" />
              <h3 className="landing-feature__title">Observability-ready</h3>
            </div>
            <p className="landing-feature__body">
              Clear service boundaries and health endpoints make it
              straightforward to add Prometheus metrics and OpenTelemetry tracing
              across the React UI and Go APIs when you are ready.
            </p>
          </article>
        </div>
      </section>

      <section className="card card--panel landing-cta reveal" data-reveal>
        <h2 className="landing-cta__title">Ready to explore?</h2>
        <p className="landing-cta__text">
          Sign in to manage your private task list, or browse the repo for charts,
          CI, and GitOps wiring.
        </p>
        <div className="landing-cta__actions">
          <Link to="/todos" className="btn primary">
            Go to todos
          </Link>
          <Link to="/login" className="btn secondary">
            Sign in
          </Link>
        </div>
      </section>
    </div>
  )
}
