import { StatusBadge } from '../../components/admin/status'
import { releaseDisplayStatus, type ReleaseNote } from './release-note-model'

export function ReleaseStatus({ record }: { record: Pick<ReleaseNote, 'publication_state' | 'published_at'> }) {
  const status = releaseDisplayStatus(record)
  return <StatusBadge tone={status === 'Draft' ? 'neutral' : status === 'Scheduled' ? 'info' : 'success'}>{status}</StatusBadge>
}
