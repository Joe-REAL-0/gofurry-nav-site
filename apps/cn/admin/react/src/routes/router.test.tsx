import { isValidElement } from 'react'
import { matchRoutes } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { DATAOPS_READ_CAPABILITY } from '../lib/capabilities'
import { router } from './router'

describe('content workspace routing', () => {
  it.each([
    ['/nav/sites', 'nav/sites'], ['/nav/sites/42', 'nav/sites/:id'],
    ['/game/games', 'game/games'], ['/game/games/17', 'game/games/:id'],
  ])('matches %s to the dedicated route', (pathname, expected) => {
    const matches = matchRoutes(router.routes, pathname)
    expect(matches?.at(-1)?.route.path).toBe(expected)
  })

  it.each([
    ['/collaboration', 'collaboration'], ['/collection', 'collection'], ['/metrics', 'metrics'], ['/changes', 'changes'],
    ['/system/data-operations', 'system/data-operations'], ['/system/audit', 'system/audit'], ['/system/accounts', 'system/accounts'],
  ])('matches operational route %s natively', (pathname, expected) => {
    const matches = matchRoutes(router.routes, pathname)
    expect(matches?.at(-1)?.route.path).toBe(expected)
  })

  it.each([
    ['/nav/site-groups', 'nav'], ['/nav/sayings', 'nav'],
    ['/game/tags', 'game'], ['/game/comments', 'game'], ['/game/prizes', 'game'],
  ])('binds generic resource route %s to its explicit domain', (pathname, section) => {
    const matches = matchRoutes(router.routes, pathname)
    const element = matches?.at(-1)?.route.element

    expect(matches?.at(-1)?.route.path).toBe(`${section}/:resource`)
    expect(isValidElement<{ section?: string }>(element) ? element.props.section : undefined).toBe(section)
  })

  it('guards Data Operations with the canonical backend capability', () => {
    const matches = matchRoutes(router.routes, '/system/data-operations')
    const guard = matches
      ?.map((match) => match.route.element)
      .find((element) => isValidElement<{ capability?: string }>(element) && element.props.capability)
    const capability = isValidElement<{ capability?: string }>(guard) ? guard.props.capability : undefined

    expect(capability).toBe(DATAOPS_READ_CAPABILITY)
    expect(capability).not.toBe('data_ops.read')
  })
})

it('guards Collaboration independently from content', () => {
 const matches = matchRoutes(router.routes, '/collaboration')
 expect(matches?.some(({ route }) => isValidElement<{ capability?: string }>(route.element) && route.element.props.capability === 'collaboration.read')).toBe(true)
})


it.each([
  ['/nav/update-notices', 'nav/update-notices'],
  ['/nav/update-notices/new', 'nav/update-notices/new'],
  ['/nav/update-notices/17', 'nav/update-notices/:id'],
])('routes %s to the dedicated Release Notes workspace', (pathname, path) => {
  const matches = matchRoutes(router.routes, pathname)
  expect(matches?.at(-1)?.route.path).toBe(path)
  const parent = matches?.at(-2)?.route
  const paths = parent?.children?.map(route => route.path) ?? []
  expect(paths.indexOf(path)).toBeLessThan(paths.indexOf('nav/:resource'))
  expect(matches?.some(({route}) => isValidElement<{ capability?: string }>(route.element) && route.element.props.capability === 'content.read')).toBe(true)
})
