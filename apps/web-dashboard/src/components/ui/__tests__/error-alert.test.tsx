import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ErrorAlert } from '../error-alert'

jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

// Radix Slot used by Button
jest.mock('radix-ui', () => ({
  Slot: {
    Root: ({ children, ...props }: any) => {
      if (React.isValidElement(children)) {
        return React.cloneElement(children, { ...props, ...(children.props as object) })
      }
      return children
    },
  },
}))

describe('ErrorAlert', () => {
  it('renders nothing when error is null', () => {
    const { container } = render(<ErrorAlert error={null} />)
    expect(container.firstChild).toBeNull()
  })

  it('renders error message when provided', () => {
    render(<ErrorAlert error="Something went wrong" />)
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
  })

  it('renders with destructive alert role', () => {
    render(<ErrorAlert error="Oops" />)
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  it('does not render retry button when onRetry is not provided', () => {
    render(<ErrorAlert error="Oops" />)
    expect(screen.queryByRole('button', { name: /Retry/ })).not.toBeInTheDocument()
  })

  it('renders retry button when onRetry is provided', () => {
    render(<ErrorAlert error="Oops" onRetry={jest.fn()} />)
    expect(screen.getByRole('button', { name: /Retry/ })).toBeInTheDocument()
  })

  it('invokes onRetry when retry clicked', async () => {
    const user = userEvent.setup()
    const onRetry = jest.fn()
    render(<ErrorAlert error="Oops" onRetry={onRetry} />)
    await user.click(screen.getByRole('button', { name: /Retry/ }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('renders dismiss button when onDismiss is provided', () => {
    render(<ErrorAlert error="Oops" onDismiss={jest.fn()} />)
    // dismiss button has no accessible text other than icon
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThanOrEqual(1)
  })

  it('invokes onDismiss when dismiss is clicked', async () => {
    const user = userEvent.setup()
    const onDismiss = jest.fn()
    render(<ErrorAlert error="Oops" onDismiss={onDismiss} />)
    const buttons = screen.getAllByRole('button')
    await user.click(buttons[buttons.length - 1])
    expect(onDismiss).toHaveBeenCalledTimes(1)
  })

  it('uses custom retry text', () => {
    render(<ErrorAlert error="Oops" onRetry={jest.fn()} retryText="Try again" />)
    expect(screen.getByRole('button', { name: /Try again/ })).toBeInTheDocument()
  })
})
