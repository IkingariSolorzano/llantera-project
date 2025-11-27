import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, inject, signal, OnInit, OnDestroy, PLATFORM_ID } from '@angular/core';
import { Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { LlNotificationsService, LlNotification } from './ll-notifications.service';
import { LlAuthService } from '../auth/ll-auth.service';

@Component({
  selector: 'app-ll-notification-bell',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="relative">
      <!-- Botón de campana -->
      <button
        type="button"
        (click)="togglePanel()"
        class="relative p-2 rounded-lg text-gray-500 hover:bg-gray-100 dark:hover:bg-[#15203c] transition-colors"
      >
        <i class="material-symbols-outlined !text-[22px]">notifications</i>
        @if (notificationsService.unreadCount() > 0) {
          <span class="absolute top-0 right-0 w-5 h-5 bg-[#E60012] text-white text-[10px] font-bold rounded-full flex items-center justify-center">
            {{ notificationsService.unreadCount() > 99 ? '99+' : notificationsService.unreadCount() }}
          </span>
        }
      </button>

      <!-- Panel de notificaciones -->
      @if (showPanel()) {
        <div class="absolute right-0 top-12 w-80 bg-white dark:bg-[#0c1427] rounded-xl shadow-2xl border border-gray-200 dark:border-[#172036] z-50 overflow-hidden">
          <!-- Header -->
          <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-[#172036]">
            <h3 class="font-semibold text-gray-900 dark:text-white">Notificaciones</h3>
            @if (notificationsService.unreadCount() > 0) {
              <button
                type="button"
                (click)="markAllRead()"
                class="text-xs text-[#E60012] hover:underline"
              >
                Marcar todas como leídas
              </button>
            }
          </div>

          <!-- Lista -->
          <div class="max-h-96 overflow-y-auto">
            @if (loading()) {
              <div class="flex items-center justify-center py-8">
                <div class="w-6 h-6 border-2 border-[#E60012] border-t-transparent rounded-full animate-spin"></div>
              </div>
            } @else if (notifications().length === 0) {
              <div class="py-8 text-center text-gray-500 dark:text-gray-400">
                <i class="material-symbols-outlined !text-[36px] opacity-50 mb-2">notifications_off</i>
                <p class="text-sm">No tienes notificaciones</p>
              </div>
            } @else {
              @for (notification of notifications(); track notification.id) {
                <div
                  class="p-3 border-b border-gray-100 dark:border-[#172036] hover:bg-gray-50 dark:hover:bg-[#0a0f1a] cursor-pointer transition-colors"
                  [ngClass]="{'bg-blue-50 dark:bg-blue-900/10': !notification.read}"
                  (click)="handleNotificationClick(notification)"
                >
                  <div class="flex gap-3">
                    <div class="flex-shrink-0">
                      <div class="w-10 h-10 rounded-full bg-gray-100 dark:bg-[#15203c] flex items-center justify-center">
                        <i class="material-symbols-outlined !text-[18px] text-gray-500">
                          {{ notificationsService.getIcon(notification.type) }}
                        </i>
                      </div>
                    </div>
                    <div class="flex-1 min-w-0">
                      <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
                        {{ notification.title }}
                      </p>
                      <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
                        {{ notification.message }}
                      </p>
                      <p class="text-xs text-gray-400 mt-1">
                        {{ notificationsService.getRelativeTime(notification.createdAt) }}
                      </p>
                    </div>
                    @if (!notification.read) {
                      <div class="flex-shrink-0">
                        <div class="w-2 h-2 rounded-full bg-[#E60012]"></div>
                      </div>
                    }
                  </div>
                </div>
              }
            }
          </div>
        </div>
      }
    </div>

    <!-- Overlay para cerrar -->
    @if (showPanel()) {
      <div 
        class="fixed inset-0 z-40"
        (click)="togglePanel()"
      ></div>
    }
  `,
  styles: [`
    :host {
      display: block;
    }
  `]
})
export class LlNotificationBellComponent implements OnInit, OnDestroy {
  readonly notificationsService = inject(LlNotificationsService);
  private readonly router = inject(Router);
  private readonly auth = inject(LlAuthService);
  private readonly platformId = inject(PLATFORM_ID);
  
  readonly showPanel = signal(false);
  readonly loading = signal(false);
  readonly notifications = signal<LlNotification[]>([]);

  private pollingSub?: Subscription;

  ngOnInit(): void {
    // Cargar conteo inicial
    this.notificationsService.countUnread().subscribe();

    // Iniciar polling
    this.pollingSub = this.notificationsService.startPolling().subscribe();
  }

  ngOnDestroy(): void {
    this.pollingSub?.unsubscribe();
  }

  togglePanel(): void {
    const isOpen = !this.showPanel();
    this.showPanel.set(isOpen);
    
    if (isOpen) {
      this.loadNotifications();
    }
  }

  loadNotifications(): void {
    this.loading.set(true);
    this.notificationsService.list(20, 0).subscribe({
      next: (response) => {
        this.notifications.set(response.items || []);
        this.loading.set(false);
      },
      error: () => {
        this.loading.set(false);
      }
    });
  }

  markRead(notification: LlNotification): void {
    if (notification.read) return;
    
    this.notificationsService.markAsRead(notification.id).subscribe({
      next: () => {
        this.notifications.update(list =>
          list.map(n => n.id === notification.id ? { ...n, read: true } : n)
        );
      }
    });
  }

  markAllRead(): void {
    this.notificationsService.markAllAsRead().subscribe({
      next: () => {
        this.notifications.update(list =>
          list.map(n => ({ ...n, read: true }))
        );
      }
    });
  }

  handleNotificationClick(notification: LlNotification): void {
    // Marcar como leída
    this.markRead(notification);
    
    // Cerrar el panel
    this.showPanel.set(false);
    
    // Navegar según el tipo de notificación y rol del usuario
    this.navigateToNotification(notification);
  }

  private navigateToNotification(notification: LlNotification): void {
    const isAdmin = this.auth.user()?.role === 'admin';
    const type = notification.type;

    // Notificaciones relacionadas con pedidos
    if (type === 'order_created' || type === 'order_updated' || 
        type === 'order_shipped' || type === 'order_delivered' || 
        type === 'order_cancelled') {
      
      if (isAdmin) {
        // Admin va a la gestión de pedidos
        this.router.navigate(['/dashboard/pedidos']).then(() => {
          // Emitir evento para recargar pedidos (útil si ya está en la pantalla)
          if (isPlatformBrowser(this.platformId)) {
            window.dispatchEvent(new CustomEvent('ll-reload-pedidos'));
          }
        });
      } else {
        // Cliente va a sus pedidos (emitir evento para cambiar tab)
        if (isPlatformBrowser(this.platformId)) {
          window.dispatchEvent(new CustomEvent('ll-change-tab', { detail: 'pedidos' }));
        }
      }
      return;
    }

    // Notificaciones de factura
    if (type === 'invoice_ready') {
      if (isAdmin) {
        this.router.navigate(['/dashboard/ventas']);
      } else {
        if (isPlatformBrowser(this.platformId)) {
          window.dispatchEvent(new CustomEvent('ll-change-tab', { detail: 'compras' }));
        }
      }
      return;
    }
  }
}
