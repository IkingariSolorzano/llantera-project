import { Component, DestroyRef, inject, signal, computed } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { NgClass } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { tap, catchError, finalize } from 'rxjs/operators';
import { of } from 'rxjs';

import { LlUsuariosService, LlUsuarioDTO } from './ll-usuarios.service';
import { LlEmpresasService, LlEmpresaDTO } from '../empresas/ll-empresas.service';

@Component({
  selector: 'app-ll-usuarios-list',
  standalone: true,
  imports: [RouterLink, FormsModule, NgClass],
  templateUrl: './ll-usuarios-list.component.html',
  styleUrl: './ll-usuarios-list.component.scss',
})
export class LlUsuariosListComponent {
  private readonly usuariosService = inject(LlUsuariosService);
  private readonly router = inject(Router);
  private readonly empresasService = inject(LlEmpresasService);
  private readonly destroyRef = inject(DestroyRef);

  readonly usuarios = signal<LlUsuarioDTO[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly searchTerm = signal('');
  readonly roleFilter = signal<string>('');
  readonly usuariosCount = computed(() => this.usuarios().length);
  readonly deletingId = signal<string | null>(null);
  readonly companyOptions = signal<LlEmpresaDTO[]>([]);

  constructor() {
    this.loadUsuarios();
    this.loadCompanies();
  }

  onSearchChange(value: string): void {
    const sanitized = value.trimStart();
    this.searchTerm.set(sanitized);
    this.loadUsuarios(sanitized);
  }

  onRoleFilterChange(value: string): void {
    this.roleFilter.set(value);
    this.loadUsuarios(this.searchTerm());
  }

  retry(): void {
    this.loadUsuarios(this.searchTerm());
  }

  formatNombre(usuario: LlUsuarioDTO): string {
    const fallback = [usuario.firstName, usuario.firstLastName, usuario.secondLastName]
      .filter((part) => !!part?.trim())
      .join(' ')
      .trim();

    return (usuario.name ?? '').trim() || fallback || usuario.email;
  }

  onViewUsuario(id: string): void {
    this.router.navigate(['/dashboard/usuarios', id, 'detalle']);
  }

  onEditUsuario(id: string): void {
    this.router.navigate(['/dashboard/usuarios', id, 'editar']);
  }

  onDeleteUsuario(usuario: LlUsuarioDTO): void {
    if (this.deletingId()) {
      return;
    }
    const confirmed = confirm(`¿Eliminar al usuario "${this.formatNombre(usuario)}"?`);
    if (!confirmed) {
      return;
    }

    this.deletingId.set(usuario.id);
    this.usuariosService
      .delete(usuario.id)
      .pipe(
        finalize(() => this.deletingId.set(null)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: () => this.loadUsuarios(this.searchTerm()),
        error: () => {
          this.error.set('No se pudo eliminar el usuario.');
        },
      });
  }

  getCompanyLabel(companyId?: number): string {
    if (!companyId) {
      return 'Sin asignar';
    }
    const company = this.companyOptions().find((c) => c.id === companyId);
    if (company) {
      return company.socialReason;
    }
    return `Empresa #${companyId}`;
  }

  formatRole(usuario: LlUsuarioDTO | null | undefined): string {
    const role = (usuario?.role || 'customer').toLowerCase();
    switch (role) {
      case 'admin':
        return 'Administrador';
      case 'employee':
        return 'Empleado';
      default:
        return 'Cliente';
    }
  }

  formatNivel(usuario: LlUsuarioDTO | null | undefined): string {
    const level = (usuario?.level || 'public').toLowerCase();
    switch (level) {
      case 'public':
        return 'Público';
      case 'empresa':
        return 'Empresa';
      case 'distribuidor':
        return 'Distribuidor';
      case 'mayorista':
        return 'Mayorista';
      case 'silver':
        return 'Plata';
      case 'gold':
        return 'Oro';
      case 'platinum':
        return 'Platino';
      default:
        return usuario?.level || 'public';
    }
  }

  private loadUsuarios(search = ''): void {
    this.loading.set(true);
    this.error.set(null);

    const role = this.roleFilter();

    this.usuariosService
      .list(search, role || undefined)
      .pipe(
        tap((items) => this.usuarios.set(items)),
        catchError((err) => {
          console.error('Error al cargar usuarios', err);
          this.error.set('No se pudieron cargar los usuarios.');
          return of([]);
        }),
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  private loadCompanies(): void {
    this.empresasService
      .list('', 100, 0)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (companies) => this.companyOptions.set(companies),
        error: (err) => {
          console.error('Error al cargar empresas', err);
          this.companyOptions.set([]);
        },
      });
  }
}
