import { inject } from '@angular/core';
import { CanActivateFn, Router, RouterStateSnapshot, UrlTree } from '@angular/router';

import { LlAuthService } from './ll-auth.service';

function buildRedirectToLogin(router: Router, state: RouterStateSnapshot): UrlTree {
  return router.createUrlTree(['/login'], {
    queryParams: { returnUrl: state.url },
  });
}

export const authGuard: CanActivateFn = (_route, state) => {
  const auth = inject(LlAuthService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }

  return buildRedirectToLogin(router, state);
};

export const customerGuard: CanActivateFn = (_route, state) => {
  const auth = inject(LlAuthService);
  const router = inject(Router);

  if (!auth.isAuthenticated()) {
    return buildRedirectToLogin(router, state);
  }

  const user = auth.user();
  const role = (user?.role || 'customer').toString().toLowerCase();

  if (role === 'customer') {
    return true;
  }

  // Empleados y administradores deben usar el panel de administración.
  return router.createUrlTree(['/dashboard']);
};

export const adminEmployeeGuard: CanActivateFn = (_route, state) => {
  const auth = inject(LlAuthService);
  const router = inject(Router);

  if (!auth.isAuthenticated()) {
    return buildRedirectToLogin(router, state);
  }

  const user = auth.user();
  const role = (user?.role || '').toString().toLowerCase();

  if (role === 'admin' || role === 'employee') {
    return true;
  }

  // Clientes autenticados deben ir a su panel de cliente.
  return router.createUrlTree(['/cliente']);
};

export const guestGuard: CanActivateFn = (_route, _state) => {
  const auth = inject(LlAuthService);
  const router = inject(Router);

  if (!auth.isAuthenticated()) {
    return true;
  }

  const user = auth.user();
  const role = (user?.role || 'customer').toString().toLowerCase();

  if (role === 'customer') {
    return router.createUrlTree(['/cliente']);
  }

  return router.createUrlTree(['/dashboard']);
};
