import { describe, it, expect } from 'vitest'
import {
  DEFAULT_STORAGE_COPY_METHOD,
  MIGRATION_FORM_DEFAULTS,
  ROLLING_FORM_DEFAULTS,
  STORAGE_COPY_METHOD_OPTIONS
} from './constants'

describe('storage copy method defaults', () => {
  it('defaults to vJailbreak Accelerated Copy', () => {
    expect(DEFAULT_STORAGE_COPY_METHOD).toBe('HotAdd')
  })

  it('offers the default as a selectable option, listed first', () => {
    const values = STORAGE_COPY_METHOD_OPTIONS.map((option) => option.value)
    expect(values).toContain(DEFAULT_STORAGE_COPY_METHOD)
    expect(values[0]).toBe(DEFAULT_STORAGE_COPY_METHOD)
  })

  it('preselects the same method in the Migration Form and the Cluster Conversion form', () => {
    expect(MIGRATION_FORM_DEFAULTS.storageCopyMethod).toBe(DEFAULT_STORAGE_COPY_METHOD)
    expect(ROLLING_FORM_DEFAULTS.storageCopyMethod).toBe(
      MIGRATION_FORM_DEFAULTS.storageCopyMethod
    )
  })

  it('keeps removeVMwareTools on by default in both forms', () => {
    expect(MIGRATION_FORM_DEFAULTS.removeVMwareTools).toBe(true)
    expect(ROLLING_FORM_DEFAULTS.removeVMwareTools).toBe(true)
  })
})
