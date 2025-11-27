import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { tap, catchError, finalize } from 'rxjs/operators';
import { of } from 'rxjs';

import { LlEmpresasService, LlEmpresaDTO } from './ll-empresas.service';

@Component({
  selector: 'app-ll-empresas-list',
  standalone: true,
  imports: [RouterLink, FormsModule],
  templateUrl: './ll-empresas-list.component.html',
  styleUrl: './ll-empresas-list.component.scss',
})
export class LlEmpresasListComponent {
  private readonly empresasService = inject(LlEmpresasService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  readonly empresas = signal<LlEmpresaDTO[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly searchTerm = signal('');
  readonly totalEmpresas = computed(() => this.empresas().length);
  readonly deletingId = signal<number | null>(null);

  constructor() {
    this.loadEmpresas();
  }

  onSearchChange(value: string): void {
    const sanitized = value.trimStart();
    this.searchTerm.set(sanitized);
    this.loadEmpresas(sanitized);
  }

  retry(): void {
    this.loadEmpresas(this.searchTerm());
  }

  onViewEmpresa(id: number): void {
    this.router.navigate(['/dashboard/empresas', id, 'editar']);
  }

  onEditEmpresa(id: number): void {
    this.router.navigate(['/dashboard/empresas', id, 'editar']);
  }

  onDeleteEmpresa(empresa: LlEmpresaDTO): void {
    if (this.deletingId()) {
      return;
    }

    const confirmed = confirm(`¿Eliminar la empresa "${empresa.socialReason}"?`);
    if (!confirmed) {
      return;
    }

    this.deletingId.set(empresa.id);
    this.empresasService
      .delete(empresa.id)
      .pipe(
        finalize(() => this.deletingId.set(null)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: () => this.loadEmpresas(this.searchTerm()),
        error: () => {
          this.error.set('No se pudo eliminar la empresa.');
        },
      });
  }

  private loadEmpresas(search = ''): void {
    this.loading.set(true);
    this.error.set(null);

    this.empresasService
      .list(search)
      .pipe(
        tap((items) => this.empresas.set(items)),
        catchError((err) => {
          console.error('Error al cargar empresas', err);
          this.error.set('No se pudieron cargar las empresas.');
          return of([]);
        }),
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  displayContactList(values: string[]): string {
    if (!values || !values.length) {
      return 'No definido';
    }
    return values.join(', ');
  }
}
