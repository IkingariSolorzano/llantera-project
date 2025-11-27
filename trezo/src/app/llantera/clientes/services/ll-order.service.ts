import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DEFAULT_API_BASE_URL } from '../../../core/config/api.config';

const API_BASE_URL = DEFAULT_API_BASE_URL;

export type OrderStatus = 'solicitado' | 'preparando' | 'enviado' | 'entregado' | 'cancelado';
export type PaymentMethod = 'transferencia' | 'tarjeta' | 'efectivo';
export type PaymentMode = 'contado' | 'credito' | 'parcialidades' | 'anticipo';

export interface OrderItem {
  id: number;
  orderId: number;
  tireSku: string;
  tireMeasure: string;
  tireBrand?: string;
  tireModel?: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
  createdAt: string;
}

export interface ShippingAddress {
  id?: number;
  street: string;
  exteriorNumber: string;
  interiorNumber?: string;
  neighborhood: string;
  postalCode: string;
  city: string;
  state: string;
  reference?: string;
  phone: string;
}

export interface BillingInfo {
  id?: number;
  rfc: string;
  razonSocial: string;
  regimenFiscal: string;
  usoCfdi: string;
  postalCode: string;
  email?: string;
}

export interface Order {
  id: number;
  orderNumber: string;
  userId: number;
  status: OrderStatus;
  shippingAddress: ShippingAddress;
  paymentMethod: PaymentMethod;
  paymentMode: PaymentMode;
  paymentInstallments?: number;
  paymentNotes?: string;
  requiresInvoice: boolean;
  billingInfo?: BillingInfo;
  items: OrderItem[];
  subtotal: number;
  iva: number;
  shippingCost: number;
  total: number;
  invoiceXmlPath?: string;
  invoicePdfPath?: string;
  customerNotes?: string;
  adminNotes?: string;
  createdAt: string;
  updatedAt: string;
  shippedAt?: string;
  deliveredAt?: string;
  cancelledAt?: string;
}

export interface CreateOrderItemRequest {
  tireSku: string;
  tireMeasure: string;
  tireBrand?: string;
  tireModel?: string;
  quantity: number;
  unitPrice: number;
}

export interface CreateOrderRequest {
  items: CreateOrderItemRequest[];
  shippingAddress: ShippingAddress;
  paymentMethod: PaymentMethod;
  paymentMode: PaymentMode;
  paymentInstallments?: number;
  paymentNotes?: string;
  requiresInvoice: boolean;
  billingInfo?: BillingInfo;
  customerNotes?: string;
  /** Subtotal sin IVA */
  subtotal?: number;
  /** IVA (16%) */
  iva?: number;
  /** Total con IVA */
  total?: number;
}

export interface OrderListResponse {
  items: Order[];
  total: number;
  limit: number;
  offset: number;
}

export interface UpdateStatusRequest {
  status: OrderStatus;
  adminNotes?: string;
}

@Injectable({
  providedIn: 'root'
})
export class LlOrderService {
  private readonly http = inject(HttpClient);
  // Nota: El backend espera trailing slash
  private readonly baseUrl = `${API_BASE_URL}/orders/`;
  private readonly adminBaseUrl = `${API_BASE_URL}/admin/orders/`;

  // Operaciones de cliente
  listMyOrders(limit = 20, offset = 0): Observable<OrderListResponse> {
    const params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());
    return this.http.get<OrderListResponse>(this.baseUrl, { params });
  }

  getMyOrder(id: number): Observable<Order> {
    return this.http.get<Order>(`${this.baseUrl}${id}`);
  }

  createOrder(request: CreateOrderRequest): Observable<Order> {
    return this.http.post<Order>(this.baseUrl, request);
  }

  /** Cancelar pedido propio del cliente (solo si está en estado 'solicitado') */
  cancelMyOrder(id: number): Observable<Order> {
    return this.http.patch<Order>(`${this.baseUrl}${id}/status`, { status: 'cancelado' });
  }

  // Operaciones de admin
  listAllOrders(limit = 20, offset = 0, status?: OrderStatus, search?: string): Observable<OrderListResponse> {
    let params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());
    
    if (status) {
      params = params.set('status', status);
    }
    if (search) {
      params = params.set('search', search);
    }

    return this.http.get<OrderListResponse>(this.adminBaseUrl, { params });
  }

  getOrder(id: number): Observable<Order> {
    return this.http.get<Order>(`${this.adminBaseUrl}${id}`);
  }

  updateOrderStatus(id: number, request: UpdateStatusRequest): Observable<Order> {
    return this.http.patch<Order>(`${this.adminBaseUrl}${id}`, request);
  }

  uploadInvoice(orderId: number, xmlFile?: File, pdfFile?: File): Observable<Order> {
    const formData = new FormData();
    if (xmlFile) {
      formData.append('xml', xmlFile);
    }
    if (pdfFile) {
      formData.append('pdf', pdfFile);
    }
    return this.http.post<Order>(`${this.adminBaseUrl}${orderId}/invoice`, formData);
  }

  downloadInvoiceFile(filePath: string, filename: string): void {
    // Descargar archivo con autenticación usando HttpClient
    this.http.get(`${API_BASE_URL}/files/${filePath}`, { responseType: 'blob' }).subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      },
      error: (err) => {
        console.error('Error descargando archivo:', err);
        alert('Error al descargar el archivo');
      }
    });
  }

  // Helpers
  getStatusLabel(status: OrderStatus): string {
    const labels: Record<OrderStatus, string> = {
      solicitado: 'Solicitado',
      preparando: 'Preparando',
      enviado: 'Enviado',
      entregado: 'Entregado',
      cancelado: 'Cancelado',
    };
    return labels[status] || status;
  }

  getPaymentMethodLabel(method: PaymentMethod): string {
    const labels: Record<PaymentMethod, string> = {
      transferencia: 'Transferencia bancaria',
      tarjeta: 'Tarjeta de crédito/débito',
      efectivo: 'Pago en efectivo',
    };
    return labels[method] || method;
  }

  getPaymentModeLabel(mode: PaymentMode): string {
    const labels: Record<PaymentMode, string> = {
      contado: 'Pago de contado',
      credito: 'Crédito',
      parcialidades: 'Pago en parcialidades',
      anticipo: 'Anticipo',
    };
    return labels[mode] || mode;
  }
}
