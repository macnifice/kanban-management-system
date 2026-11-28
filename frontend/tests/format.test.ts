import { test } from 'node:test'
import assert from 'node:assert/strict'
import { fmtDuration } from '../src/format.ts'

const SEC = 1 / 3600
const MIN = 1 / 60

test('segundos', () => {
  assert.equal(fmtDuration(0), '0 s')
  assert.equal(fmtDuration(30 * SEC), '30 s')
  assert.equal(fmtDuration(59.4 * SEC), '59 s')
})

test('segundos que redondean a 60 pasan a minutos', () => {
  assert.equal(fmtDuration(59.6 * SEC), '1 min')
})

test('minutos', () => {
  assert.equal(fmtDuration(1 * MIN), '1 min')
  assert.equal(fmtDuration(30 * MIN), '30 min')
  assert.equal(fmtDuration(59.4 * MIN), '59 min')
})

test('minutos que redondean a 60 pasan a horas', () => {
  assert.equal(fmtDuration(59.6 * MIN), '1.0 h')
})

test('horas', () => {
  assert.equal(fmtDuration(1), '1.0 h')
  assert.equal(fmtDuration(36), '36.0 h')
  assert.equal(fmtDuration(47.94), '47.9 h')
})

test('horas que redondean a 48 pasan a días', () => {
  assert.equal(fmtDuration(47.96), '2.0 d')
})

test('días', () => {
  assert.equal(fmtDuration(48), '2.0 d')
  assert.equal(fmtDuration(72), '3.0 d')
})
