import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, inject, signal, OnInit, PLATFORM_ID } from '@angular/core';
import { LlOrderService, Order, OrderStatus } from '../services/ll-order.service';
import { LlPdfGeneratorService } from '../../services/ll-pdf-generator.service';

@Component({
  selector: 'app-ll-cliente-pedidos',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './ll-cliente-pedidos.component.html',
})
export class LlClientePedidosComponent implements OnInit {
  private readonly orderService = inject(LlOrderService);
  private readonly pdfService = inject(LlPdfGeneratorService);
  private readonly platformId = inject(PLATFORM_ID);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly selectedPedido = signal<Order | null>(null);
  readonly pedidos = signal<Order[]>([]);
  
  // Modal de confirmación de cancelación
  readonly showCancelModal = signal(false);
  readonly pedidoToCancel = signal<Order | null>(null);
  readonly cancelling = signal(false);

  readonly estadoConfig: Record<OrderStatus, { label: string; color: string; icon: string }> = {
    solicitado: { label: 'Solicitado', color: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400', icon: 'schedule' },
    preparando: { label: 'Preparando', color: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400', icon: 'inventory_2' },
    enviado: { label: 'Enviado', color: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400', icon: 'local_shipping' },
    entregado: { label: 'Entregado', color: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400', icon: 'check_circle' },
    cancelado: { label: 'Cancelado', color: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400', icon: 'cancel' },
  };

  readonly pasos: OrderStatus[] = ['solicitado', 'preparando', 'enviado', 'entregado'];

  ngOnInit(): void {
    this.loadOrders();
  }

  loadOrders(): void {
    this.loading.set(true);
    this.error.set(null);

    this.orderService.listMyOrders(50, 0).subscribe({
      next: (response) => {
        this.pedidos.set(response.items || []);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Error cargando pedidos:', err);
        this.error.set('Error al cargar los pedidos');
        this.loading.set(false);
      }
    });
  }

  openDetail(pedido: Order): void {
    this.selectedPedido.set(pedido);
  }

  closeDetail(): void {
    this.selectedPedido.set(null);
  }

  getEstadoIndex(estado: OrderStatus): number {
    return this.pasos.indexOf(estado);
  }

  isPasoCompletado(pedido: Order, paso: OrderStatus): boolean {
    if (pedido.status === 'cancelado') return false;
    return this.getEstadoIndex(pedido.status) >= this.getEstadoIndex(paso);
  }

  isPasoActual(pedido: Order, paso: OrderStatus): boolean {
    return pedido.status === paso;
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

  formatDireccion(pedido: Order): string {
    const addr = pedido.shippingAddress;
    if (!addr) return 'Sin dirección';
    let dir = `${addr.street} #${addr.exteriorNumber}`;
    if (addr.interiorNumber) dir += ` Int. ${addr.interiorNumber}`;
    dir += `, ${addr.neighborhood}, C.P. ${addr.postalCode}, ${addr.city}, ${addr.state}`;
    return dir;
  }

  getMetodoPagoLabel(metodo: string): string {
    return this.orderService.getPaymentMethodLabel(metodo as any);
  }

  getModalidadPagoLabel(modalidad: string): string {
    return this.orderService.getPaymentModeLabel(modalidad as any);
  }

  canDownloadComprobante(pedido: Order): boolean {
    return pedido.status !== 'cancelado';
  }

  downloadComprobante(pedido: Order): void {
    this.pdfService.generateOrderReceipt(pedido);
  }

  canCancelOrder(pedido: Order): boolean {
    return pedido.status === 'solicitado';
  }

  openCancelModal(pedido: Order): void {
    this.pedidoToCancel.set(pedido);
    this.showCancelModal.set(true);
  }

  closeCancelModal(): void {
    this.showCancelModal.set(false);
    this.pedidoToCancel.set(null);
  }

  confirmCancelOrder(): void {
    const pedido = this.pedidoToCancel();
    if (!pedido) return;

    this.cancelling.set(true);
    this.orderService.cancelMyOrder(pedido.id).subscribe({
      next: (updated) => {
        this.pedidos.update(list =>
          list.map(p => p.id === updated.id ? updated : p)
        );
        if (this.selectedPedido()?.id === updated.id) {
          this.selectedPedido.set(updated);
        }
        this.cancelling.set(false);
        this.closeCancelModal();
        
        if (isPlatformBrowser(this.platformId)) {
          window.dispatchEvent(new CustomEvent('ll-reload-inventory'));
        }
      },
      error: (err) => {
        console.error('Error cancelando pedido:', err);
        this.error.set('Error al cancelar el pedido');
        this.cancelling.set(false);
        this.closeCancelModal();
      }
    });
  }
}
