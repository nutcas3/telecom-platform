import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LoadingState, InlineLoading, ActionLoading } from '../loading-state'

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

describe('LoadingState', () => {
  it('renders children when not loading and no error', () => {
    render(
      <LoadingState loading={false}>
        <div>Content here</div>
      </LoadingState>
    )
    expect(screen.getByText('Content here')).toBeInTheDocument()
  })

  it('shows loading state when loading is true', () => {
    render(
      <LoadingState loading={true}>
        <div>Should not be visible</div>
      </LoadingState>
    )
    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByText('Should not be visible')).not.toBeInTheDocument()
  })

  it('shows custom loading text', () => {
    render(
      <LoadingState loading={true} loadingText="Fetching data...">
        <div>hidden</div>
      </LoadingState>
    )
    expect(screen.getByText('Fetching data...')).toBeInTheDocument()
  })

  it('shows error state when error is present', () => {
    render(
      <LoadingState loading={false} error="Network error">
        <div>hidden</div>
      </LoadingState>
    )
    expect(screen.getByText('Network error')).toBeInTheDocument()
    expect(screen.queryByText('hidden')).not.toBeInTheDocument()
  })

  it('renders retry button when onRetry provided in error state', () => {
    render(
      <LoadingState loading={false} error="Error" onRetry={jest.fn()}>
        <div>hidden</div>
      </LoadingState>
    )
    expect(screen.getByRole('button', { name: /Retry/ })).toBeInTheDocument()
  })

  it('calls onRetry when retry button is clicked', async () => {
    const user = userEvent.setup()
    const onRetry = jest.fn()
    render(
      <LoadingState loading={false} error="Error" onRetry={onRetry}>
        <div>hidden</div>
      </LoadingState>
    )
    await user.click(screen.getByRole('button', { name: /Retry/ }))
    expect(onRetry).toHaveBeenCalledTimes(1)
  })
})

describe('InlineLoading', () => {
  it('shows loading text when loading', () => {
    render(
      <InlineLoading loading={true}>
        <span>Content</span>
      </InlineLoading>
    )
    expect(screen.getByText('Loading...')).toBeInTheDocument()
    expect(screen.queryByText('Content')).not.toBeInTheDocument()
  })

  it('shows children when not loading', () => {
    render(
      <InlineLoading loading={false}>
        <span>Content</span>
      </InlineLoading>
    )
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('supports custom text', () => {
    render(
      <InlineLoading loading={true} text="Please wait">
        <span>hidden</span>
      </InlineLoading>
    )
    expect(screen.getByText('Please wait')).toBeInTheDocument()
  })
})

describe('ActionLoading', () => {
  it('renders children button when not loading', () => {
    render(
      <ActionLoading loading={false}>
        <button>Save</button>
      </ActionLoading>
    )
    expect(screen.getByRole('button', { name: 'Save' })).toBeInTheDocument()
  })

  it('renders disabled loading button when loading', () => {
    render(
      <ActionLoading loading={true} action="saving">
        <button>Save</button>
      </ActionLoading>
    )
    const button = screen.getByRole('button')
    expect(button).toBeDisabled()
    expect(button.textContent).toContain('saving')
  })
})
