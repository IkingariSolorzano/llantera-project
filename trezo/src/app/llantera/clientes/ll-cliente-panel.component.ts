import { CommonModule, NgClass, isPlatformBrowser } from '@angular/common';
import { Component, inject, signal, computed, OnInit, OnDestroy, PLATFORM_ID } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { LlClientesCatalogPageComponent } from './catalogo/ll-clientes-catalog-page.component';
import { LlClienteCuentaComponent } from './cuenta/ll-cliente-cuenta.component';
import { LlClienteDireccionesComponent } from './direcciones/ll-cliente-direcciones.component';
import { LlClientePedidosComponent } from './pedidos/ll-cliente-pedidos.component';
import { LlClienteComprasComponent } from './compras/ll-cliente-compras.component';
import { LlClienteCarritoComponent } from './cart/ll-cliente-carrito.component';
import { LlAuthService } from '../auth/ll-auth.service';
import { ToggleService } from '../../common/header/toggle.service';
import { LlCartService } from './cart/ll-cart.service';
import { LlNotificationBellComponent } from '../notifications/ll-notification-bell.component';

export type ClienteTab = 'catalogo' | 'carrito' | 'pedidos' | 'compras' | 'cuenta' | 'direcciones';

@Component({
  selector: 'app-ll-cliente-panel',
  standalone: true,
  imports: [
    CommonModule,
    NgClass,
    RouterLink,
    LlClientesCatalogPageComponent,
    LlClientePedidosComponent,
    LlClienteComprasComponent,
    LlClienteCuentaComponent,
    LlClienteDireccionesComponent,
    LlClienteCarritoComponent,
    LlNotificationBellComponent,
  ],
  templateUrl: './ll-cliente-panel.component.html',
  styleUrl: './ll-cliente-panel.component.scss',
})
export class LlClientePanelComponent implements OnInit, OnDestroy {
  private readonly auth = inject(LlAuthService);
  private readonly toggleService = inject(ToggleService);
  private readonly router = inject(Router);
  private readonly platformId = inject(PLATFORM_ID);
  readonly cart = inject(LlCartService);

  readonly activeTab = signal<ClienteTab>('catalogo');
  readonly showCartPreview = signal(false);
  readonly showUserMenu = signal(false);
  readonly isCartHovered = signal(false);
  private cartHoverTimeout: number | null = null;
  private tabChangeHandler: ((e: Event) => void) | null = null;

  readonly userName = computed(() => this.auth.user()?.name || 'Usuario');
  readonly userEmail = computed(() => this.auth.user()?.email || '');
  readonly userInitials = computed(() => {
    const name = this.userName();
    const parts = name.split(' ').filter(p => p.length > 0);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  });

  constructor() {
    this.toggleService.initializeTheme();
  }

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.tabChangeHandler = (e: Event) => {
        const customEvent = e as CustomEvent<ClienteTab>;
        if (customEvent.detail) {
          this.setTab(customEvent.detail);
        }
      };
      window.addEventListener('ll-change-tab', this.tabChangeHandler);
    }
  }

  ngOnDestroy(): void {
    if (isPlatformBrowser(this.platformId) && this.tabChangeHandler) {
      window.removeEventListener('ll-change-tab', this.tabChangeHandler);
    }
  }

  setTab(tab: ClienteTab): void {
    this.activeTab.set(tab);
    this.closeAllDropdowns();
  }

  logout(): void {
    this.closeAllDropdowns();
    this.auth.logout();
  }

  toggleTheme(): void {
    this.toggleService.toggleTheme();
  }

  toggleCartPreview(): void {
    this.showUserMenu.set(false);
    this.showCartPreview.update(v => !v);
  }

  onCartMouseEnter(): void {
    if (this.cartHoverTimeout) {
      clearTimeout(this.cartHoverTimeout);
      this.cartHoverTimeout = null;
    }
    this.showUserMenu.set(false);
    this.isCartHovered.set(true);
    this.showCartPreview.set(true);
  }

  onCartMouseLeave(): void {
    this.cartHoverTimeout = window.setTimeout(() => {
      this.isCartHovered.set(false);
      this.showCartPreview.set(false);
      this.cartHoverTimeout = null;
    }, 200);
  }

  goToCart(): void {
    this.closeAllDropdowns();
    this.activeTab.set('carrito');
  }

  incrementCartItem(sku: string): void {
    this.cart.incrementQuantity(sku);
  }

  decrementCartItem(sku: string): void {
    this.cart.decrementQuantity(sku);
  }

  toggleUserMenu(): void {
    this.showCartPreview.set(false);
    this.showUserMenu.update(v => !v);
  }

  closeAllDropdowns(): void {
    this.showCartPreview.set(false);
    this.showUserMenu.set(false);
  }

  removeFromCart(sku: string): void {
    this.cart.removeItem(sku);
  }

  updateCartQuantity(sku: string, event: Event): void {
    const input = event.target as HTMLInputElement;
    const quantity = parseInt(input.value, 10);
    if (!isNaN(quantity)) {
      this.cart.updateQuantity(sku, quantity);
    }
  }

  goToCheckout(): void {
    this.closeAllDropdowns();
    this.router.navigate(['/cliente/checkout']);
  }

  formatPrice(price: number): string {
    return price.toLocaleString('es-MX', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }
}
