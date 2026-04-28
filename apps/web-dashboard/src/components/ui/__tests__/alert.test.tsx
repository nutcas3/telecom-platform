import React from 'react'
import { render, screen } from '@testing-library/react'
import { Alert, AlertTitle, AlertDescription, AlertAction } from '../alert'

jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('Alert', () => {
  it('renders with role=alert', () => {
    render(<Alert>Alert content</Alert>)
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('has data-slot=alert', () => {
    render(<Alert>Content</Alert>)
    expect(screen.getByRole('alert')).toHaveAttribute('data-slot', 'alert')
  })

  it('applies default variant', () => {
    render(<Alert>Default</Alert>)
    expect(screen.getByRole('alert').className).toContain('bg-card')
  })

  it('applies destructive variant', () => {
    render(<Alert variant="destructive">Error</Alert>)
    expect(screen.getByRole('alert').className).toContain('text-destructive')
  })

  it('renders AlertTitle with correct data-slot', () => {
    render(<AlertTitle>Title</AlertTitle>)
    const title = screen.getByText('Title')
    expect(title).toHaveAttribute('data-slot', 'alert-title')
  })

  it('renders AlertDescription with correct data-slot', () => {
    render(<AlertDescription>Description text</AlertDescription>)
    const desc = screen.getByText('Description text')
    expect(desc).toHaveAttribute('data-slot', 'alert-description')
  })

  it('renders AlertAction with correct data-slot', () => {
    render(<AlertAction><button>Dismiss</button></AlertAction>)
    const button = screen.getByRole('button', { name: 'Dismiss' })
    expect(button.parentElement).toHaveAttribute('data-slot', 'alert-action')
  })

  it('renders composed alert with title and description', () => {
    render(
      <Alert>
        <AlertTitle>Heads up!</AlertTitle>
        <AlertDescription>This is an alert.</AlertDescription>
      </Alert>
    )
    expect(screen.getByText('Heads up!')).toBeInTheDocument()
    expect(screen.getByText('This is an alert.')).toBeInTheDocument()
  })

  it('merges custom className on Alert', () => {
    render(<Alert className="custom-alert">X</Alert>)
    expect(screen.getByRole('alert').className).toContain('custom-alert')
  })
})
