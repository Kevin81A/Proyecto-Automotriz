import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';

import { ProtectedRoute } from './ProtectedRoute';
import * as SessionContext from '../shared/SessionContext';

describe('ProtectedRoute', () => {
  it('redirects to /login when user is not authenticated', () => {
    vi.spyOn(SessionContext, 'useSession').mockReturnValue({
      session: null,
      signIn: vi.fn(),
      signOut: vi.fn(),
      isAdministrator: false,
    });

    render(
      <MemoryRouter initialEntries={['/customers']}>
        <Routes>
          <Route path="/login" element={<div>Login Screen</div>} />
          <Route
            path="/customers"
            element={
              <ProtectedRoute requiredRole="ADMINISTRATOR">
                <div>Customers Screen</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText('Login Screen')).toBeInTheDocument();
    expect(screen.queryByText('Customers Screen')).not.toBeInTheDocument();
  });

  it('redirects technician to /dashboard when route requires ADMINISTRATOR', () => {
    vi.spyOn(SessionContext, 'useSession').mockReturnValue({
      session: {
        token: 'tech-token',
        expiresAt: '2026-09-20T00:00:00Z',
        userId: 'tech-1',
        username: 'jperez',
        fullName: 'Juan Perez',
        role: 'TECHNICIAN',
      },
      signIn: vi.fn(),
      signOut: vi.fn(),
      isAdministrator: false,
    });

    render(
      <MemoryRouter initialEntries={['/customers']}>
        <Routes>
          <Route path="/dashboard" element={<div>Dashboard Screen</div>} />
          <Route
            path="/customers"
            element={
              <ProtectedRoute requiredRole="ADMINISTRATOR">
                <div>Customers Screen</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText('Dashboard Screen')).toBeInTheDocument();
    expect(screen.queryByText('Customers Screen')).not.toBeInTheDocument();
  });

  it('renders children when user has the required ADMINISTRATOR role', () => {
    vi.spyOn(SessionContext, 'useSession').mockReturnValue({
      session: {
        token: 'admin-token',
        expiresAt: '2026-09-20T00:00:00Z',
        userId: 'admin-1',
        username: 'admin',
        fullName: 'Administrador',
        role: 'ADMINISTRATOR',
      },
      signIn: vi.fn(),
      signOut: vi.fn(),
      isAdministrator: true,
    });

    render(
      <MemoryRouter initialEntries={['/customers']}>
        <Routes>
          <Route path="/dashboard" element={<div>Dashboard Screen</div>} />
          <Route
            path="/customers"
            element={
              <ProtectedRoute requiredRole="ADMINISTRATOR">
                <div>Customers Screen</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText('Customers Screen')).toBeInTheDocument();
  });
});
