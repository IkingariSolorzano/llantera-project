import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, PLATFORM_ID, computed, inject, signal, OnInit } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { finalize } from 'rxjs/operators';

import { LlAuthService } from './ll-auth.service';
import { ToggleService } from '../../common/header/toggle.service';

@Component({
  selector: 'app-ll-login-page',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './ll-login-page.component.html',
  styleUrl: './ll-login-page.component.scss',
})
export class LlLoginPageComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(LlAuthService);
  private readonly toggleService = inject(ToggleService);
  private readonly platformId = inject(PLATFORM_ID);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly currentYear = new Date().getFullYear();
  readonly isDarkMode = signal(false);

  ngOnInit(): void {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }
    
    // Inicializar tema global usando el mismo servicio que el header
    this.toggleService.initializeTheme();

    const isDark = document.documentElement.classList.contains('dark');
    this.isDarkMode.set(isDark);
  }

  toggleTheme(): void {
    // Usar ToggleService para mantener consistencia y persistencia del tema
    this.toggleService.toggleTheme();

    if (isPlatformBrowser(this.platformId)) {
      const isDark = document.documentElement.classList.contains('dark');
      this.isDarkMode.set(isDark);
    }
  }

  readonly form = this.fb.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(8)]],
  });

  readonly disableForm = computed(() => this.loading());

  onSubmit(): void {
    this.error.set(null);
    this.form.markAllAsTouched();

    if (this.form.invalid) {
      this.error.set('Revisa tu correo y contraseña.');
      return;
    }

    const raw = this.form.value;
    const email = (raw.email ?? '').toString();
    const password = (raw.password ?? '').toString();

    this.loading.set(true);

    this.auth
      .login(email, password)
      .pipe(
        finalize(() => this.loading.set(false)),
      )
      .subscribe({
        next: () => {
          // La redirección se maneja dentro de LlAuthService.redirectAfterLogin
        },
        error: (err) => {
          let message = 'No se pudo iniciar sesión. Intenta de nuevo.';
          if (err && err.error) {
            const backendError = err.error;
            if (typeof backendError === 'string' && backendError.trim()) {
              message = backendError;
            } else if (
              typeof backendError === 'object' &&
              typeof backendError.error === 'string' &&
              backendError.error.trim()
            ) {
              message = backendError.error;
            }
          }
          this.error.set(message);
        },
      });
  }

  isInvalid(controlName: 'email' | 'password'): boolean {
    const control = this.form.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }
}
