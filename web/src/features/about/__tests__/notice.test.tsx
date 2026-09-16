import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { DefaultAboutContent } from '../index'

describe('service usage notice', () => {
  it('states the audience region and provides the reference notice sections without claiming an operating location', () => {
    render(<DefaultAboutContent />)
    expect(
      screen.getByRole('heading', { name: 'Service usage notice' })
    ).toBeInTheDocument()
    expect(
      screen.getByText(/This platform serves only users outside mainland China/)
    ).toBeInTheDocument()
    for (const name of [
      'Permitted use',
      'Data and security',
      'Account security',
      'Abuse handling',
    ]) {
      expect(screen.getByRole('heading', { name })).toBeInTheDocument()
    }
    expect(
      screen.queryByText(/operated in|individually operated/)
    ).not.toBeInTheDocument()
    expect(
      screen
        .getAllByRole('separator')
        .every((element) =>
          element.classList.contains('max-w-[min(36rem,80%)]')
        )
    ).toBe(true)
  })
})
