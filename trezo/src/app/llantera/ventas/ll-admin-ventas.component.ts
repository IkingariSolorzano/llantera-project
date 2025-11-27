import { CommonModule } from '@angular/common';
import { Component, inject, signal, computed, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { LlOrderService, Order, OrderStatus } from '../clientes/services/ll-order.service';

@Component({
  selector: 'app-ll-admin-ventas',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './ll-admin-ventas.component.html',
})
export class LlAdminVentasComponent implements OnInit {
  private readonly orderService = inject(LlOrderService);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly selectedVenta = signal<Order | null>(null);
  readonly filterFacturada = signal<'todas' | 'facturadas' | 'sin_factura'>('todas');
  readonly searchTerm = signal('');
  readonly ventas = signal<Order[]>([]);
  readonly uploadingInvoice = signal(false);
  readonly showUploadModal = signal(false);
  readonly uploadVenta = signal<Order | null>(null);

  // Solo mostrar pedidos entregados (ventas completadas)
  readonly filteredVentas = computed(() => {
    const filter = this.filterFacturada();
    let result = this.ventas().filter(v => v.status === 'entregado');
    
    if (filter === 'facturadas') {
      result = result.filter(v => v.requiresInvoice && v.invoicePdfPath);
    } else if (filter === 'sin_factura') {
      result = result.filter(v => v.requiresInvoice && !v.invoicePdfPath);
    }
    
    return result;
  });

  ngOnInit(): void {
    this.loadVentas();
  }

  loadVentas(): void {
    this.loading.set(true);
    this.error.set(null);

    // Cargar solo pedidos entregados
    this.orderService.listAllOrders(100, 0, 'entregado', this.searchTerm() || undefined).subscribe({
      next: (response) => {
        this.ventas.set(response.items || []);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Error cargando ventas:', err);
        this.error.set('Error al cargar las ventas');
        this.loading.set(false);
      }
    });
  }

  onSearch(): void {
    this.loadVentas();
  }

  openDetail(venta: Order): void {
    this.selectedVenta.set(venta);
  }

  closeDetail(): void {
    this.selectedVenta.set(null);
  }

  setFilter(filter: 'todas' | 'facturadas' | 'sin_factura'): void {
    this.filterFacturada.set(filter);
  }

  openUploadModal(venta: Order): void {
    this.uploadVenta.set(venta);
    this.showUploadModal.set(true);
  }

  closeUploadModal(): void {
    this.showUploadModal.set(false);
    this.uploadVenta.set(null);
  }

  onFileSelected(event: Event, tipo: 'xml' | 'pdf'): void {
    const input = event.target as HTMLInputElement;
    if (!input.files?.length) return;
    
    const file = input.files[0];
    const venta = this.uploadVenta();
    if (!venta) return;

    this.uploadingInvoice.set(true);
    
    const xmlFile = tipo === 'xml' ? file : undefined;
    const pdfFile = tipo === 'pdf' ? file : undefined;

    this.orderService.uploadInvoice(venta.id, xmlFile, pdfFile).subscribe({
      next: (updatedOrder) => {
        this.uploadingInvoice.set(false);
        // Actualizar la venta en la lista
        this.ventas.update(list => 
          list.map(v => v.id === updatedOrder.id ? updatedOrder : v)
        );
        // Actualizar venta seleccionada si está abierta
        if (this.selectedVenta()?.id === updatedOrder.id) {
          this.selectedVenta.set(updatedOrder);
        }
        this.uploadVenta.set(updatedOrder);
        input.value = ''; // Reset input
      },
      error: (err) => {
        console.error('Error subiendo factura:', err);
        this.error.set('Error al subir el archivo');
        this.uploadingInvoice.set(false);
        input.value = '';
      }
    });
  }

  descargarFactura(venta: Order, tipo: 'xml' | 'pdf'): void {
    const path = tipo === 'xml' ? venta.invoiceXmlPath : venta.invoicePdfPath;
    if (!path) {
      alert(`No hay archivo ${tipo.toUpperCase()} disponible`);
      return;
    }
    const filename = `factura_${venta.orderNumber}.${tipo}`;
    this.orderService.downloadInvoiceFile(path, filename);
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

  formatDireccion(venta: Order): string {
    const addr = venta.shippingAddress;
    if (!addr) return 'Sin dirección';
    let dir = `${addr.street} #${addr.exteriorNumber}`;
    if (addr.interiorNumber) dir += ` Int. ${addr.interiorNumber}`;
    dir += `, ${addr.neighborhood}, C.P. ${addr.postalCode}, ${addr.city}, ${addr.state}`;
    return dir;
  }

  getMetodoPagoLabel(metodo: string): string {
    return this.orderService.getPaymentMethodLabel(metodo as any);
  }

  getConteoFacturadas(): number {
    return this.ventas().filter(v => v.status === 'entregado' && v.requiresInvoice && v.invoicePdfPath).length;
  }

  getConteoSinFactura(): number {
    return this.ventas().filter(v => v.status === 'entregado' && v.requiresInvoice && !v.invoicePdfPath).length;
  }

  hasFactura(venta: Order): boolean {
    return !!(venta.invoicePdfPath || venta.invoiceXmlPath);
  }
}
