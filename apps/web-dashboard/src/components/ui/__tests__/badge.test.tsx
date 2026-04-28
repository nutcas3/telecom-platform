import React from 'react'
import { render, screen } from '@testing-library/react'
import { Badge } from '../badge'

jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('Badge', () => {
  it('renders with default variant', () => {
    render(<Badge>New</Badge>)
    expect(screen.getByText('New')).toBeInTheDocument()
  })

  it('applies default variant classes', () => {
    render(<Badge>Default</Badge>)
    const badge = screen.getByText('Default')
    expect(badge.className).toContain('bg-primary')
  })

  it('applies secondary variant', () => {
    render(<Badge variant="secondary">Secondary</Badge>)
    expect(screen.getByText('Secondary').className).toContain('bg-secondary')
  })

  it('applies destructive variant', () => {
    render(<Badge variant="destructive">Error</Badge>)
    expect(screen.getByText('Error').className).toContain('text-destructive')
  })

  it('applies success variant', () => {
    render(<Badge variant="success">OK</Badge>)
    expect(screen.getByText('OK').className).toContain('text-emerald-600')
  })

  it('applies warning variant', () => {
    render(<Badge variant="warning">Warn</Badge>)
    expect(screen.getByText('Warn').className).toContain('text-amber-600')
  })

  it('applies outline variant', () => {
    render(<Badge variant="outline">Outline</Badge>)
    expect(screen.getByText('Outline').className).toContain('text-foreground')
  })

  it('merges custom className', () => {
    render(<Badge className="my-custom">Tag</Badge>)
    expect(screen.getByText('Tag').className).toContain('my-custom')
  })

  it('forwards HTML attributes', () => {
    render(<Badge data-testid="badge-id" title="tooltip">X</Badge>)
    const badge = screen.getByTestId('badge-id')
    expect(badge).toHaveAttribute('title', 'tooltip')
  })
})
