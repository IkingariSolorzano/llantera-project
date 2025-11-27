import { inject } from '@angular/core';
import { HttpInterceptorFn } from '@angular/common/http';

import { LlAuthService } from './ll-auth.service';
import { API_BASE_URL } from '../../core/config/api.config';

/**
 * Interceptor que agrega el header Authorization: Bearer <token>
 * a las peticiones dirigidas al backend, excepto al endpoint de login.
 */
export const llAuthInterceptor: HttpInterceptorFn = (req, next) => {
  const baseUrl = inject(API_BASE_URL);
  const auth = inject(LlAuthService);

  const token = auth.token();

  // Solo interceptar llamadas contra la API propia.
  const isApiCall = typeof baseUrl === 'string' && baseUrl.length > 0 && req.url.startsWith(baseUrl);
  const isLogin = isApiCall && req.url.endsWith('/auth/login');

  if (!isApiCall || !token || isLogin) {
    return next(req);
  }

  const authReq = req.clone({
    setHeaders: {
      Authorization: `Bearer ${token}`,
    },
  });

  return next(authReq);
};
