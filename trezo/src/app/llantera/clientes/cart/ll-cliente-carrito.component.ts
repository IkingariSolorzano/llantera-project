import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { LlCartService } from './ll-cart.service';

@Component({
  selector: 'app-ll-cliente-carrito',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './ll-cliente-carrito.component.html',
})
export class LlClienteCarritoComponent {
  readonly cart = inject(LlCartService);
  private readonly router = inject(Router);

  readonly showToast = signal(false);
  readonly toastMessage = signal('');
  readonly toastType = signal<'success' | 'error'>('success');
  private toastTimeout: number | null = null;

  incrementQuantity(sku: string): void {
    const result = this.cart.incrementQuantity(sku);
    if (!result.success) {
      this.showNotification(result.message, 'error');
    }
  }

  decrementQuantity(sku: string): void {
    const result = this.cart.decrementQuantity(sku);
    if (!result.success) {
      this.showNotification(result.message, 'error');
    }
  }

  removeItem(sku: string): void {
    this.cart.removeItem(sku);
    this.showNotification('Producto eliminado del carrito', 'success');
  }

  clearCart(): void {
    this.cart.clearCart();
    this.showNotification('Carrito vaciado', 'success');
  }

  goToCheckout(): void {
    this.router.navigate(['/cliente/checkout']);
  }

  goToCatalog(): void {
    // Emitir evento para cambiar de tab - se manejará en el componente padre
    window.dispatchEvent(new CustomEvent('ll-change-tab', { detail: 'catalogo' }));
  }

  formatPrice(price: number): string {
    return price.toLocaleString('es-MX', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
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
}
