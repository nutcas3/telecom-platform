import React from 'react'
import { render, screen } from '@testing-library/react'
import { StatCard } from '../stat-card'
import { Activity } from 'lucide-react'

// Mock the UI components to avoid dependency issues
jest.mock('@/components/ui/card', () => ({
  Card: ({ children, className, ...props }: any) => (
    <div data-testid="card" className={className} {...props}>
      {children}
    </div>
  ),
  CardHeader: ({ children, className, ...props }: any) => (
    <div data-testid="card-header" className={className} {...props}>
      {children}
    </div>
  ),
  CardContent: ({ children, className, ...props }: any) => (
    <div data-testid="card-content" className={className} {...props}>
      {children}
    </div>
  ),
  CardTitle: ({ children, className, ...props }: any) => (
    <div data-testid="card-title" className={className} {...props}>
      {children}
    </div>
  ),
}))

// Mock the utils function
jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('StatCard', () => {
  it('renders basic stat card with title, value, and icon', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        icon={Activity}
      />
    )

    expect(screen.getByText('Total Users')).toBeInTheDocument()
    expect(screen.getByText('1,234')).toBeInTheDocument()
    expect(screen.getByTestId('card')).toBeInTheDocument()
    expect(screen.getByTestId('card-header')).toBeInTheDocument()
    expect(screen.getByTestId('card-content')).toBeInTheDocument()
  })

  it('renders with description', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        description="Active users this month"
        icon={Activity}
      />
    )

    expect(screen.getByText('Active users this month')).toBeInTheDocument()
  })

  it('renders with positive trend', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        trend={{ value: 12, positive: true }}
        icon={Activity}
      />
    )

    const trendSpan = document.querySelector('.text-emerald-600')
    expect(trendSpan).toBeInTheDocument()
    expect(trendSpan).toHaveClass('text-emerald-600')
    expect(trendSpan?.textContent).toContain('12')
    expect(trendSpan?.textContent).toContain('+')
  })

  it('renders with negative trend', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        trend={{ value: 5, positive: false }}
        icon={Activity}
      />
    )

    // Component only adds '+' for positive, nothing for negative - text is split across nodes
    const trendSpan = document.querySelector('.text-red-600')
    expect(trendSpan).toBeInTheDocument()
    expect(trendSpan).toHaveClass('text-red-600')
    expect(trendSpan?.textContent).toContain('5')
  })

  it('renders with both trend and description', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        trend={{ value: 12, positive: true }}
        description="Active users this month"
        icon={Activity}
      />
    )

    const trendSpan = document.querySelector('.text-emerald-600')
    expect(trendSpan?.textContent).toContain('12')
    expect(screen.getByText(/Active users this month/)).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        icon={Activity}
        className="custom-class"
      />
    )

    const card = screen.getByTestId('card')
    expect(card).toHaveClass('custom-class')
  })

  it('renders numeric value correctly', () => {
    render(
      <StatCard
        title="Total Users"
        value={1234}
        icon={Activity}
      />
    )

    expect(screen.getByText('1234')).toBeInTheDocument()
  })

  it('does not render trend or description when not provided', () => {
    render(
      <StatCard
        title="Total Users"
        value="1,234"
        icon={Activity}
      />
    )

    // Should not have any trend indicators or description text
    expect(screen.queryByText(/%/)).not.toBeInTheDocument()
    expect(screen.getByTestId('card-content')).toBeInTheDocument()
    expect(screen.getByTestId('card-content').children).toHaveLength(1) // Only the value
  })
})
