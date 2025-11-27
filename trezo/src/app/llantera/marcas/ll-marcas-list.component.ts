import { CommonModule, NgClass } from '@angular/common';
import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { catchError, finalize, tap } from 'rxjs/operators';
import { of } from 'rxjs';

import { LlLlantasBrandsService, LlTireBrandDTO } from '../llantas/ll-llantas-brands.service';

@Component({
  selector: 'app-ll-marcas-list',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule, NgClass],
  templateUrl: './ll-marcas-list.component.html',
  styleUrl: './ll-marcas-list.component.scss',
})
export class LlMarcasListComponent {
  private readonly brandsService = inject(LlLlantasBrandsService);
  private readonly destroyRef = inject(DestroyRef);

  readonly allBrands = signal<LlTireBrandDTO[]>([]);
  readonly searchTerm = signal('');
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly deletingId = signal<number | null>(null);

  readonly brands = computed(() => {
    const term = this.searchTerm().trim().toLowerCase();
    const items = this.allBrands();
    if (!term) {
      return items;
    }
    return items.filter((brand) => {
      const name = (brand.nombre ?? '').toLowerCase();
      const aliases = (brand.aliases ?? []).join(' ').toLowerCase();
      return name.includes(term) || aliases.includes(term);
    });
  });

  readonly totalBrands = computed(() => this.brands().length);

  readonly showForm = signal(false);
  readonly editingBrand = signal<LlTireBrandDTO | null>(null);
  readonly formNombre = signal('');
  readonly formAliases = signal<string[]>([]);
  readonly aliasInput = signal('');
  readonly saving = signal(false);
  readonly formError = signal<string | null>(null);

  readonly deleteConfirmBrand = signal<LlTireBrandDTO | null>(null);
  readonly deleteError = signal<string | null>(null);

  constructor() {
    this.loadBrands();
  }

  onSearchChange(value: string): void {
    const sanitized = value.trimStart();
    this.searchTerm.set(sanitized);
  }

  retry(): void {
    this.loadBrands();
  }

  onFormNombreChange(value: string): void {
    this.formNombre.set(value);
  }

  onAliasInputChange(value: string): void {
    this.aliasInput.set(value);
  }

  openCreate(): void {
    this.editingBrand.set(null);
    this.formNombre.set('');
    this.formAliases.set([]);
    this.aliasInput.set('');
    this.formError.set(null);
    this.showForm.set(true);
  }

  openEdit(brand: LlTireBrandDTO): void {
    this.editingBrand.set(brand);
    this.formNombre.set(brand.nombre ?? '');
    this.formAliases.set([...(brand.aliases ?? [])]);
    this.aliasInput.set('');
    this.formError.set(null);
    this.showForm.set(true);
  }

  closeForm(): void {
    if (this.saving()) {
      return;
    }
    this.showForm.set(false);
    this.editingBrand.set(null);
    this.formNombre.set('');
    this.formAliases.set([]);
    this.aliasInput.set('');
    this.formError.set(null);
  }

  addAlias(): void {
    const value = this.aliasInput().trim();
    if (!value) {
      return;
    }
    const current = this.formAliases();
    const exists = current.some((a) => a.toLowerCase() === value.toLowerCase());
    if (exists) {
      this.aliasInput.set('');
      return;
    }
    this.formAliases.set([...current, value]);
    this.aliasInput.set('');
  }

  removeAlias(alias: string): void {
    this.formAliases.set(this.formAliases().filter((a) => a !== alias));
  }

  saveForm(): void {
    const nombre = this.formNombre().trim();
    if (!nombre) {
      this.formError.set('El nombre de la marca es obligatorio.');
      return;
    }

    this.formError.set(null);
    const payload = this.buildPayload();
    const editing = this.editingBrand();

    this.saving.set(true);

    const request$ = editing
      ? this.brandsService.update(editing.id, payload)
      : this.brandsService.create(payload);

    request$
      .pipe(
        tap((saved) => {
          if (!saved) {
            return;
          }
          const items = this.allBrands();
          if (editing) {
            this.allBrands.set(items.map((b) => (b.id === saved.id ? saved : b)));
          } else {
            this.allBrands.set([...items, saved]);
          }
          this.showForm.set(false);
          this.editingBrand.set(null);
          this.formNombre.set('');
          this.formAliases.set([]);
          this.aliasInput.set('');
        }),
        catchError((err) => {
          console.error('Error al guardar marca', err);
          const message =
            (err?.error && (err.error.message || err.error.error)) ||
            'No se pudo guardar la marca.';
          this.formError.set(message);
          return of<LlTireBrandDTO | null>(null);
        }),
        finalize(() => this.saving.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  openDeleteConfirm(brand: LlTireBrandDTO): void {
    this.deleteConfirmBrand.set(brand);
    this.deleteError.set(null);
  }

  cancelDelete(): void {
    if (this.deletingId()) {
      return;
    }
    this.deleteConfirmBrand.set(null);
    this.deleteError.set(null);
  }

  confirmDelete(): void {
    const brand = this.deleteConfirmBrand();
    if (!brand || this.deletingId()) {
      return;
    }

    this.deletingId.set(brand.id);
    this.deleteError.set(null);

    this.brandsService
      .delete(brand.id)
      .pipe(
        tap(() => {
          this.allBrands.set(this.allBrands().filter((b) => b.id !== brand.id));
          this.deleteConfirmBrand.set(null);
        }),
        catchError((err) => {
          console.error('Error al eliminar marca', err);
          const message =
            (err?.error && (err.error.message || err.error.error)) ||
            'No se pudo eliminar la marca.';
          this.deleteError.set(message);
          return of<void>(undefined);
        }),
        finalize(() => this.deletingId.set(null)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  private loadBrands(): void {
    this.loading.set(true);
    this.error.set(null);

    this.brandsService
      .list()
      .pipe(
        tap((items) => this.allBrands.set(items ?? [])),
        catchError((err) => {
          console.error('Error al cargar marcas', err);
          this.error.set('No se pudieron cargar las marcas.');
          this.allBrands.set([]);
          return of<LlTireBrandDTO[]>([]);
        }),
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  private buildPayload(): { nombre: string; aliases: string[] } {
    const nombre = this.formNombre().trim();
    const rawAliases = this.formAliases();
    const normalized = rawAliases
      .map((a) => a.trim())
      .filter((a) => !!a)
      .filter((value, index, self) => self.findIndex((x) => x.toLowerCase() === value.toLowerCase()) === index);
    return { nombre, aliases: normalized };
  }
}
