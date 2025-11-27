import { Injectable, signal, computed, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { LlTireCatalogItemDTO } from '../../llantas/ll-llantas-catalog.service';
import { LlAuthService } from '../../auth/ll-auth.service';
import { DEFAULT_API_BASE_URL } from '../../../core/config/api.config';

export interface CartItem {
  item: LlTireCatalogItemDTO;
  quantity: number;
  addedAt: Date;
  maxStock: number;
}

// Interfaces para respuestas del backend
export interface CartItemFromServer {
  id: number;
  cartId: number;
  tireSku: string;
  quantity: number;
  addedAt: string;
  updatedAt: string;
  tireMeasure?: string;
  tireBrand?: string;
  tireModel?: string;
  tireImage?: string;
  price?: number;
  stock?: number;
}

export interface CartFromServer {
  id: number;
  userId: string;
  items: CartItemFromServer[];
  subtotal: number;
  itemCount: number;
  createdAt: string;
  updatedAt: string;
}

const CART_STORAGE_KEY = 'll_cart';

@Injectable({ providedIn: 'root' })
export class LlCartService {
  private readonly platformId = inject(PLATFORM_ID);
  private readonly http = inject(HttpClient);
  private readonly auth = inject(LlAuthService);
  private readonly baseUrl = DEFAULT_API_BASE_URL;
  
  private readonly _items = signal<CartItem[]>([]);
  private readonly _loading = signal(false);
  private readonly _synced = signal(false);
  
  readonly items = this._items.asReadonly();
  readonly loading = this._loading.asReadonly();
  
  private readonly IVA_RATE = 0.16; // 16% IVA

  readonly itemCount = computed(() => 
    this._items().reduce((sum, ci) => sum + ci.quantity, 0)
  );
  
  /** Subtotal sin IVA (suma de precios de productos) */
  readonly subtotal = computed(() =>
    this._items().reduce((sum, ci) => sum + (ci.item.price * ci.quantity), 0)
  );

  /** IVA calculado (16% del subtotal) */
  readonly iva = computed(() =>
    Math.round(this.subtotal() * this.IVA_RATE * 100) / 100
  );

  /** Total con IVA */
  readonly total = computed(() =>
    Math.round((this.subtotal() + this.iva()) * 100) / 100
  );
  
  readonly isEmpty = computed(() => this._items().length === 0);

  /** Obtener resumen de totales para mostrar/guardar */
  getCartSummary() {
    return {
      items: this._items(),
      itemCount: this.itemCount(),
      subtotal: this.subtotal(),
      iva: this.iva(),
      total: this.total(),
    };
  }

  constructor() {
    this.loadFromStorage();
    // Sincronizar con el servidor cuando el usuario esté autenticado
    this.syncWithServer();
  }

  /**
   * Sincroniza el carrito con el servidor
   */
  async syncWithServer(): Promise<void> {
    if (!isPlatformBrowser(this.platformId)) return;
    if (!this.auth.isAuthenticated()) return;
    if (this._synced()) return;

    try {
      this._loading.set(true);
      const user = this.auth.user();
      const level = user?.level || 'public';
      
      const serverCart = await firstValueFrom(
        this.http.get<CartFromServer>(`${this.baseUrl}/cart/?level=${level}`)
      );

      if (serverCart && serverCart.items && serverCart.items.length > 0) {
        // Convertir items del servidor al formato local
        const items: CartItem[] = serverCart.items.map(si => ({
          item: {
            tire: {
              id: '',
              sku: si.tireSku,
              marcaId: 0,
              modelo: si.tireModel || '',
              ancho: 0,
              rin: 0,
              construccion: '',
              tipoTubo: '',
              calificacionCapas: '',
              indiceCarga: '',
              indiceVelocidad: '',
              abreviaturaUso: '',
              descripcion: '',
              precioPublico: si.price || 0,
              urlImagen: si.tireImage || '',
              medidaOriginal: si.tireMeasure || si.tireSku,
              creadoEn: '',
              actualizadoEn: '',
            },
            price: si.price || 0,
            priceCode: 'server',
            stock: si.stock,
          },
          quantity: si.quantity,
          addedAt: new Date(si.addedAt),
          maxStock: si.stock || 999,
        }));
        this._items.set(items);
        this.saveToStorage();
      } else if (this._items().length > 0) {
        // Si el servidor está vacío pero tenemos items locales, sincronizar al servidor
        await this.pushLocalCartToServer();
      }
      
      this._synced.set(true);
    } catch (error) {
      console.error('Error sincronizando carrito:', error);
      // Mantener carrito local si falla la sincronización
    } finally {
      this._loading.set(false);
    }
  }

  /**
   * Envía el carrito local al servidor
   */
  private async pushLocalCartToServer(): Promise<void> {
    for (const item of this._items()) {
      try {
        await firstValueFrom(
          this.http.post(`${this.baseUrl}/cart/`, {
            tireSku: item.item.tire.sku,
            quantity: item.quantity,
          })
        );
      } catch (error) {
        console.error('Error enviando item al servidor:', error);
      }
    }
  }

  private loadFromStorage(): void {
    if (!isPlatformBrowser(this.platformId)) return;
    
    try {
      const raw = localStorage.getItem(CART_STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw) as CartItem[];
        this._items.set(parsed.map(ci => ({
          ...ci,
          addedAt: new Date(ci.addedAt)
        })));
      }
    } catch {
      // Ignorar errores de parsing
    }
  }

  private saveToStorage(): void {
    if (!isPlatformBrowser(this.platformId)) return;
    
    try {
      localStorage.setItem(CART_STORAGE_KEY, JSON.stringify(this._items()));
    } catch {
      // Ignorar errores de almacenamiento
    }
  }

  /**
   * Agrega un producto al carrito
   * @returns true si se agregó correctamente, false si no hay stock suficiente
   */
  addItem(catalogItem: LlTireCatalogItemDTO, quantity: number = 1): { success: boolean; message: string } {
    if (quantity <= 0) {
      return { success: false, message: 'La cantidad debe ser mayor a 0' };
    }

    const existingIndex = this._items().findIndex(
      ci => ci.item.tire.sku === catalogItem.tire.sku
    );

    // Usar stock real del catálogo o un valor por defecto
    const availableStock = catalogItem.stock ?? 999;

    if (existingIndex >= 0) {
      const current = this._items()[existingIndex];
      const newQuantity = current.quantity + quantity;
      
      if (newQuantity > availableStock) {
        return { 
          success: false, 
          message: `Solo hay ${availableStock} unidades disponibles. Ya tienes ${current.quantity} en el carrito.` 
        };
      }

      this._items.update(items => 
        items.map((ci, i) => 
          i === existingIndex 
            ? { ...ci, quantity: newQuantity, maxStock: availableStock }
            : ci
        )
      );
    } else {
      if (quantity > availableStock) {
        return { 
          success: false, 
          message: `Solo hay ${availableStock} unidades disponibles.` 
        };
      }

      this._items.update(items => [
        ...items,
        { item: catalogItem, quantity, addedAt: new Date(), maxStock: availableStock }
      ]);
    }

    this.saveToStorage();
    this.syncAddItemToServer(catalogItem.tire.sku, quantity);
    return { success: true, message: 'Producto agregado al carrito' };
  }

  private syncAddItemToServer(sku: string, quantity: number): void {
    if (!this.auth.isAuthenticated()) return;
    this.http.post(`${this.baseUrl}/cart/`, { tireSku: sku, quantity }).subscribe({
      error: (err) => console.error('Error sincronizando item:', err)
    });
  }

  updateQuantity(sku: string, quantity: number): { success: boolean; message: string } {
    if (quantity <= 0) {
      this.removeItem(sku);
      return { success: true, message: 'Producto eliminado del carrito' };
    }

    const cartItem = this._items().find(ci => ci.item.tire.sku === sku);
    if (!cartItem) {
      return { success: false, message: 'Producto no encontrado en el carrito' };
    }

    const availableStock = cartItem.maxStock ?? cartItem.item.stock ?? 999;

    if (quantity > availableStock) {
      return { 
        success: false, 
        message: `Solo hay ${availableStock} unidades disponibles.` 
      };
    }

    this._items.update(items =>
      items.map(ci =>
        ci.item.tire.sku === sku
          ? { ...ci, quantity }
          : ci
      )
    );

    this.saveToStorage();
    this.syncUpdateQuantityToServer(sku, quantity);
    return { success: true, message: 'Cantidad actualizada' };
  }

  private syncUpdateQuantityToServer(sku: string, quantity: number): void {
    if (!this.auth.isAuthenticated()) return;
    this.http.put(`${this.baseUrl}/cart/items/${sku}`, { quantity }).subscribe({
      error: (err) => console.error('Error actualizando cantidad:', err)
    });
  }

  incrementQuantity(sku: string): { success: boolean; message: string } {
    const cartItem = this._items().find(ci => ci.item.tire.sku === sku);
    if (!cartItem) {
      return { success: false, message: 'Producto no encontrado' };
    }
    return this.updateQuantity(sku, cartItem.quantity + 1);
  }

  decrementQuantity(sku: string): { success: boolean; message: string } {
    const cartItem = this._items().find(ci => ci.item.tire.sku === sku);
    if (!cartItem) {
      return { success: false, message: 'Producto no encontrado' };
    }
    return this.updateQuantity(sku, cartItem.quantity - 1);
  }

  removeItem(sku: string): void {
    this._items.update(items => 
      items.filter(ci => ci.item.tire.sku !== sku)
    );
    this.saveToStorage();
    this.syncRemoveItemFromServer(sku);
  }

  private syncRemoveItemFromServer(sku: string): void {
    if (!this.auth.isAuthenticated()) return;
    this.http.delete(`${this.baseUrl}/cart/items/${sku}`).subscribe({
      error: (err) => console.error('Error eliminando item:', err)
    });
  }

  clearCart(): void {
    this._items.set([]);
    this.saveToStorage();
    this.syncClearCartOnServer();
  }

  private syncClearCartOnServer(): void {
    if (!this.auth.isAuthenticated()) return;
    this.http.delete(`${this.baseUrl}/cart/`).subscribe({
      error: (err) => console.error('Error vaciando carrito:', err)
    });
  }

  getItemBySku(sku: string): CartItem | undefined {
    return this._items().find(ci => ci.item.tire.sku === sku);
  }

  isInCart(sku: string): boolean {
    return this._items().some(ci => ci.item.tire.sku === sku);
  }
}
