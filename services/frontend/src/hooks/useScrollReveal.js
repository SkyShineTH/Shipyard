import { useEffect, useRef } from 'react'

/** Adds .reveal--visible when element enters viewport (respects reduced motion). */
export function useScrollReveal(options = {}) {
  const ref = useRef(null)
  const { rootMargin = '0px 0px -8% 0px', threshold = 0.12 } = options

  useEffect(() => {
    const el = ref.current
    if (!el) return undefined

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      el.classList.add('reveal--visible')
      return undefined
    }

    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('reveal--visible')
            obs.unobserve(entry.target)
          }
        })
      },
      { rootMargin, threshold },
    )

    el.querySelectorAll('[data-reveal]').forEach((node) => obs.observe(node))

    return () => obs.disconnect()
  }, [rootMargin, threshold])

  return ref
}
