import { CommonModule, NgClass } from '@angular/common';
import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { catchError, finalize, tap } from 'rxjs/operators';
import { of } from 'rxjs';

import {
  LlCustomerRequestDTO,
  LlSolicitudesClientesService,
} from './ll-solicitudes-clientes.service';
import {
  LlUsuarioDTO,
  LlUsuariosService,
} from '../usuarios/ll-usuarios.service';

@Component({
  selector: 'app-ll-solicitudes-list',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, NgClass],
  templateUrl: './ll-solicitudes-list.component.html',
  styleUrl: './ll-solicitudes-list.component.scss',
})
export class LlSolicitudesListComponent {
  private readonly solicitudesService = inject(LlSolicitudesClientesService);
  private readonly usuariosService = inject(LlUsuariosService);
  private readonly destroyRef = inject(DestroyRef);

  readonly solicitudes = signal<LlCustomerRequestDTO[]>([]);
  readonly empleados = signal<LlUsuarioDTO[]>([]);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);

  readonly searchTerm = signal('');
  readonly statusFilter = signal<string>('');
  readonly employeeFilter = signal<string>('');

  readonly updatingId = signal<string | null>(null);

  readonly totalSolicitudes = computed(
    () => this.solicitudes().length,
  );

  readonly statusOptions: { value: string; label: string }[] = [
    { value: 'pendiente', label: 'Pendiente' },
    { value: 'vista', label: 'Vista' },
    { value: 'atendida', label: 'Atendida' },
  ];

  readonly editingAgreement = signal<LlCustomerRequestDTO | null>(null);
  readonly agreementDraft = signal('');
  readonly savingAgreement = signal(false);
  readonly agreementError = signal<string | null>(null);

  constructor() {
    this.loadSolicitudes();
    this.loadEmpleados();
  }

  onSearchChange(value: string): void {
    const sanitized = value.trimStart();
    this.searchTerm.set(sanitized);
    this.loadSolicitudes();
  }

  onStatusFilterChange(value: string): void {
    this.statusFilter.set(value);
    this.loadSolicitudes();
  }

  onEmployeeFilterChange(value: string): void {
    this.employeeFilter.set(value);
    this.loadSolicitudes();
  }

  retry(): void {
    this.loadSolicitudes();
  }

  formatStatus(status: string | null | undefined): string {
    const s = (status || '').toLowerCase();
    switch (s) {
      case 'pending':
      case 'pendiente':
        return 'Pendiente';
      case 'viewed':
      case 'vista':
        return 'Vista';
      case 'handled':
      case 'atendida':
        return 'Atendida';
      default:
        return status || 'Desconocido';
    }
  }

  formatRequestType(requestType: string | null | undefined): string {
    const t = (requestType || '').toLowerCase();
    switch (t) {
      case 'distribuidor':
        return 'Quiero ser distribuidor';
      case 'cliente':
        return 'Quiero hacerme cliente';
      case 'cotizacion':
        return 'Necesito cotización';
      case 'beneficios':
        return 'Información de beneficios';
      default:
        return requestType || 'Otro';
    }
  }

  getEmpleadoNombre(id: string | null | undefined): string {
    if (!id) {
      return 'Sin asignar';
    }
    const empleado = this.empleados().find((e) => e.id === id);
    if (!empleado) {
      return `Empleado #${id}`;
    }
    const nombre = (empleado.name || '').trim();
    return nombre || empleado.email;
  }

  onChangeStatus(solicitud: LlCustomerRequestDTO, newStatus: string): void {
    if (!solicitud.id || !newStatus || solicitud.status === newStatus) {
      return;
    }

    this.updatingId.set(solicitud.id);

    this.solicitudesService
      .update(solicitud.id, { status: newStatus })
      .pipe(
        tap((updated) => {
          this.solicitudes.set(
            this.solicitudes().map((item) =>
              item.id === updated.id
                ? { ...item, status: updated.status, attendedAt: updated.attendedAt }
                : item,
            ),
          );
        }),
        catchError((err) => {
          console.error('Error al cambiar estado de solicitud', err);
          this.error.set('No se pudo cambiar el estado de la solicitud.');
          return of(solicitud);
        }),
        finalize(() => this.updatingId.set(null)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe();
  }

  openAgreementModal(request: LlCustomerRequestDTO): void {
    if (
      !request.employeeId ||
      request.status === 'handled' ||
      request.status === 'atendida'
    ) {
      return;
    }
    this.editingAgreement.set(request);
    this.agreementDraft.set(request.agreement || '');
    this.agreementError.set(null);
  }

  closeAgreementModal(): void {
    if (this.savingAgreement()) {
      return;
    }
    this.editingAgreement.set(null);
    this.agreementDraft.set('');
    this.agreementError.set(null);
  }

  onAgreementDraftChange(value: string): void {
    this.agreementDraft.set(value);
  }

  saveAgreement(): void {
    const current = this.editingAgreement();
    if (!current) {
      return;
    }

    const agreement = this.agreementDraft().trim();

    if (!current.employeeId) {
      this.agreementError.set(
        'Debes asignar un empleado antes de registrar un acuerdo.',
      );
      return;
    }

    if (current.status === 'handled' || current.status === 'atendida') {
      this.agreementError.set(
        'No es posible editar el acuerdo de una solicitud atendida.',
      );
      return;
    }

    this.savingAgreement.set(true);
    this.agreementError.set(null);

    this.solicitudesService
      .update(current.id, { agreement })
      .pipe(
        tap((updated) => {
          this.solicitudes.set(
            this.solicitudes().map((item) =>
              item.id === updated.id
                ? {
                    ...item,
                    agreement: updated.agreement,
                    attendedAt: updated.attendedAt,
                  }
                : item,
            ),
          );
          this.editingAgreement.set(null);
          this.agreementDraft.set('');
        }),
        catchError((err) => {
          console.error('Error al guardar el acuerdo de la solicitud', err);
          this.agreementError.set(
            'No se pudo guardar el acuerdo de la solicitud.',
          );
          return of(current);
        }),
        finalize(() => this.savingAgreement.set(false)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe();
  }

  onAssignEmpleado(
    solicitud: LlCustomerRequestDTO,
    empleadoId: string,
  ): void {
    if (!solicitud.id) {
      return;
    }

    const normalizedId = empleadoId && empleadoId.trim() ? empleadoId.trim() : null;

    const payload: any = {
      employeeId: normalizedId,
    };

    const currentStatus = (solicitud.status || '').toLowerCase();
    if (normalizedId && (currentStatus === 'pending' || currentStatus === 'pendiente')) {
      payload.status = 'vista';
    } else if (!normalizedId && (currentStatus === 'vista' || currentStatus === 'viewed')) {
      payload.status = 'pendiente';
    }

    this.updatingId.set(solicitud.id);

    this.solicitudesService
      .update(solicitud.id, payload)
      .pipe(
        tap((updated) => {
          this.solicitudes.set(
            this.solicitudes().map((item) =>
              item.id === updated.id
                ? {
                    ...item,
                    employeeId: updated.employeeId ?? null,
                    status: updated.status,
                    attendedAt: updated.attendedAt,
                  }
                : item,
            ),
          );
        }),
        catchError((err) => {
          console.error('Error al asignar empleado a solicitud', err);
          this.error.set('No se pudo asignar el empleado a la solicitud.');
          return of(solicitud);
        }),
        finalize(() => this.updatingId.set(null)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe();
  }

  private loadSolicitudes(): void {
    this.loading.set(true);
    this.error.set(null);

    const search = this.searchTerm();
    const status = this.statusFilter();
    const employeeId = this.employeeFilter();

    this.solicitudesService
      .list(search, status || undefined, employeeId || undefined, 50, 0)
      .pipe(
        tap((items) => this.solicitudes.set(items)),
        catchError((err) => {
          console.error('Error al cargar solicitudes de clientes', err);
          this.error.set('No se pudieron cargar las solicitudes.');
          this.solicitudes.set([]);
          return of<LlCustomerRequestDTO[]>([]);
        }),
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe();
  }

  private loadEmpleados(): void {
    this.usuariosService
      .list('', 'employee', 100, 0)
      .pipe(
        tap((items) => {
          this.empleados.set(items);
        }),
        catchError((err) => {
          console.error('Error al cargar empleados para solicitudes', err);
          this.empleados.set([]);
          return of<LlUsuarioDTO[]>([]);
        }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe();
  }
}
