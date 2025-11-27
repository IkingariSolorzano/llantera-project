import { inject, Injectable, computed, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { Router } from '@angular/router';

import { API_BASE_URL } from '../../core/config/api.config';

export interface LlAuthUser {
  id: string;
  email: string;
  name: string;
  role: 'customer' | 'employee' | 'admin' | string;
  level: string;
  priceLevelId?: number | null;
}

export interface LlAuthLoginResponse {
  token: string;
  expiresAt: string; // ISO string, servidor: ahora + 6 días
  user: LlAuthUser;
}

const AUTH_TOKEN_KEY = 'll_auth_token';
const AUTH_USER_KEY = 'll_auth_user';
const AUTH_EXPIRES_AT_KEY = 'll_auth_expiresAt';

@Injectable({ providedIn: 'root' })
export class LlAuthService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);
  private readonly router = inject(Router);

  readonly user = signal<LlAuthUser | null>(LlAuthService.loadUserFromStorage());
  readonly token = signal<string | null>(LlAuthService.loadTokenFromStorage());

  readonly isAuthenticated = computed(() => !!this.token());

  constructor() {
    this.validateStoredSession();
  }

  login(email: string, password: string): Observable<LlAuthLoginResponse> {
    const payload = {
      email: email.trim().toLowerCase(),
      password: password.trim(),
    };

    return this.http
      .post<LlAuthLoginResponse>(`${this.baseUrl}/auth/login`, payload)
      .pipe(
        tap((response) => {
          this.setSession(response);
          this.redirectAfterLogin(response.user);
        }),
      );
  }

  logout(): void {
    this.token.set(null);
    this.user.set(null);
    localStorage.removeItem(AUTH_TOKEN_KEY);
    localStorage.removeItem(AUTH_USER_KEY);
    localStorage.removeItem(AUTH_EXPIRES_AT_KEY);
    this.router.navigate(['/login']);
  }

  private setSession(response: LlAuthLoginResponse): void {
    const { token, user, expiresAt } = response;

    this.token.set(token);
    this.user.set(user);

    try {
      localStorage.setItem(AUTH_TOKEN_KEY, token);
      localStorage.setItem(AUTH_USER_KEY, JSON.stringify(user));
      if (expiresAt) {
        localStorage.setItem(AUTH_EXPIRES_AT_KEY, expiresAt);
      }
    } catch {
      // Ignorar errores de almacenamiento (modo privado, etc.)
    }
  }

  private static loadTokenFromStorage(): string | null {
    try {
      const raw = localStorage.getItem(AUTH_TOKEN_KEY);
      return raw && raw.trim() ? raw : null;
    } catch {
      return null;
    }
  }

  private static loadUserFromStorage(): LlAuthUser | null {
    try {
      const raw = localStorage.getItem(AUTH_USER_KEY);
      if (!raw) {
        return null;
      }
      const parsed = JSON.parse(raw) as LlAuthUser;
      if (!parsed || !parsed.id || !parsed.email) {
        return null;
      }
      return parsed;
    } catch {
      return null;
    }
  }

  private redirectAfterLogin(user: LlAuthUser): void {
    const role = (user.role || 'customer').toString().toLowerCase();

    if (role === 'customer') {
      this.router.navigate(['/cliente']);
      return;
    }

    // Empleado y admin por ahora van al dashboard principal
    this.router.navigate(['/dashboard']);
  }

  private validateStoredSession(): void {
    let token: string | null;
    let user: LlAuthUser | null;
    try {
      token = this.token();
      user = this.user();
    } catch {
      token = null;
      user = null;
    }

    if (!token || !user) {
      this.token.set(null);
      this.user.set(null);
      try {
        localStorage.removeItem(AUTH_TOKEN_KEY);
        localStorage.removeItem(AUTH_USER_KEY);
        localStorage.removeItem(AUTH_EXPIRES_AT_KEY);
      } catch {}
      return;
    }

    let rawExpires: string | null = null;
    try {
      rawExpires = localStorage.getItem(AUTH_EXPIRES_AT_KEY);
    } catch {
      rawExpires = null;
    }
    if (!rawExpires) {
      this.logout();
      return;
    }
    const ts = Date.parse(rawExpires);
    if (Number.isNaN(ts) || ts <= Date.now()) {
      this.logout();
    }
  }

  goToMyAccount(): void {
    const current = this.user();
    if (!current) {
      this.router.navigate(['/login']);
      return;
    }
    this.redirectAfterLogin(current);
  }
}
