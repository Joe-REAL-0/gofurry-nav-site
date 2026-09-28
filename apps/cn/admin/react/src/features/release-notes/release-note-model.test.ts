import { describe, expect, it } from 'vitest'
import { isFutureReleaseTime, parseReleaseTime, releaseDisplayStatus, releaseNoteForm, releaseNotePayload, releaseNoteSchema, releaseWallTime } from './release-note-model'

describe('Release Notes presentation and form', () => {
  const now = Date.parse('2026-09-28T04:00:00Z')
  it('derives badges against China wall time, without persisting scheduled', () => {
    expect(releaseDisplayStatus({ publication_state: 'draft', published_at: '2099-01-01 12:00:00' }, now)).toBe('Draft')
    expect(releaseDisplayStatus({ publication_state: 'published', published_at: '2026-09-28 12:00:01' }, now)).toBe('Scheduled')
    expect(releaseDisplayStatus({ publication_state: 'published', published_at: '2026-09-28 12:00:00' }, now)).toBe('Published')
    expect(releaseDisplayStatus({ publication_state: 'published', published_at: null }, now)).toBe('Published')
    expect(releaseDisplayStatus({ publication_state: 'published', published_at: '2026-09-28T12:00:01+08:00' }, now)).toBe('Scheduled')
  })
  it('parses dates consistently and rejects invalid calendar dates', () => {
    for (const value of ['2026-09-28 12:00:00', '2026-09-28T12:00', '2026-09-28T04:00:00Z']) expect(parseReleaseTime(value)).toBe(now)
    for (const value of ['', 'bad', '2026-02-30 12:00:00', '2026-09-28 25:00:00']) expect(parseReleaseTime(value)).toBeNull()
    expect(releaseWallTime('2026-09-28T04:00:00Z')).toBe('2026-09-28 12:00:00')
    expect(isFutureReleaseTime('', now)).toBe(false)
    expect(isFutureReleaseTime('2026-09-28 11:59:59', now)).toBe(false)
  })
  it('allows incomplete drafts and arbitrary versions, but normalizes metadata only', () => {
    const draft = releaseNoteForm()
    expect(releaseNoteSchema.safeParse(draft).success).toBe(true)
    expect(releaseNotePayload({ ...draft, version: ' Maintenance Update ', commit_sha: ' ABCDEF0 ', body: '# Source\n' })).toMatchObject({ version: 'Maintenance Update', commit_sha: 'abcdef0', body: '# Source\n', published_at: '' })
    expect(releaseNotePayload(draft)).toMatchObject({ version: null, commit_sha: null })
    expect(releaseNotePayload(draft)).not.toHaveProperty('publication_state')
    expect(releaseNoteSchema.safeParse({ ...draft, commit_sha: 'not-hex' }).success).toBe(false)
  })
})
