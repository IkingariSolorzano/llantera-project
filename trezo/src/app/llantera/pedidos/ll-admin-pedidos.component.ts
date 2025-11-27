import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, inject, signal, computed, OnInit, OnDestroy, PLATFORM_ID } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { LlOrderService, Order, OrderStatus, OrderItem } from '../clientes/services/ll-order.service';
import { LlPdfGeneratorService } from '../services/ll-pdf-generator.service';

@Component({
  selector: 'app-ll-admin-pedidos',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './ll-admin-pedidos.component.html',
})
export class LlAdminPedidosComponent implements OnInit, OnDestroy {
  private readonly orderService = inject(LlOrderService);
  private readonly pdfService = inject(LlPdfGeneratorService);
  private readonly platformId = inject(PLATFORM_ID);
  
  private reloadHandler = () => this.loadOrders();

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly selectedPedido = signal<Order | null>(null);
  readonly filterEstado = signal<OrderStatus | 'todos'>('todos');
  readonly searchTerm = signal('');
  readonly pedidos = signal<Order[]>([]);
  readonly total = signal(0);
  
  // Modal de confirmación de cancelación
  readonly showCancelModal = signal(false);
  readonly pedidoToCancel = signal<Order | null>(null);
  readonly cancelling = signal(false);

  readonly estadoConfig: Record<OrderStatus, { label: string; color: string; bgColor: string; icon: string }> = {
    solicitado: { label: 'Solicitado', color: 'text-blue-700 dark:text-blue-400', bgColor: 'bg-blue-100 dark:bg-blue-900/30', icon: 'schedule' },
    preparando: { label: 'Preparando', color: 'text-yellow-700 dark:text-yellow-400', bgColor: 'bg-yellow-100 dark:bg-yellow-900/30', icon: 'inventory_2' },
    enviado: { label: 'Enviado', color: 'text-purple-700 dark:text-purple-400', bgColor: 'bg-purple-100 dark:bg-purple-900/30', icon: 'local_shipping' },
    entregado: { label: 'Entregado', color: 'text-green-700 dark:text-green-400', bgColor: 'bg-green-100 dark:bg-green-900/30', icon: 'check_circle' },
    cancelado: { label: 'Cancelado', color: 'text-red-700 dark:text-red-400', bgColor: 'bg-red-100 dark:bg-red-900/30', icon: 'cancel' },
  };

  // Quitamos 'entregado' porque esos pasan a ventas
  readonly estadosDisponibles: OrderStatus[] = ['solicitado', 'preparando', 'enviado', 'cancelado'];

  // Filtrar pedidos (entregados van a ventas, no se muestran aquí)
  readonly filteredPedidos = computed(() => {
    const estado = this.filterEstado();
    const search = this.searchTerm().toLowerCase();
    let result = this.pedidos().filter(p => p.status !== 'entregado');
    
    if (estado !== 'todos') {
      result = result.filter(p => p.status === estado);
    }
    
    if (search) {
      result = result.filter(p => 
        p.orderNumber?.toLowerCase().includes(search) ||
        p.id.toString().includes(search)
      );
    }
    
    return result;
  });

  ngOnInit(): void {
    this.loadOrders();
    
    // Escuchar evento de recarga desde notificaciones
    if (isPlatformBrowser(this.platformId)) {
      window.addEventListener('ll-reload-pedidos', this.reloadHandler);
    }
  }

  ngOnDestroy(): void {
    if (isPlatformBrowser(this.platformId)) {
      window.removeEventListener('ll-reload-pedidos', this.reloadHandler);
    }
  }

  loadOrders(): void {
    this.loading.set(true);
    this.error.set(null);

    // Siempre cargar TODOS los pedidos para tener conteos correctos
    this.orderService.listAllOrders(500, 0, undefined, undefined).subscribe({
      next: (response) => {
        this.pedidos.set(response.items || []);
        this.total.set(response.total);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Error cargando pedidos:', err);
        this.error.set('Error al cargar los pedidos');
        this.loading.set(false);
      }
    });
  }

  onSearch(): void {
    // El filtrado se hace en el computed, solo limpiar selección
    this.selectedPedido.set(null);
  }

  openDetail(pedido: Order): void {
    this.selectedPedido.set(pedido);
  }

  closeDetail(): void {
    this.selectedPedido.set(null);
  }

  setFilter(estado: OrderStatus | 'todos'): void {
    this.filterEstado.set(estado);
    this.selectedPedido.set(null);
    // No recargamos, el filtrado es reactivo con computed
  }

  cambiarEstado(pedido: Order, nuevoEstado: OrderStatus): void {
    // Si es cancelación, mostrar modal de confirmación
    if (nuevoEstado === 'cancelado') {
      this.openCancelModal(pedido);
      return;
    }
    
    this.ejecutarCambioEstado(pedido, nuevoEstado);
  }

  private ejecutarCambioEstado(pedido: Order, nuevoEstado: OrderStatus): void {
    this.loading.set(true);
    this.orderService.updateOrderStatus(pedido.id, { status: nuevoEstado }).subscribe({
      next: (updated) => {
        this.pedidos.update(list =>
          list.map(p => p.id === updated.id ? updated : p)
        );
        if (this.selectedPedido()?.id === updated.id) {
          this.selectedPedido.set(updated);
        }
        this.loading.set(false);
        
        // Emitir evento para actualizar inventario cuando se cancela o entrega
        if ((nuevoEstado === 'cancelado' || nuevoEstado === 'entregado') && isPlatformBrowser(this.platformId)) {
          window.dispatchEvent(new CustomEvent('ll-reload-inventory'));
        }
      },
      error: (err) => {
        console.error('Error actualizando estado:', err);
        this.error.set('Error al actualizar el estado del pedido');
        this.loading.set(false);
      }
    });
  }

  // Modal de confirmación de cancelación
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
    this.orderService.updateOrderStatus(pedido.id, { status: 'cancelado' }).subscribe({
      next: (updated) => {
        this.pedidos.update(list =>
          list.map(p => p.id === updated.id ? updated : p)
        );
        if (this.selectedPedido()?.id === updated.id) {
          this.selectedPedido.set(updated);
        }
        this.cancelling.set(false);
        this.closeCancelModal();
        
        // Emitir evento para actualizar inventario
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

  getSiguienteEstado(estado: OrderStatus): OrderStatus | null {
    switch (estado) {
      case 'solicitado': return 'preparando';
      case 'preparando': return 'enviado';
      case 'enviado': return 'entregado';
      default: return null;
    }
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return new Intl.DateTimeFormat('es-MX', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
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

  getConteoEstado(estado: OrderStatus): number {
    return this.pedidos().filter(p => p.status === estado).length;
  }

  // Total de pedidos activos (sin entregados)
  getTotalPedidosActivos(): number {
    return this.pedidos().filter(p => p.status !== 'entregado').length;
  }

  downloadComprobante(pedido: Order): void {
    this.pdfService.generateOrderReceipt(pedido);
  }
}
