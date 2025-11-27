import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormArray, FormBuilder, ReactiveFormsModule, Validators, FormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs/operators';

import { LlEmpresasService, LlEmpresaBasePayload, LlEmpresaDTO } from './ll-empresas.service';
import { LlUsuariosService, LlUsuarioDTO } from '../usuarios/ll-usuarios.service';

@Component({
  selector: 'app-ll-empresa-form',
  standalone: true,
  imports: [RouterLink, ReactiveFormsModule, CommonModule, FormsModule],
  templateUrl: './ll-empresa-form.component.html',
  styleUrl: './ll-empresa-form.component.scss',
})
export class LlEmpresaFormComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);
  private readonly empresasService = inject(LlEmpresasService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly usuariosService = inject(LlUsuariosService);

  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal<string | null>(null);
  readonly isEditMode = signal(false);
  readonly empresaId = signal<number | null>(null);

  private readonly rfcPattern = /^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$/i;

  readonly contactUsers = signal<LlUsuarioDTO[]>([]);
  readonly contactLoading = signal(false);
  readonly contactError = signal<string | null>(null);
  readonly contactSearchTerm = signal('');
  readonly filteredContactUsers = computed(() => {
    const term = this.contactSearchTerm().trim().toLowerCase();
    const all = this.contactUsers();
    if (!term) {
      return all;
    }
    return all.filter((u) => {
      const parts = [
        u.name,
        u.firstName,
        u.firstLastName,
        u.secondLastName,
        u.email,
      ]
        .filter((v) => !!v)
        .join(' ') // nombre completo + correo
        .toLowerCase();
      return parts.includes(term);
    });
  });

  readonly showConfirmModal = signal(false);
  readonly confirmText = signal('');
  private pendingPayload: LlEmpresaBasePayload | null = null;

  readonly pageTitle = computed(() => (this.isEditMode() ? 'Editar empresa' : 'Nueva empresa'));
  readonly submitLabel = computed(() => (this.isEditMode() ? 'Guardar cambios' : 'Crear empresa'));

  readonly form = this.fb.group({
    keyName: ['', Validators.required],
    socialReason: ['', Validators.required],
    rfc: ['', [Validators.minLength(12), Validators.maxLength(13), Validators.pattern(this.rfcPattern)]],
    address: [''],
    emails: this.fb.array([this.fb.control('', Validators.email)]),
    phones: this.fb.array([this.fb.control('', [Validators.pattern(/^\d{10}$/)])]),
    mainContactId: [''],
  });

  get emails(): FormArray {
    return this.form.get('emails') as FormArray;
  }

  get phones(): FormArray {
    return this.form.get('phones') as FormArray;
  }

  constructor() {
    this.loadContactUsers();
    this.route.paramMap.pipe(takeUntilDestroyed(this.destroyRef)).subscribe((params) => {
      const idParam = params.get('id');
      if (idParam) {
        const id = Number(idParam);
        this.isEditMode.set(true);
        this.empresaId.set(id);
        this.loadEmpresa(id);
      } else {
        this.isEditMode.set(false);
      }
    });
  }

  addEmail(): void {
    this.emails.push(this.fb.control('', Validators.email));
  }

  removeEmail(index: number): void {
    if (this.emails.length > 1) {
      this.emails.removeAt(index);
    }
  }

  addPhone(): void {
    this.phones.push(this.fb.control('', [Validators.pattern(/^\d{10}$/)]));
  }

  removePhone(index: number): void {
    if (this.phones.length > 1) {
      this.phones.removeAt(index);
    }
  }

  onSubmit(): void {
    this.error.set(null);
    this.success.set(null);
    this.form.markAllAsTouched();

    if (this.form.invalid) {
      this.error.set('Por favor revisa los campos marcados en rojo.');
      return;
    }

    const payload = this.buildPayload();
    const message = this.buildConfirmationMessage(payload);
    this.pendingPayload = payload;
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

  isEmailInvalid(index: number): boolean {
    const control = this.emails.at(index);
    return control.invalid && (control.dirty || control.touched);
  }

  isPhoneInvalid(index: number): boolean {
    const control = this.phones.at(index);
    return control.invalid && (control.dirty || control.touched);
  }

  private loadEmpresa(id: number): void {
    this.loading.set(true);
    this.error.set(null);

    this.empresasService
      .getById(id)
      .pipe(
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: (empresa) => this.populateForm(empresa),
        error: () => this.error.set('No se pudo cargar la empresa solicitada.'),
      });
  }

  private populateForm(empresa: LlEmpresaDTO): void {
    this.form.patchValue({
      keyName: empresa.keyName,
      socialReason: empresa.socialReason,
      rfc: empresa.rfc?.toUpperCase() ?? '',
      address: empresa.address,
      mainContactId: empresa.mainContactId ?? '',
    });

    this.setDynamicArray(this.emails, empresa.emails, Validators.email);
    this.setDynamicArray(
      this.phones,
      this.normalizePhonesArray(empresa.phones),
      [Validators.pattern(/^\d{10}$/)]
    );
  }

  private setDynamicArray(array: FormArray, values?: (string | null)[], validator?: any): void {
    array.clear();
    const items = values && values.length ? values : [''];
    for (const value of items) {
      array.push(this.fb.control(value, validator));
    }
  }

  private buildPayload(): LlEmpresaBasePayload {
    const raw = this.form.value;

    const trimmedRfc = raw.rfc?.trim().toUpperCase() || '';

    return {
      keyName: raw.keyName?.trim() || '',
      socialReason: raw.socialReason?.trim() || '',
      rfc: trimmedRfc || undefined,
      address: raw.address?.trim() || undefined,
      emails: this.toSanitizedArray(raw.emails),
      phones: this.toNormalizedPhonesArray(raw.phones),
      mainContactId: raw.mainContactId?.trim() || undefined,
    };
  }

  private buildConfirmationMessage(payload: LlEmpresaBasePayload): string {
    const keyName = payload.keyName || '(sin clave)';
    const socialReason = payload.socialReason || '(sin razón social)';
    const rfc = payload.rfc || 'N/A';
    const address = payload.address || 'Sin especificar';
    const emails = payload.emails.length ? payload.emails.join(', ') : 'Sin correos';
    const phones = payload.phones.length ? payload.phones.join(', ') : 'Sin teléfonos';
    const mainContact = this.getContactLabel(payload.mainContactId);

    const header = this.isEditMode()
      ? 'Se guardarán los siguientes cambios de la empresa:\n\n'
      : 'Se creará una nueva empresa con los siguientes datos:\n\n';

    const lines = [
      `- Razón social: ${socialReason}`,
      `- Clave interna: ${keyName}`,
      `- RFC: ${rfc}`,
      `- Dirección: ${address}`,
      `- Correos: ${emails}`,
      `- Teléfonos: ${phones}`,
      `- Contacto principal (ID): ${mainContact}`,
    ];

    const footer = '\n¿Deseas continuar y guardar los cambios?';

    return header + lines.join('\n') + footer;
  }

  private toSanitizedArray(values?: (string | null | undefined)[] | null): string[] {
    if (!values) {
      return [];
    }
    return values
      .map((value) => value?.trim())
      .filter((value): value is string => !!value);
  }

  private toNormalizedPhonesArray(values?: (string | null | undefined)[] | null): string[] {
    if (!values) {
      return [];
    }
    return values
      .map((value) => this.normalizePhone(value))
      .filter((value): value is string => !!value);
  }

  confirmSave(): void {
    const payload = this.pendingPayload ?? this.buildPayload();
    this.pendingPayload = null;
    this.showConfirmModal.set(false);
    this.executeSave(payload);
  }

  closeConfirmModal(): void {
    this.showConfirmModal.set(false);
  }

  onContactSearchChange(value: string): void {
    this.contactSearchTerm.set(value.trim());
  }

  onRfcInput(event: Event): void {
    const input = event.target as HTMLInputElement;
    const upper = input.value.toUpperCase();
    if (upper !== input.value) {
      input.value = upper;
    }
    this.form.get('rfc')?.setValue(upper, { emitEvent: false });
  }

  private executeSave(payload: LlEmpresaBasePayload): void {
    const empresaId = this.empresaId();

    let request$;
    if (this.isEditMode() && empresaId) {
      request$ = this.empresasService.update(empresaId, payload);
    } else {
      request$ = this.empresasService.create(payload);
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
            ? 'Empresa actualizada correctamente.'
            : 'Empresa creada correctamente.';
          this.success.set(message);
          this.navigateToList();
        },
        error: (err) => this.handleSaveError(err),
      });
  }

  private handleSaveError(err: any): void {
    let message = 'No se pudo guardar la empresa. Inténtalo más tarde.';
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

  private loadContactUsers(): void {
    this.contactLoading.set(true);
    this.contactError.set(null);

    this.usuariosService
      .list('', undefined, 100, 0)
      .pipe(
        finalize(() => this.contactLoading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: (users) => this.contactUsers.set(users),
        error: () => {
          this.contactUsers.set([]);
          this.contactError.set('No se pudieron cargar los usuarios.');
        },
      });
  }

  private normalizePhone(value?: string | null): string {
    if (!value) {
      return '';
    }
    const digits = value.toString().replace(/\D+/g, '');
    return digits.slice(0, 10);
  }

  private normalizePhonesArray(values?: (string | null)[] | null): string[] {
    if (!values) {
      return [];
    }
    return values.map((v) => this.normalizePhone(v));
  }

  private getContactLabel(mainContactId?: string | null): string {
    if (!mainContactId) {
      return 'Sin asignar';
    }
    const users = this.contactUsers();
    const match = users.find((u) => u.id === mainContactId);
    if (!match) {
      return `ID #${mainContactId}`;
    }

    const fullNameParts = [
      match.firstName,
      match.firstLastName,
      match.secondLastName,
    ].filter((part) => !!part && !!part.trim());
    const fullName = fullNameParts.join(' ').trim() || match.email;

    return `${fullName} (${match.email})`;
  }

  private navigateToList(): void {
    this.router.navigate(['/dashboard/empresas']);
  }
}
