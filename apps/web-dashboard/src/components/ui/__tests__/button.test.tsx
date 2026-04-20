import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from '../button'

// Mock Radix UI Slot - clone child and pass props through
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

// Mock the utils function
jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('Button', () => {
  it('renders with default props', () => {
    render(<Button>Click me</Button>)
    
    const button = screen.getByRole('button', { name: 'Click me' })
    expect(button).toBeInTheDocument()
    expect(button).toHaveAttribute('data-variant', 'default')
    expect(button).toHaveAttribute('data-size', 'default')
    expect(button).toHaveAttribute('data-slot', 'button')
  })

  it('renders with different variants', () => {
    const { rerender } = render(<Button variant="outline">Outline</Button>)
    
    let button = screen.getByRole('button', { name: 'Outline' })
    expect(button).toHaveAttribute('data-variant', 'outline')

    rerender(<Button variant="destructive">Destructive</Button>)
    button = screen.getByRole('button', { name: 'Destructive' })
    expect(button).toHaveAttribute('data-variant', 'destructive')

    rerender(<Button variant="ghost">Ghost</Button>)
    button = screen.getByRole('button', { name: 'Ghost' })
    expect(button).toHaveAttribute('data-variant', 'ghost')

    rerender(<Button variant="link">Link</Button>)
    button = screen.getByRole('button', { name: 'Link' })
    expect(button).toHaveAttribute('data-variant', 'link')
  })

  it('renders with different sizes', () => {
    const { rerender } = render(<Button size="sm">Small</Button>)
    
    let button = screen.getByRole('button', { name: 'Small' })
    expect(button).toHaveAttribute('data-size', 'sm')

    rerender(<Button size="lg">Large</Button>)
    button = screen.getByRole('button', { name: 'Large' })
    expect(button).toHaveAttribute('data-size', 'lg')

    rerender(<Button size="xs">Extra Small</Button>)
    button = screen.getByRole('button', { name: 'Extra Small' })
    expect(button).toHaveAttribute('data-size', 'xs')

    rerender(<Button size="icon">Icon</Button>)
    button = screen.getByRole('button', { name: 'Icon' })
    expect(button).toHaveAttribute('data-size', 'icon')
  })

  it('applies custom className', () => {
    render(<Button className="custom-class">Custom</Button>)
    
    const button = screen.getByRole('button', { name: 'Custom' })
    expect(button).toHaveClass('custom-class')
  })

  it('handles click events', async () => {
    const user = userEvent.setup()
    const handleClick = jest.fn()
    
    render(<Button onClick={handleClick}>Click me</Button>)
    
    const button = screen.getByRole('button', { name: 'Click me' })
    await user.click(button)
    
    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('can be disabled', () => {
    render(<Button disabled>Disabled</Button>)
    
    const button = screen.getByRole('button', { name: 'Disabled' })
    expect(button).toBeDisabled()
  })

  it('handles disabled state', async () => {
    const user = userEvent.setup()
    const handleClick = jest.fn()
    
    render(<Button disabled onClick={handleClick}>Disabled</Button>)
    
    const button = screen.getByRole('button', { name: 'Disabled' })
    expect(button).toBeDisabled()
    
    await user.click(button)
    expect(handleClick).not.toHaveBeenCalled()
  })

  it('renders as child component when asChild is true', () => {
    render(
      <Button asChild>
        <a href="/test">Link Button</a>
      </Button>
    )
    
    const link = screen.getByRole('link', { name: 'Link Button' })
    expect(link).toBeInTheDocument()
    expect(link).toHaveAttribute('href', '/test')
    expect(link).toHaveAttribute('data-slot', 'button')
  })

  it('passes through additional props', () => {
    render(<Button data-testid="test-button" type="submit">Submit</Button>)
    
    const button = screen.getByTestId('test-button')
    expect(button).toHaveAttribute('type', 'submit')
  })

  it('renders with icons', () => {
    render(
      <Button>
        <span data-testid="icon">Icon</span>
        With Icon
      </Button>
    )
    
    const button = screen.getByRole('button', { name: /With Icon/ })
    const icon = screen.getByTestId('icon')
    
    expect(button).toContainElement(icon)
    expect(icon).toBeInTheDocument()
  })

  it('handles form submission', async () => {
    const user = userEvent.setup()
    const handleSubmit = jest.fn((e) => e.preventDefault())
    
    render(
      <form onSubmit={handleSubmit}>
        <Button type="submit">Submit</Button>
      </form>
    )
    
    const button = screen.getByRole('button', { name: 'Submit' })
    await user.click(button)
    
    expect(handleSubmit).toHaveBeenCalledTimes(1)
  })
})
