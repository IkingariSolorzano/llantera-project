import { CommonModule } from '@angular/common';
import { Component, inject, signal, OnInit } from '@angular/core';
import { LlOrderService, Order, OrderStatus } from '../services/ll-order.service';

@Component({
  selector: 'app-ll-cliente-compras',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './ll-cliente-compras.component.html',
})
export class LlClienteComprasComponent implements OnInit {
  private readonly orderService = inject(LlOrderService);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly compras = signal<Order[]>([]);
  readonly selected = signal<Order | null>(null);

  ngOnInit(): void {
    this.loadCompras();
  }

  loadCompras(): void {
    this.loading.set(true);
    this.error.set(null);

    // Cargar pedidos y filtrar solo los entregados (compras completadas)
    this.orderService.listMyOrders(50, 0).subscribe({
      next: (response) => {
        const entregados = (response.items || []).filter(o => o.status === 'entregado');
        this.compras.set(entregados);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Error cargando compras:', err);
        this.error.set('Error al cargar las compras');
        this.loading.set(false);
      }
    });
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return new Intl.DateTimeFormat('es-MX', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    }).format(date);
  }

  formatPrice(price: number): string {
    return price.toLocaleString('es-MX', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  openDetail(compra: Order): void {
    this.selected.set(compra);
  }

  closeDetail(): void {
    this.selected.set(null);
  }

  downloadComprobante(compra: Order): void {
    // TODO: Implementar descarga de comprobante de compra (PDF/Factura)
    console.log('Descargar comprobante de compra:', compra.orderNumber);
    alert('Descarga de comprobante en desarrollo');
  }
}
