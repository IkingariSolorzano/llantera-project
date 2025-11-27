import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, DestroyRef, ElementRef, QueryList, ViewChildren, computed, inject, signal, OnInit, OnDestroy, PLATFORM_ID } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { catchError, finalize, tap } from 'rxjs/operators';
import { of } from 'rxjs';

import {
  LlLlantasCatalogService,
  LlTireCatalogFilter,
  LlTireDTO,
  LlTireAdminDTO,
  LlTireAdminUpdatePayload,
  ApiListResponse,
} from './ll-llantas-catalog.service';
import { LlLlantasBrandsService, LlTireBrandDTO } from './ll-llantas-brands.service';

@Component({
  selector: 'app-ll-llantas-list',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './ll-llantas-list.component.html',
  styleUrl: './ll-llantas-list.component.scss',
})
export class LlLlantasListComponent implements OnInit, OnDestroy {
  private readonly catalogService = inject(LlLlantasCatalogService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly brandsService = inject(LlLlantasBrandsService);
  private readonly platformId = inject(PLATFORM_ID);
  
  private reloadInventoryHandler = () => this.loadTires(this.searchTerm(), this.pageIndex(), this.pageSize());

  @ViewChildren('cantidadInput')
  private cantidadInputs!: QueryList<ElementRef<HTMLInputElement>>;

  readonly adminItems = signal<LlTireAdminDTO[]>([]);
  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly searchTerm = signal('');
  readonly pageIndex = signal(0);
  readonly pageSize = signal(25);
  readonly totalCount = signal(0);
  readonly totalTires = computed(() => this.totalCount());
  readonly totalPages = computed(() => {
    const size = this.pageSize();
    const total = this.totalCount();
    if (size <= 0) return 0;
    return Math.max(1, Math.ceil(total / size));
  });

  readonly pageButtons = computed(() => {
    const total = this.totalPages();
    const current = this.pageIndex();
    const buttons: { type: 'page' | 'ellipsis'; index?: number; label: string }[] = [];

    if (total <= 1) {
      return buttons;
    }

    const pushPage = (i: number) => {
      buttons.push({ type: 'page', index: i, label: String(i + 1) });
    };
    const pushEllipsis = () => {
      if (buttons.length === 0 || buttons[buttons.length - 1].type === 'ellipsis') {
        return;
      }
      buttons.push({ type: 'ellipsis', label: '…' });
    };

    if (total <= 7) {
      for (let i = 0; i < total; i++) {
        pushPage(i);
      }
      return buttons;
    }

    const addRange = (start: number, end: number) => {
      for (let i = start; i <= end; i++) {
        if (i >= 0 && i < total) {
          pushPage(i);
        }
      }
    };

    // Siempre primera página
    pushPage(0);

    const windowStart = Math.max(1, current - 1);
    const windowEnd = Math.min(total - 2, current + 1);

    if (windowStart > 1) {
      pushEllipsis();
    }

    addRange(windowStart, windowEnd);

    if (windowEnd < total - 2) {
      pushEllipsis();
    }

    // Siempre última página
    pushPage(total - 1);

    return buttons;
  });

  readonly brands = signal<LlTireBrandDTO[]>([]);
  readonly priceColumns = signal<string[]>([]);

  readonly editingSku = signal<string | null>(null);
  readonly editBuffer = signal<LlTireAdminUpdatePayload | null>(null);
  readonly confirmEditTire = signal<LlTireDTO | null>(null);

  readonly showImportConfirm = signal(false);
  readonly importFileName = signal<string | null>(null);
  private pendingImportFile: File | null = null;

  readonly isImporting = signal(false);
  readonly showImportResult = signal(false);
  readonly importResultProcessed = signal<number | null>(null);
  readonly importProgress = signal(0);

  private importProgressTimer: ReturnType<typeof setInterval> | null = null;

  constructor() {
    this.loadTires();
    this.loadBrands();
  }

  ngOnInit(): void {
    // Escuchar evento de recarga de inventario (cuando se crea/cancela/entrega un pedido)
    if (isPlatformBrowser(this.platformId)) {
      window.addEventListener('ll-reload-inventory', this.reloadInventoryHandler);
    }
  }

  ngOnDestroy(): void {
    if (isPlatformBrowser(this.platformId)) {
      window.removeEventListener('ll-reload-inventory', this.reloadInventoryHandler);
    }
  }

  downloadCatalog(): void {
    const filter: LlTireCatalogFilter = {
      search: this.searchTerm() || undefined,
    };

    this.catalogService
      .exportAdminXlsx(filter)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (blob) => {
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = 'catalogo_llantas.xlsx';
          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);
          window.URL.revokeObjectURL(url);
        },
        error: (err) => {
          console.error('Error al exportar catálogo de llantas', err);
          this.error.set('No se pudo descargar el catálogo de llantas.');
        },
      });
  }

  onInventoryFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input || !input.files || input.files.length === 0) {
      return;
    }
    const file = input.files[0];
    console.log('Archivo de inventario seleccionado:', file.name, file.size);

    // Guardamos el archivo seleccionado y mostramos un modal de confirmación.
    this.pendingImportFile = file;
    this.importFileName.set(file.name);
    this.showImportConfirm.set(true);

    // Limpiamos el input para permitir volver a seleccionar el mismo archivo si se cancela.
    input.value = '';
  }

  confirmImport(): void {
    const file = this.pendingImportFile;
    if (!file) {
      this.showImportConfirm.set(false);
      return;
    }

    this.showImportConfirm.set(false);
    this.importProgress.set(0);
    this.startImportProgressTimer();
    this.isImporting.set(true);
    this.catalogService
      .importAdminXlsx(file)
      .pipe(
        finalize(() => {
          this.isImporting.set(false);
          this.stopImportProgressTimer();
          this.pendingImportFile = null;
          this.importFileName.set(null);
        }),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: (result) => {
          const processed = result?.processed ?? 0;
          this.importResultProcessed.set(processed);
          this.showImportResult.set(true);
          this.loadTires(this.searchTerm(), this.pageIndex(), this.pageSize());
        },
        error: (err) => {
          console.error('Error al importar catálogo de llantas', err);
          this.error.set('No se pudo importar el archivo de catálogo. Revisa el formato.');
        },
      });
  }

  cancelImport(): void {
    this.showImportConfirm.set(false);
    this.pendingImportFile = null;
    this.importFileName.set(null);
  }

  closeImportResult(): void {
    this.showImportResult.set(false);
  }

  private startImportProgressTimer(): void {
    if (this.importProgressTimer) {
      clearInterval(this.importProgressTimer);
    }

    let elapsedMs = 0;
    const intervalMs = 100;

    this.importProgressTimer = setInterval(() => {
      elapsedMs += intervalMs;
      const elapsedSeconds = elapsedMs / 1000;

      let value: number;
      if (elapsedSeconds <= 5) {
        // Primer tramo: 0 a 50% en 5 segundos.
        value = (elapsedSeconds / 5) * 50;
      } else {
        // Segundo tramo: de 50% a un máximo de 90% de forma lenta.
        const extra = elapsedSeconds - 5;
        const factor = Math.min(extra / 20, 1); // tras ~20s adicionales llega cerca del máximo.
        value = 50 + factor * 40; // 50% + 40% = 90%
      }

      if (value > 90) {
        value = 90;
      }

      this.importProgress.set(value);
    }, intervalMs);
  }

  private stopImportProgressTimer(): void {
    if (this.importProgressTimer) {
      clearInterval(this.importProgressTimer);
      this.importProgressTimer = null;
    }
  }

  startEdit(sku: string): void {
    const current = this.adminItems().find((i) => i.tire.sku === sku);
    if (!current) {
      return;
    }

    const precios: { [code: string]: number | null } = {};
    const prices = current.prices || {};
    for (const code of this.priceColumns()) {
      const value = prices[code];
      precios[code] = typeof value === 'number' ? value : null;
    }

    this.editingSku.set(sku);
    this.editBuffer.set({
      cantidad: current.inventory?.cantidad ?? null,
      precios,
    });

    // Enfocar el campo de cantidad de la fila editada
    setTimeout(() => {
      const index = this.adminItems().findIndex((i) => i.tire.sku === sku);
      if (index < 0 || !this.cantidadInputs) {
        return;
      }
      const inputRef = this.cantidadInputs.get(index);
      if (inputRef?.nativeElement) {
        inputRef.nativeElement.focus();
        inputRef.nativeElement.select();
      }
    }, 0);
  }

  cancelEdit(): void {
    this.editingSku.set(null);
    this.editBuffer.set(null);
  }

  updateEditCantidad(value: string | number): void {
    const buffer = this.editBuffer();
    if (!buffer) {
      return;
    }
    const parsed =
      value === null || value === undefined || value === ''
        ? null
        : typeof value === 'number'
          ? value
          : Number(value);
    if (parsed !== null && Number.isNaN(parsed)) {
      return;
    }
    this.editBuffer.set({
      ...buffer,
      cantidad: parsed,
    });
  }

  updateEditPrice(code: string, value: string | number): void {
    const buffer = this.editBuffer();
    if (!buffer) {
      return;
    }
    const precios = { ...(buffer.precios || {}) };
    if (value === null || value === undefined || value === '') {
      precios[code] = null;
    } else {
      const parsed = typeof value === 'number' ? value : Number(value);
      if (Number.isNaN(parsed)) {
        return;
      }
      precios[code] = parsed;
    }
    this.editBuffer.set({
      ...buffer,
      precios,
    });
  }

  saveEdit(tire: LlTireDTO): void {
    const sku = tire.sku;
    const buffer = this.editBuffer();
    if (!buffer) {
      return;
    }

    const payload: LlTireAdminUpdatePayload = {
      cantidad: buffer.cantidad,
      precios: buffer.precios,
    };

    this.loading.set(true);

    this.catalogService
      .updateAdmin(sku, payload)
      .pipe(
        tap(() => {
          this.editingSku.set(null);
          this.editBuffer.set(null);
          this.loadTires(this.searchTerm());
        }),
        catchError((err) => {
          console.error('Error al guardar cambios de llanta', err);
          this.error.set('No se pudieron guardar los cambios de la llanta.');
          return of(null);
        }),
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  onEnterEdit(tire: LlTireDTO): void {
    this.confirmEditTire.set(tire);
  }

  confirmSave(): void {
    const tire = this.confirmEditTire();
    if (!tire) {
      return;
    }
    this.confirmEditTire.set(null);
    this.saveEdit(tire);
  }

  cancelConfirm(): void {
    this.confirmEditTire.set(null);
  }

  onEscapeEdit(): void {
    this.cancelEdit();
  }

  onSearchChange(value: string): void {
    const sanitized = value.trimStart();
    this.searchTerm.set(sanitized);
    this.pageIndex.set(0);
    this.loadTires(sanitized, 0, this.pageSize());
  }

  retry(): void {
    this.loadTires(this.searchTerm(), this.pageIndex(), this.pageSize());
  }

  onEditTire(tire: LlTireDTO): void {
    this.startEdit(tire.sku);
  }

  onRowDblClick(tire: LlTireDTO): void {
    this.onEditTire(tire);
  }

  onOpenTireForm(tire: LlTireDTO): void {
    if (!tire.sku) {
      return;
    }
    this.router.navigate(['/dashboard/llantas/editar', tire.sku]);
  }

  getBrandName(marcaId: number | null | undefined): string {
    if (!marcaId) {
      return 'Sin marca';
    }
    const brand = this.brands().find((b) => b.id === marcaId);
    return brand ? brand.nombre : `Marca #${marcaId}`;
  }

  formatMedida(tire: LlTireDTO): string {
    if (tire.medidaOriginal && tire.medidaOriginal.trim()) {
      return tire.medidaOriginal;
    }
    const ancho = tire.ancho ? String(tire.ancho) : '';
    const perfil = typeof tire.perfil === 'number' ? `/${tire.perfil}` : '';
    const rinValue = typeof tire.rin === 'number' ? tire.rin : null;
    const rin = rinValue !== null
      ? ` R${Number.isInteger(rinValue) ? rinValue.toFixed(0) : rinValue}`
      : '';
    const value = `${ancho}${perfil}${rin}`.trim();
    return value || 'N/D';
  }

  formatRadial(tire: LlTireDTO): string {
    const rinValue = typeof tire.rin === 'number' ? tire.rin : null;
    if (rinValue === null) {
      return '';
    }
    const base = Number.isInteger(rinValue) ? rinValue.toFixed(0) : String(rinValue);
    return `R${base}`;
  }

  changePageSize(size: number | string): void {
    const parsed = typeof size === 'number' ? size : Number(size);
    if (!parsed || Number.isNaN(parsed)) {
      return;
    }
    this.pageSize.set(parsed);
    this.pageIndex.set(0);
    this.loadTires(this.searchTerm(), 0, parsed);
  }

  goToPage(page: number): void {
    const currentTotalPages = this.totalPages();
    if (page < 0 || page >= currentTotalPages) {
      return;
    }
    this.pageIndex.set(page);
    this.loadTires(this.searchTerm(), page, this.pageSize());
  }

  prevPage(): void {
    this.goToPage(this.pageIndex() - 1);
  }

  nextPage(): void {
    this.goToPage(this.pageIndex() + 1);
  }

  private loadTires(search = '', pageIndex = this.pageIndex(), pageSize = this.pageSize()): void {
    this.loading.set(true);
    this.error.set(null);

    const filter: LlTireCatalogFilter = {
      search,
      limit: pageSize,
      offset: pageIndex * pageSize,
    };

    this.catalogService
      .listAdmin(filter)
      .pipe(
        tap((response: ApiListResponse<LlTireAdminDTO>) => {
          const items = response.data ?? [];
          this.adminItems.set(items);
          this.totalCount.set(response.meta?.total ?? items.length ?? 0);

          const codes = new Set<string>();
          for (const item of items ?? []) {
            const prices = item.prices || {};
            for (const key of Object.keys(prices)) {
              if (key && key.trim()) {
                codes.add(key.trim());
              }
            }
          }
          this.priceColumns.set(Array.from(codes).sort());
        }),
        catchError((err) => {
          console.error('Error al cargar llantas', err);
          this.error.set('No se pudieron cargar las llantas.');
          this.adminItems.set([]);
          this.priceColumns.set([]);
          this.totalCount.set(0);
          return of<ApiListResponse<LlTireAdminDTO>>({ data: [], meta: { total: 0, limit: filter.limit ?? 0, offset: filter.offset ?? 0 } });
        }),
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }

  private loadBrands(): void {
    this.brandsService
      .list()
      .pipe(
        tap((items) => this.brands.set(items)),
        catchError((err) => {
          console.error('Error al cargar marcas de llantas', err);
          return of<LlTireBrandDTO[]>([]);
        }),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe();
  }
}
