import { useEffect, useState } from 'react'
import { api } from '../api'
import type { Metrics } from '../types'
import { useToast } from '../components/Toast'
import BoardTabs from '../components/BoardTabs'
import { fmtDuration } from '../format'

export default function MetricsPage({ id }: { id: number }) {
  const toast = useToast()
  const [m, setM] = useState<Metrics | null>(null)
  const [name, setName] = useState('')

  useEffect(() => {
    api.metrics(id).then(setM).catch((e) => toast(e.message))
    api.board(id).then((b) => setName(b.name)).catch(() => {})
  }, [id, toast])

  if (!m) return (
    <>
      <BoardTabs boardId={id} name={name || '…'} active="metrics" />
      <main className="page"><p className="muted">Cargando…</p></main>
    </>
  )
  const maxCount = Math.max(1, ...m.columns.map((c) => Math.max(c.count, c.wip_limit ?? 0)))

  return (
    <>
    <BoardTabs boardId={id} name={name} active="metrics" />
    <main className="page">

      <section className="tiles">
        <div className="tile"><span>Total de tarjetas</span><strong>{m.total_cards}</strong></div>
        <div className="tile"><span>Sin iniciar / otras</span><strong>{m.totals_by_kind.normal}</strong></div>
        <div className="tile"><span>En curso</span><strong>{m.totals_by_kind.active}</strong></div>
        <div className="tile"><span>Terminadas</span><strong>{m.totals_by_kind.done}</strong></div>
      </section>

      <section className="panel">
        <h3>Tarjetas por columna · WIP actual vs. límite</h3>
        <table className="metrics-table">
          <thead><tr><th>Columna</th><th>Tarjetas</th><th>WIP</th><th /></tr></thead>
          <tbody>
            {m.columns.map((c) => {
              const full = c.wip_limit !== null && c.count >= c.wip_limit
              const over = c.wip_limit !== null && c.count > c.wip_limit
              return (
                <tr key={c.id}>
                  <td>{c.title} <span className={`kind kind-${c.kind}`}>{c.kind === 'active' ? 'en curso' : c.kind === 'done' ? 'terminado' : ''}</span></td>
                  <td>{c.count}</td>
                  <td className={over ? 'bad' : full ? 'warn' : ''}>
                    {c.wip_limit === null ? <span className="muted">sin límite</span> : `${c.count} / ${c.wip_limit}`}
                  </td>
                  <td className="bar-cell">
                    <div className="bar-track">
                      <div className={`bar-fill ${over ? 'bad' : full ? 'warn' : ''}`} style={{ width: `${(c.count / maxCount) * 100}%` }} />
                      {c.wip_limit !== null && <div className="bar-limit" style={{ left: `${(c.wip_limit / maxCount) * 100}%` }} title={`Límite ${c.wip_limit}`} />}
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </section>

      <section className="panel">
        <h3>Cycle time <span className="muted">(primera entrada a «en curso» → entrada a «terminado»)</span></h3>
        {m.cycle_time.count === 0 ? (
          <p className="muted">Aún no hay tarjetas terminadas que hayan pasado por una columna «en curso». Configura el tipo de cada columna con ⚙.</p>
        ) : (
          <>
            <div className="tiles">
              <div className="tile"><span>Promedio</span><strong>{fmtDuration(m.cycle_time.avg_hours)}</strong></div>
              <div className="tile"><span>Mediana</span><strong>{fmtDuration(m.cycle_time.median_hours)}</strong></div>
              <div className="tile"><span>Percentil 85</span><strong>{fmtDuration(m.cycle_time.p85_hours)}</strong></div>
              <div className="tile"><span>Tarjetas medidas</span><strong>{m.cycle_time.count}</strong></div>
            </div>
            <table className="metrics-table">
              <thead><tr><th>Últimas terminadas</th><th>Cycle time</th></tr></thead>
              <tbody>
                {m.cycle_time.recent.map((r) => (
                  <tr key={r.card_id}><td>{r.title}</td><td>{fmtDuration(r.hours)}</td></tr>
                ))}
              </tbody>
            </table>
          </>
        )}
      </section>
    </main>
    </>
  )
}
