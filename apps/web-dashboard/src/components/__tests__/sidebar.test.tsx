import React from 'react'
import { render, screen } from '@testing-library/react'
import { Sidebar } from '../sidebar'

// Mock next/link
jest.mock('next/link', () => {
  const Link = ({ children, href, className }: any) => (
    <a href={href} className={className}>{children}</a>
  )
  Link.displayName = 'Link'
  return { __esModule: true, default: Link }
})

// Mock next/navigation - usePathname is mocked per test
const mockUsePathname = jest.fn()
jest.mock('next/navigation', () => ({
  usePathname: () => mockUsePathname(),
}))

jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('Sidebar', () => {
  beforeEach(() => {
    mockUsePathname.mockReturnValue('/')
  })

  it('renders the brand name', () => {
    render(<Sidebar />)
    expect(screen.getByText('Telecom Admin')).toBeInTheDocument()
  })

  it('renders all navigation items', () => {
    render(<Sidebar />)
    const labels = [
      'Dashboard',
      'Subscribers',
      'Usage & Billing',
      'Payments',
      'eSIM Profiles',
      'System Health',
      'Chaos Engineering',
      'Settings',
    ]
    labels.forEach((label) => {
      expect(screen.getByText(label)).toBeInTheDocument()
    })
  })

  it('renders links with correct hrefs', () => {
    render(<Sidebar />)
    expect(screen.getByRole('link', { name: /Dashboard/ })).toHaveAttribute('href', '/')
    expect(screen.getByRole('link', { name: /Subscribers/ })).toHaveAttribute('href', '/subscribers')
    expect(screen.getByRole('link', { name: /Settings/ })).toHaveAttribute('href', '/settings')
  })

  it('marks Dashboard active when pathname is /', () => {
    mockUsePathname.mockReturnValue('/')
    render(<Sidebar />)
    const dashboardLink = screen.getByRole('link', { name: /Dashboard/ })
    expect(dashboardLink.className).toContain('bg-sidebar-accent')
  })

  it('marks a nested route active when pathname starts with that href', () => {
    mockUsePathname.mockReturnValue('/subscribers/123')
    render(<Sidebar />)
    const link = screen.getByRole('link', { name: /Subscribers/ })
    expect(link.className).toContain('bg-sidebar-accent')
  })

  it('does not mark Dashboard active on non-root paths', () => {
    mockUsePathname.mockReturnValue('/subscribers')
    render(<Sidebar />)
    const dashboardLink = screen.getByRole('link', { name: /Dashboard/ })
    // Dashboard uses exact match - should NOT contain active bg class
    expect(dashboardLink.className).toContain('text-sidebar-foreground/70')
  })
})
