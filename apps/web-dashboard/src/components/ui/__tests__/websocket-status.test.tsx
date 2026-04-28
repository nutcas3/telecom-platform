import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WebSocketStatus } from '../websocket-status'

jest.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

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

const mockReconnect = jest.fn()
const mockHookValue = {
  isConnected: true,
  reconnect: mockReconnect,
  reconnectAttempts: 0,
}

jest.mock('@/hooks/use-websocket', () => ({
  useWebSocketConnection: () => mockHookValue,
}))

describe('WebSocketStatus', () => {
  beforeEach(() => {
    mockReconnect.mockClear()
    mockHookValue.isConnected = true
    mockHookValue.reconnectAttempts = 0
  })

  it('shows Live badge when connected', () => {
    mockHookValue.isConnected = true
    render(<WebSocketStatus />)
    expect(screen.getByText('Live')).toBeInTheDocument()
  })

  it('does not show Reconnect button when connected', () => {
    mockHookValue.isConnected = true
    render(<WebSocketStatus />)
    expect(screen.queryByRole('button', { name: /Reconnect/ })).not.toBeInTheDocument()
  })

  it('shows Offline badge when disconnected', () => {
    mockHookValue.isConnected = false
    render(<WebSocketStatus />)
    expect(screen.getByText('Offline')).toBeInTheDocument()
  })

  it('shows Reconnect button when disconnected', () => {
    mockHookValue.isConnected = false
    render(<WebSocketStatus />)
    expect(screen.getByRole('button', { name: /Reconnect/ })).toBeInTheDocument()
  })

  it('calls reconnect when button clicked', async () => {
    const user = userEvent.setup()
    mockHookValue.isConnected = false
    render(<WebSocketStatus />)
    await user.click(screen.getByRole('button', { name: /Reconnect/ }))
    expect(mockReconnect).toHaveBeenCalledTimes(1)
  })

  it('shows reconnect attempts count when > 0', () => {
    mockHookValue.isConnected = false
    mockHookValue.reconnectAttempts = 3
    render(<WebSocketStatus />)
    expect(screen.getByText(/Attempts:/)).toBeInTheDocument()
    expect(screen.getByText(/3/)).toBeInTheDocument()
  })

  it('hides attempts count when 0', () => {
    mockHookValue.isConnected = true
    mockHookValue.reconnectAttempts = 0
    render(<WebSocketStatus />)
    expect(screen.queryByText(/Attempts:/)).not.toBeInTheDocument()
  })
})
