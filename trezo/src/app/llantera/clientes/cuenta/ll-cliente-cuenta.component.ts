import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { LlAuthService } from '../../auth/ll-auth.service';

@Component({
  selector: 'app-ll-cliente-cuenta',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './ll-cliente-cuenta.component.html',
})
export class LlClienteCuentaComponent {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(LlAuthService);

  readonly user = this.auth.user;
  readonly loading = signal(false);
  readonly success = signal<string | null>(null);
  readonly error = signal<string | null>(null);
  readonly showPasswordForm = signal(false);

  readonly profileForm = this.fb.group({
    name: [this.user()?.name || '', [Validators.required, Validators.minLength(3)]],
    email: [{ value: this.user()?.email || '', disabled: true }],
    phone: [''],
  });

  readonly passwordForm = this.fb.group({
    currentPassword: ['', [Validators.required, Validators.minLength(8)]],
    newPassword: ['', [Validators.required, Validators.minLength(8)]],
    confirmPassword: ['', [Validators.required]],
  });

  togglePasswordForm(): void {
    this.showPasswordForm.update(v => !v);
    this.passwordForm.reset();
    this.clearMessages();
  }

  clearMessages(): void {
    this.success.set(null);
    this.error.set(null);
  }

  onProfileSubmit(): void {
    this.clearMessages();
    this.profileForm.markAllAsTouched();

    if (this.profileForm.invalid) {
      this.error.set('Por favor corrige los errores del formulario.');
      return;
    }

    this.loading.set(true);

    // TODO: Implementar llamada al backend para actualizar perfil
    setTimeout(() => {
      this.loading.set(false);
      this.success.set('Perfil actualizado correctamente.');
    }, 1000);
  }

  onPasswordSubmit(): void {
    this.clearMessages();
    this.passwordForm.markAllAsTouched();

    if (this.passwordForm.invalid) {
      this.error.set('Por favor corrige los errores del formulario.');
      return;
    }

    const { newPassword, confirmPassword } = this.passwordForm.value;
    if (newPassword !== confirmPassword) {
      this.error.set('Las contraseñas no coinciden.');
      return;
    }

    this.loading.set(true);

    // TODO: Implementar llamada al backend para cambiar contraseña
    setTimeout(() => {
      this.loading.set(false);
      this.success.set('Contraseña actualizada correctamente.');
      this.passwordForm.reset();
      this.showPasswordForm.set(false);
    }, 1000);
  }

  isProfileInvalid(controlName: string): boolean {
    const control = this.profileForm.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  isPasswordInvalid(controlName: string): boolean {
    const control = this.passwordForm.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }
}
