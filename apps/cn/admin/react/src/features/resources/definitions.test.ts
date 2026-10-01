import { createElement, Fragment } from 'react'
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { findResource, resourceDefinitions } from './definitions'

describe('Resource Engine definitions', () => {
  it('keeps simple resources schema-driven and excludes Site/Game workspaces', () => {
    const keys = resourceDefinitions.map((definition) => `${definition.section}/${definition.key}`)
    expect(new Set(keys).size).toBe(keys.length)
    expect(keys).toEqual(expect.arrayContaining(['nav/sayings', 'nav/site-groups', 'game/tags', 'game/comments', 'game/prizes']))
    expect(keys).not.toContain('nav/update-notices')
    expect(findResource('nav', 'update-notices')).toBeUndefined()
    expect(keys).not.toContain('nav/sites')
    expect(keys).not.toContain('game/games')
  })

  it('uses Zod to reject invalid business input', () => {
    const sayings = findResource('nav', 'sayings')
    expect(sayings?.schema.safeParse({ language: 'zh', author: '', saying: '' }).success).toBe(false)
    expect(sayings?.schema.safeParse({ language: 'zh', author: 'GoFurry', saying: '保持好奇。' }).success).toBe(true)
  })

  it('constrains long comment previews without changing the editor value', () => {
    const comments = findResource('game', 'comments')
    const content = comments?.columns.find((column) => column.key === 'content')
    render(createElement(Fragment, null, content?.format?.('完整的长评论内容', {})))
    expect(screen.getByTitle('完整的长评论内容')).toHaveClass('max-w-[36rem]', 'text-ellipsis')
  })

  it('normalizes prize key lines only at submission and retains both languages', () => {
    const definition = findResource('game', 'prizes')!
    const draft = { ...definition.defaults, title: '中文', title_en: 'English', start_time: '2026-10-01 19:00', end_time: '2026-10-07 18:00',
      desc: '描述', desc_en: 'Description', prize: { title: '礼品卡', title_en: 'Gift card', platform: '平台', platform_en: 'Platform', keys: [' A\r', '', '  ', 'B ', ''] } }
    const parsed = definition.schema.parse(draft)
    expect(parsed).toEqual({ ...draft, prize: { ...draft.prize, keys: ['A', 'B'] } })
    expect(draft.prize.keys).toEqual([' A\r', '', '  ', 'B ', ''])
    expect(definition.schema.safeParse({ ...draft, prize: { ...draft.prize, keys: [' ', '\r', ''] } }).success).toBe(false)
  })
})
