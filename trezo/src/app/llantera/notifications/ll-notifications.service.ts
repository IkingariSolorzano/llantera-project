import { inject, Injectable, signal } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, interval, switchMap, tap } from 'rxjs';
import { API_BASE_URL } from '../../core/config/api.config';

export interface LlNotification {
  id: number;
  userId: string;
  type: string;
  title: string;
  message: string;
  data?: unknown;
  read: boolean;
  createdAt: string;
}

interface NotificationListResponse {
  items: LlNotification[];
  total: number;
  limit: number;
  offset: number;
}

@Injectable({ providedIn: 'root' })
export class LlNotificationsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  // Signal para el conteo de no leídas
  readonly unreadCount = signal(0);

  list(limit = 50, offset = 0, unreadOnly = false): Observable<NotificationListResponse> {
    let params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    if (unreadOnly) {
      params = params.set('unread', 'true');
    }

    return this.http.get<NotificationListResponse>(`${this.baseUrl}/notifications/`, { params });
  }

  countUnread(): Observable<{ count: number }> {
    return this.http.get<{ count: number }>(`${this.baseUrl}/notifications/count`).pipe(
      tap(response => this.unreadCount.set(response.count))
    );
  }

  markAsRead(id: number): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/notifications/${id}/read`, {}).pipe(
      tap(() => this.unreadCount.update(c => Math.max(0, c - 1)))
    );
  }

  markAllAsRead(): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/notifications/read-all`, {}).pipe(
      tap(() => this.unreadCount.set(0))
    );
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/notifications/${id}`);
  }

  // Inicia polling para actualizar conteo cada 30 segundos
  startPolling(): Observable<{ count: number }> {
    return interval(30000).pipe(
      switchMap(() => this.countUnread())
    );
  }

  // Helper para obtener icono según tipo
  getIcon(type: string): string {
    const icons: Record<string, string> = {
      order_created: 'shopping_cart',
      order_updated: 'sync',
      order_shipped: 'local_shipping',
      order_delivered: 'check_circle',
      order_cancelled: 'cancel',
      invoice_ready: 'receipt_long',
      general: 'notifications',
    };
    return icons[type] || 'notifications';
  }

  // Helper para formato de tiempo relativo
  getRelativeTime(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Ahora';
    if (diffMins < 60) return `Hace ${diffMins} min`;
    if (diffHours < 24) return `Hace ${diffHours} h`;
    if (diffDays < 7) return `Hace ${diffDays} días`;
    
    return date.toLocaleDateString('es-MX', { day: '2-digit', month: 'short' });
  }
}
