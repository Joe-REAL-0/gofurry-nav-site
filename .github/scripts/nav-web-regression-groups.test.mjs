import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'
import { regressionGroups, regressionInventory, regressionFiles, validateRegressionGroups } from './nav-web-regression-groups.mjs'

const copyGroups = () => Object.fromEntries(Object.entries(regressionGroups).map(([group, files]) => [group, [...files]]))

test('every current regression spec has exactly one explicit group', () => {
  assert.deepEqual(Object.keys(regressionGroups), ['games', 'sites-other', 'insights-nav'])
  validateRegressionGroups()
  assert.deepEqual(Object.values(regressionGroups).flat().sort(), regressionInventory())
})

test('an existing spec cannot lose its owner', () => {
  const groups = copyGroups(), omitted = groups.games.pop()
  assert.throws(() => validateRegressionGroups(groups), error => error.message.includes(`Unclassified regression spec: ${omitted}`))
})

test('new specs require deliberate classification', () => {
  assert.throws(() => validateRegressionGroups(regressionGroups, [...regressionInventory(), 'future.spec.ts']), /Unclassified regression spec: future/)
})

test('deleted or renamed listed files fail closed', () => {
  const removed = regressionGroups.games[0]
  assert.throws(() => validateRegressionGroups(regressionGroups, regressionInventory().filter(file => file !== removed)), /Missing regression spec/)
})

test('duplicates within or across groups fail', () => {
  for (const group of ['games', 'sites-other']) {
    const groups = copyGroups()
    groups[group].push(groups.games[0])
    assert.throws(() => validateRegressionGroups(groups), /Duplicate regression spec/)
  }
})

test('empty and unknown groups cannot silently run the whole suite', () => {
  assert.throws(() => validateRegressionGroups({ ...regressionGroups, games: [] }), /Empty regression group/)
  assert.throws(() => regressionFiles('future'), /Unknown regression group/)
})

test('CLI emits only explicit Playwright paths and rejects absent/unknown selectors', () => {
  const script = fileURLToPath(new URL('./nav-web-regression-groups.mjs', import.meta.url))
  for (const group of Object.keys(regressionGroups)) {
    const result = spawnSync(process.execPath, [script, group], { encoding: 'utf8' })
    assert.equal(result.status, 0, result.stderr)
    assert.deepEqual(result.stdout.trim().split(/\r?\n/), regressionFiles(group))
  }
  for (const args of [[], ['future'], ['games', 'sites-other']]) {
    const result = spawnSync(process.execPath, [script, ...args], { encoding: 'utf8' })
    assert.equal(result.status, 1)
    assert.equal(result.stdout, '')
  }
})
