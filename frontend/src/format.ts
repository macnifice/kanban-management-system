// Formato de duraciones para el dashboard de métricas.
// Se redondea primero y después se elige la unidad (evita «60 s», «60 min» y «48.0 h»).
export function fmtDuration(hours: number): string {
  const seconds = Math.round(hours * 3600)
  if (seconds < 60) return `${seconds} s`

  const minutes = Math.round(hours * 60)
  if (minutes < 60) return `${minutes} min`

  const tenthsOfHour = Math.round(hours * 10)
  if (tenthsOfHour < 480) return `${(tenthsOfHour / 10).toFixed(1)} h`

  return `${(hours / 24).toFixed(1)} d`
}
