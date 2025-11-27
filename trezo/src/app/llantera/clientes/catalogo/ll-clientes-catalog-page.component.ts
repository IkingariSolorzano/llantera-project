import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, inject, signal, computed, OnInit, OnDestroy, PLATFORM_ID } from '@angular/core';
import { FormsModule } from '@angular/forms';

import {
  LlLlantasCatalogService,
  LlTireCatalogItemDTO,
  LlTireDTO,
  LlTireCatalogFilter,
} from '../../llantas/ll-llantas-catalog.service';
import { LlAuthService } from '../../auth/ll-auth.service';
import { LlCartService } from '../cart/ll-cart.service';

@Component({
  selector: 'app-ll-clientes-catalog-page',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './ll-clientes-catalog-page.component.html',
  styleUrl: './ll-clientes-catalog-page.component.scss',
})
export class LlClientesCatalogPageComponent implements OnInit, OnDestroy {
  private readonly catalogService = inject(LlLlantasCatalogService);
  private readonly auth = inject(LlAuthService);
  private readonly platformId = inject(PLATFORM_ID);
  readonly cart = inject(LlCartService);
  
  private reloadInventoryHandler = () => this.loadCatalog();

  readonly items = signal<LlTireCatalogItemDTO[]>([]);
  readonly allItems = signal<LlTireCatalogItemDTO[]>([]);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly genericImage = 'images/llantera/generic-tyre.png';
  
  // Cotizador / Filtros
  readonly showFilters = signal(true);
  readonly searchQuery = signal('');
  readonly onlyInStock = signal(false);
  
  // Opciones de filtros disponibles
  readonly anchos = signal<number[]>([]);
  readonly perfiles = signal<number[]>([]);
  readonly rines = signal<number[]>([]);
  readonly construcciones = signal<string[]>([]);
  readonly usos = signal<string[]>([]);
  readonly indicesCarga = signal<string[]>([]);
  
  // Filtros seleccionados
  readonly selectedAncho = signal<number | null>(null);
  readonly selectedPerfil = signal<number | null>(null);
  readonly selectedRin = signal<number | null>(null);
  readonly selectedConstruccion = signal<string | null>(null);
  readonly selectedUso = signal<string | null>(null);
  readonly selectedIndiceCarga = signal<string | null>(null);

  // Toast de notificación
  readonly showToast = signal(false);
  readonly toastMessage = signal('');
  readonly toastType = signal<'success' | 'error'>('success');
  private toastTimeout: number | null = null;

  readonly hasActiveFilters = computed(() => 
    this.selectedAncho() !== null ||
    this.selectedPerfil() !== null ||
    this.selectedRin() !== null ||
    this.selectedConstruccion() !== null ||
    this.selectedUso() !== null ||
    this.selectedIndiceCarga() !== null ||
    this.searchQuery().trim() !== '' ||
    this.onlyInStock()
  );

  constructor() {
    this.loadCatalog();
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

  private loadCatalog(): void {
    this.loading.set(true);
    this.error.set(null);

    const currentUser = this.auth.user();
    const rawLevel = currentUser?.level || 'public';
    const level = rawLevel.trim() || 'public';

    const filter: LlTireCatalogFilter = {
      limit: 10000,
      offset: 0,
      level,
    };

    // Aplicar filtros
    const ancho = this.selectedAncho();
    const perfil = this.selectedPerfil();
    const rin = this.selectedRin();
    const construccion = this.selectedConstruccion();
    const uso = this.selectedUso();
    const indiceCarga = this.selectedIndiceCarga();
    const search = this.searchQuery().trim();
    const inStock = this.onlyInStock();

    if (ancho !== null) filter.ancho = ancho;
    if (perfil !== null) filter.perfil = perfil;
    if (rin !== null) filter.rin = rin;
    if (construccion) filter.construccion = construccion;
    if (uso) filter.abreviatura = uso;
    if (indiceCarga) filter.indiceCarga = indiceCarga;
    if (search) filter.search = search;
    if (inStock) filter.inStock = true;

    this.catalogService.list(filter).subscribe({
      next: (items) => {
        const list = items ?? [];
        this.items.set(list);
        
        // Si es la carga inicial, guardar todos los items y construir opciones
        if (!this.hasActiveFilters()) {
          this.allItems.set(list);
        }
        
        this.buildFilterOptions(list);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Error al cargar catálogo de llantas', err);
        this.error.set('No se pudo cargar el catálogo de llantas.');
        this.loading.set(false);
      },
    });
  }

  private buildFilterOptions(items: LlTireCatalogItemDTO[]): void {
    const anchosSet = new Set<number>();
    const perfilesSet = new Set<number>();
    const rinesSet = new Set<number>();
    const construccionesSet = new Set<string>();
    const usosSet = new Set<string>();
    const indicesCargaSet = new Set<string>();

    for (const item of items) {
      const t = item.tire;
      if (typeof t.ancho === 'number' && !Number.isNaN(t.ancho)) anchosSet.add(t.ancho);
      if (typeof t.perfil === 'number' && !Number.isNaN(t.perfil)) perfilesSet.add(t.perfil);
      if (typeof t.rin === 'number' && !Number.isNaN(t.rin)) rinesSet.add(t.rin);
      if (t.construccion?.trim()) construccionesSet.add(t.construccion.trim());
      if (t.abreviaturaUso?.trim()) usosSet.add(t.abreviaturaUso.trim());
      if (t.indiceCarga?.trim()) indicesCargaSet.add(t.indiceCarga.trim());
    }

    this.anchos.set(Array.from(anchosSet).sort((a, b) => a - b));
    this.perfiles.set(Array.from(perfilesSet).sort((a, b) => a - b));
    this.rines.set(Array.from(rinesSet).sort((a, b) => a - b));
    this.construcciones.set(Array.from(construccionesSet).sort());
    this.usos.set(Array.from(usosSet).sort());
    this.indicesCarga.set(Array.from(indicesCargaSet).sort());
  }

  onFilterChange(field: string, value: string): void {
    switch (field) {
      case 'ancho':
        this.selectedAncho.set(value ? Number(value) : null);
        break;
      case 'perfil':
        this.selectedPerfil.set(value ? Number(value) : null);
        break;
      case 'rin':
        this.selectedRin.set(value ? Number(value) : null);
        break;
      case 'construccion':
        this.selectedConstruccion.set(value || null);
        break;
      case 'uso':
        this.selectedUso.set(value || null);
        break;
      case 'indiceCarga':
        this.selectedIndiceCarga.set(value || null);
        break;
    }
    this.loadCatalog();
  }

  onSearchInput(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.searchQuery.set(input.value);
  }

  onSearchSubmit(): void {
    this.loadCatalog();
  }

  onInStockToggle(checked: boolean): void {
    this.onlyInStock.set(checked);
    this.loadCatalog();
  }

  clearAllFilters(): void {
    this.selectedAncho.set(null);
    this.selectedPerfil.set(null);
    this.selectedRin.set(null);
    this.selectedConstruccion.set(null);
    this.selectedUso.set(null);
    this.selectedIndiceCarga.set(null);
    this.searchQuery.set('');
    this.onlyInStock.set(false);
    this.loadCatalog();
  }

  toggleFilters(): void {
    this.showFilters.update(v => !v);
  }

  addToCart(item: LlTireCatalogItemDTO): void {
    const result = this.cart.addItem(item, 1);
    this.showNotification(result.message, result.success ? 'success' : 'error');
  }

  isInCart(sku: string): boolean {
    return this.cart.isInCart(sku);
  }

  getCartQuantity(sku: string): number {
    const cartItem = this.cart.getItemBySku(sku);
    return cartItem?.quantity ?? 0;
  }

  incrementCartItem(sku: string): void {
    const result = this.cart.incrementQuantity(sku);
    if (!result.success) {
      this.showNotification(result.message, 'error');
    }
  }

  decrementCartItem(sku: string): void {
    const result = this.cart.decrementQuantity(sku);
    if (!result.success) {
      this.showNotification(result.message, 'error');
    }
  }

  removeFromCart(sku: string): void {
    this.cart.removeItem(sku);
    this.showNotification('Producto eliminado del carrito', 'success');
  }

  private showNotification(message: string, type: 'success' | 'error'): void {
    if (this.toastTimeout) {
      clearTimeout(this.toastTimeout);
    }
    
    this.toastMessage.set(message);
    this.toastType.set(type);
    this.showToast.set(true);
    
    this.toastTimeout = window.setTimeout(() => {
      this.showToast.set(false);
      this.toastTimeout = null;
    }, 3000);
  }

  getTireImage(item: LlTireCatalogItemDTO): string {
    const url = item.tire.urlImagen?.trim();
    return url ? url : this.genericImage;
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

  formatCargaVelocidad(tire: LlTireDTO): string {
    const carga = (tire.indiceCarga || '').trim();
    const vel = (tire.indiceVelocidad || '').trim();

    if (!carga && !vel) return 'N/D';
    if (!carga) return vel;
    if (!vel) return carga;
    return `${carga}/${vel}`;
  }

  formatUso(tire: LlTireDTO): string {
    return tire.abreviaturaUso || 'N/D';
  }

  formatPrice(price: number): string {
    return price.toLocaleString('es-MX', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  isAtMaxStock(item: LlTireCatalogItemDTO): boolean {
    const currentQty = this.getCartQuantity(item.tire.sku);
    const stock = item.stock ?? 0;
    return currentQty >= stock;
  }
}
