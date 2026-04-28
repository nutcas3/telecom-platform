import { cn } from '../utils'

describe('cn utility', () => {
  it('merges simple class names', () => {
    expect(cn('a', 'b')).toBe('a b')
  })

  it('filters out falsy values', () => {
    expect(cn('a', false, null, undefined, 'b')).toBe('a b')
  })

  it('deduplicates conflicting tailwind classes (twMerge)', () => {
    // tailwind-merge should keep the latter of conflicting utilities
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })

  it('handles conditional classnames via objects', () => {
    expect(cn('base', { active: true, disabled: false })).toContain('base')
    expect(cn('base', { active: true, disabled: false })).toContain('active')
    expect(cn('base', { active: true, disabled: false })).not.toContain('disabled')
  })

  it('handles arrays of classnames', () => {
    expect(cn(['a', 'b'], 'c')).toBe('a b c')
  })

  it('returns empty string with no args', () => {
    expect(cn()).toBe('')
  })

  it('handles nested arrays', () => {
    expect(cn(['a', ['b', 'c']])).toBe('a b c')
  })
})
