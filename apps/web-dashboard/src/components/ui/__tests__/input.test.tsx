import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input } from '../input'

jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('Input', () => {
  it('renders an input element', () => {
    render(<Input placeholder="Name" />)
    expect(screen.getByPlaceholderText('Name')).toBeInTheDocument()
  })

  it('has data-slot attribute', () => {
    render(<Input data-testid="my-input" />)
    expect(screen.getByTestId('my-input')).toHaveAttribute('data-slot', 'input')
  })

  it('passes through type attribute', () => {
    render(<Input type="email" data-testid="email-input" />)
    expect(screen.getByTestId('email-input')).toHaveAttribute('type', 'email')
  })

  it('supports password type', () => {
    render(<Input type="password" data-testid="pwd" />)
    expect(screen.getByTestId('pwd')).toHaveAttribute('type', 'password')
  })

  it('handles user typing', async () => {
    const user = userEvent.setup()
    render(<Input data-testid="txt" />)
    const input = screen.getByTestId('txt') as HTMLInputElement
    await user.type(input, 'hello')
    expect(input.value).toBe('hello')
  })

  it('fires onChange handler', async () => {
    const user = userEvent.setup()
    const onChange = jest.fn()
    render(<Input onChange={onChange} data-testid="txt" />)
    await user.type(screen.getByTestId('txt'), 'a')
    expect(onChange).toHaveBeenCalled()
  })

  it('can be disabled', () => {
    render(<Input disabled data-testid="txt" />)
    expect(screen.getByTestId('txt')).toBeDisabled()
  })

  it('respects aria-invalid for error states', () => {
    render(<Input aria-invalid="true" data-testid="txt" />)
    expect(screen.getByTestId('txt')).toHaveAttribute('aria-invalid', 'true')
  })

  it('merges custom className', () => {
    render(<Input className="custom" data-testid="txt" />)
    expect(screen.getByTestId('txt').className).toContain('custom')
  })

  it('forwards ref and name for forms', () => {
    render(<Input name="username" data-testid="txt" defaultValue="bob" />)
    const input = screen.getByTestId('txt') as HTMLInputElement
    expect(input).toHaveAttribute('name', 'username')
    expect(input.value).toBe('bob')
  })
})
