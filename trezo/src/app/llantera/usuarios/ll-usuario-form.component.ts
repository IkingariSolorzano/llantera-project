import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { Router, RouterLink, ActivatedRoute } from '@angular/router';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs/operators';
import { CommonModule } from '@angular/common';

import {
  LlUsuariosService,
  LlUsuarioDTO,
  LlUsuarioBasePayload,
} from './ll-usuarios.service';
import {
  LlEmpresasService,
  LlEmpresaDTO,
} from '../empresas/ll-empresas.service';
import { LlPriceLevelsService, LlPriceLevelDTO } from './ll-price-levels.service';

interface PriceLevelOption {
  id: number;
  label: string;
}

@Component({
  selector: 'app-ll-usuario-form',
  standalone: true,
  imports: [RouterLink, ReactiveFormsModule, CommonModule],
  templateUrl: './ll-usuario-form.component.html',
  styleUrl: './ll-usuario-form.component.scss',
})
export class LlUsuarioFormComponent {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly usuariosService = inject(LlUsuariosService);
  private readonly empresasService = inject(LlEmpresasService);
  private readonly priceLevelsService = inject(LlPriceLevelsService);
  private readonly fb = inject(FormBuilder);
  private readonly destroyRef = inject(DestroyRef);

  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal<string | null>(null);
  readonly isEditMode = signal(false);
  readonly usuarioId = signal<string | null>(null);
  readonly companyOptions = signal<LlEmpresaDTO[]>([]);
  readonly companiesLoading = signal(false);
  readonly companiesError = signal<string | null>(null);
  readonly priceLevelOptions = signal<PriceLevelOption[]>([]);
  readonly priceLevelsLoading = signal(false);
  readonly priceLevelsError = signal<string | null>(null);
  readonly showConfirmModal = signal(false);
  readonly confirmText = signal('');
  private pendingPayload: LlUsuarioBasePayload | null = null;

  readonly pageTitle = computed(() =>
    this.isEditMode() ? 'Editar usuario / cliente' : 'Nuevo usuario / cliente'
  );
  readonly submitLabel = computed(() =>
    this.isEditMode() ? 'Guardar cambios' : 'Crear usuario'
  );

  readonly form = this.fb.group({
    email: ['', [Validators.required, Validators.email]],
    firstName: ['', Validators.required],
    firstLastName: ['', Validators.required],
    secondLastName: [''],
    phone: [''],
    role: ['customer', Validators.required],
    active: [true],
    companyId: [''],
    priceLevelId: ['', Validators.required],
    profileImageUrl: [''],
    password: [''],
    addressStreet: [''],
    addressNumber: [''],
    addressNeighborhood: [''],
    addressPostalCode: [''],
    jobTitle: [''],
  });

  constructor() {
    this.loadCompanies();
    this.loadPriceLevels();
    this.route.paramMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
      const id = params.get('id');
      if (id) {
        this.isEditMode.set(true);
        this.usuarioId.set(id);
        this.configurePasswordControl(true);
        this.loadUsuario(id);
      } else {
        this.isEditMode.set(false);
        this.configurePasswordControl(false);
      }
    });
  }

  confirmSave(): void {
    const payload = this.pendingPayload ?? this.buildBasePayload();
    this.pendingPayload = null;
    this.showConfirmModal.set(false);
    this.executeSave(payload);
  }

  closeConfirmModal(): void {
    this.showConfirmModal.set(false);
  }

  onSubmit(): void {
    this.error.set(null);
    this.success.set(null);
    this.form.markAllAsTouched();

    if (this.form.invalid) {
      this.error.set('Por favor revisa los campos marcados en rojo.');
      this.logInvalidControls();
      return;
    }

    if (!this.isEditMode()) {
      const password = this.form.value.password?.trim();
      if (!password) {
        this.error.set('La contraseña es obligatoria.');
        return;
      }
    }

    const basePayload = this.buildBasePayload();
    const message = this.buildConfirmationMessage(basePayload);
    this.pendingPayload = basePayload;
    this.confirmText.set(message);
    this.showConfirmModal.set(true);
  }

  onCancel(): void {
    this.navigateToList();
  }

  isInvalid(controlName: string): boolean {
    const control = this.form.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  private navigateToList(): void {
    this.router.navigate(['/dashboard/usuarios']);
  }

  private configurePasswordControl(isEdit: boolean): void {
    const control = this.form.controls.password;
    if (isEdit) {
      control.clearValidators();
      control.enable({ emitEvent: false });
      control.setValue('');
    } else {
      control.setValidators([Validators.required, Validators.minLength(8)]);
      control.enable({ emitEvent: false });
      control.setValue('');
    }
    control.updateValueAndValidity();
  }

  private loadUsuario(id: string): void {
    this.loading.set(true);
    this.error.set(null);

    this.usuariosService
      .getById(id)
      .pipe(
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: (usuario) => this.populateForm(usuario),
        error: () => {
          this.error.set('No se pudo cargar el usuario solicitado.');
        },
      });
  }

  private loadCompanies(): void {
    this.companiesLoading.set(true);
    this.companiesError.set(null);
    const companyCtrl = this.form.controls.companyId;
    companyCtrl.disable({ emitEvent: false });

    this.empresasService
      .list('', 100, 0)
      .pipe(
        finalize(() => {
          this.companiesLoading.set(false);
          companyCtrl.enable({ emitEvent: false });
        }),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: (companies) => {
          this.companyOptions.set(companies);
        },
        error: () => {
          this.companyOptions.set([]);
          this.companiesError.set('No se pudieron cargar las empresas.');
        },
      });
  }

  private loadPriceLevels(): void {
    this.priceLevelsLoading.set(true);
    this.priceLevelsError.set(null);

    this.priceLevelsService
      .list(100, 0)
      .pipe(
        finalize(() => {
          this.priceLevelsLoading.set(false);
        }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: (response) => {
          const options: PriceLevelOption[] = response.items.map((lvl) => ({
            id: lvl.id,
            label: lvl.name,
          }));
          this.priceLevelOptions.set(options);
        },
        error: () => {
          this.priceLevelOptions.set([]);
          this.priceLevelsError.set('No se pudieron cargar los niveles de precio.');
        },
      });
  }

  private populateForm(usuario: LlUsuarioDTO): void {
    this.form.patchValue({
      email: usuario.email,
      firstName: usuario.firstName,
      firstLastName: usuario.firstLastName,
      secondLastName: usuario.secondLastName ?? '',
      phone: this.normalizePhone(usuario.phone),
      role: usuario.role ?? 'customer',
      active: usuario.active,
      companyId: usuario.companyId ? usuario.companyId.toString() : '',
      priceLevelId: usuario.priceLevelId ? usuario.priceLevelId.toString() : '',
      profileImageUrl: usuario.profileImageUrl ?? '',
      addressStreet: usuario.addressStreet ?? '',
      addressNumber: usuario.addressNumber ?? '',
      addressNeighborhood: usuario.addressNeighborhood ?? '',
      addressPostalCode: usuario.addressPostalCode ?? '',
      jobTitle: usuario.jobTitle ?? '',
    });
  }

  private buildBasePayload(): LlUsuarioBasePayload {
    const raw = this.form.value;

    const normalizedPhone = this.normalizePhone(raw.phone);

    return {
      email: raw.email?.trim() || '',
      firstName: raw.firstName?.trim() || '',
      firstLastName: raw.firstLastName?.trim() || '',
      secondLastName: raw.secondLastName?.trim() || undefined,
      phone: normalizedPhone || undefined,
      addressStreet: raw.addressStreet?.trim() || undefined,
      addressNumber: raw.addressNumber?.trim() || undefined,
      addressNeighborhood: raw.addressNeighborhood?.trim() || undefined,
      addressPostalCode: raw.addressPostalCode?.trim() || undefined,
      jobTitle: raw.jobTitle?.trim() || undefined,
      active: raw.active ?? true,
      companyId: this.parseOptionalNumber(raw.companyId),
      profileImageUrl: raw.profileImageUrl?.trim() || undefined,
      role: raw.role || 'customer',
      priceLevelId: this.parseOptionalNumber(raw.priceLevelId),
      password: raw.password?.trim() || undefined,
    };
  }

  private parseOptionalNumber(value?: string | null): number | null | undefined {
    if (value === null || value === undefined) {
      return null;
    }
    const sanitized = value.toString().trim();
    if (!sanitized) {
      return null;
    }
    const parsed = Number(sanitized);
    return Number.isNaN(parsed) ? null : parsed;
  }

  private normalizePhone(value?: string | null): string {
    if (!value) {
      return '';
    }
    const digits = value.toString().replace(/\D+/g, '');
    return digits.slice(0, 10);
  }

  private executeSave(basePayload: LlUsuarioBasePayload): void {
    let request$;

    if (this.isEditMode()) {
      const id = this.usuarioId();
      if (!id) {
        this.error.set('Identificador del usuario no disponible.');
        return;
      }
      request$ = this.usuariosService.update(id, basePayload);
    } else {
      const password = this.form.value.password?.trim() || '';
      request$ = this.usuariosService.create({
        ...basePayload,
        password,
      });
    }

    this.saving.set(true);
    request$
      .pipe(
        finalize(() => this.saving.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: () => {
          const message = this.isEditMode()
            ? 'Usuario actualizado correctamente.'
            : 'Usuario creado correctamente.';
          this.success.set(message);
          this.navigateToList();
        },
        error: (err) => {
          this.handleSaveError(err);
        },
      });
  }

  private handleSaveError(err: any): void {
    let message = 'No se pudo guardar la información. Inténtalo más tarde.';
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
  }

  private buildConfirmationMessage(payload: LlUsuarioBasePayload): string {
    const fullNameParts = [
      payload.firstName,
      payload.firstLastName,
      payload.secondLastName,
    ].filter((part) => !!part && !!part.trim());
    const fullName = fullNameParts.join(' ').trim() || '(sin nombre)';
    const email = payload.email || '(sin correo)';
    const phone = payload.phone?.trim() || 'Sin teléfono';
    const companyLabel = this.getCompanyLabelForConfirmation(payload.companyId);
    const roleLabel = this.formatRoleLabel(payload.role);
    const priceLevelLabel = this.getPriceLevelLabel(payload.priceLevelId);
    const statusLabel = payload.active ? 'Activo' : 'Inactivo';

    const header = this.isEditMode()
      ? 'Se guardarán los siguientes cambios del usuario:\n\n'
      : 'Se creará un nuevo usuario / cliente con los siguientes datos:\n\n';

    const lines = [
      `- Correo: ${email}`,
      `- Nombre: ${fullName}`,
      `- Teléfono: ${phone}`,
      `- Empresa: ${companyLabel}`,
      `- Rol: ${roleLabel}`,
      `- Nivel de precio: ${priceLevelLabel}`,
      `- Estado: ${statusLabel}`,
    ];

    const footer = '\n¿Deseas continuar y guardar los cambios?';

    return header + lines.join('\n') + footer;
  }

  private getCompanyLabelForConfirmation(
    companyId?: number | null
  ): string {
    if (!companyId) {
      return 'Sin asignar';
    }
    const company = this.companyOptions().find((c) => c.id === companyId);
    if (company) {
      return company.socialReason;
    }
    return `Empresa #${companyId}`;
  }

  private getPriceLevelLabel(priceLevelId?: number | null): string {
    if (!priceLevelId) {
      return 'Sin asignar';
    }
    const option = this.priceLevelOptions().find((opt) => opt.id === priceLevelId);
    return option?.label ?? `Nivel #${priceLevelId}`;
  }

  private formatRoleLabel(role: string | null | undefined): string {
    const normalized = (role || 'customer').toLowerCase();
    switch (normalized) {
      case 'admin':
        return 'Administrador';
      case 'employee':
        return 'Empleado';
      default:
        return 'Cliente';
    }
  }

  private logInvalidControls(): void {
    const invalid: string[] = [];
    const controls = this.form.controls;
    for (const key of Object.keys(controls)) {
      const control = controls[key as keyof typeof controls];
      if (control.invalid) {
        invalid.push(`${key}: ${JSON.stringify(control.errors)}`);
      }
    }
    // Logging solo para depuración en desarrollo
    // eslint-disable-next-line no-console
    console.warn('[LlUsuarioForm] Controles inválidos:', invalid);
  }
}
