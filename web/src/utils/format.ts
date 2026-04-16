import moment from 'moment'

export function formatDate(date?: Date): string {
  if (!date) return '-'
  return moment(date).fromNow()
}

export function formatDateFull(date?: Date): string {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}
