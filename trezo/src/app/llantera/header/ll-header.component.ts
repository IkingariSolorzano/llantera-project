import { CommonModule } from '@angular/common';
import { Component, PLATFORM_ID, inject, signal, computed } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { Renderer2 } from '@angular/core';

import { ToggleService } from '../../common/header/toggle.service';
import { LlAuthService } from '../auth/ll-auth.service';
import { LlNotificationBellComponent } from '../notifications/ll-notification-bell.component';

export interface HeaderNotification {
  id: string;
  title: string;
  message: string;
  time: string;
  type: 'order' | 'info' | 'warning';
  read: boolean;
}

@Component({
  selector: 'app-ll-header',
  standalone: true,
  imports: [CommonModule, LlNotificationBellComponent],
  templateUrl: './ll-header.component.html',
  styleUrl: './ll-header.component.scss',
})
export class LlHeaderComponent {
  private readonly platformId = inject(PLATFORM_ID);
  private readonly renderer = inject(Renderer2);
  private readonly toggleService = inject(ToggleService);
  private readonly auth = inject(LlAuthService);

  isSidebarVisible = true;
  isDarkMode = false;

  readonly showUserMenu = signal(false);
  readonly showNotifications = signal(false);
  readonly notifications = signal<HeaderNotification[]>([]);
  
  readonly notificationCount = computed(() => 
    this.notifications().filter(n => !n.read).length
  );

  get userName(): string {
    const user = this.auth.user();
    return user?.name || 'Usuario';
  }

  get userEmail(): string {
    const user = this.auth.user();
    return user?.email || '';
  }

  get userInitials(): string {
    const name = this.userName;
    const parts = name.split(' ').filter(p => p.length > 0);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  }

  constructor() {
    this.toggleService.initializeTheme();
    this.updateDarkModeState();
  }

  private updateDarkModeState(): void {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }
    this.isDarkMode = document.documentElement.classList.contains('dark');
  }

  // Toggle sidebar usando la lógica original de Trezo
  toggleSidebar(): void {
    if (!isPlatformBrowser(this.platformId)) {
      return;
    }

    const isOpen = document.body.classList.contains('sidebar-hidden');

    if (isOpen) {
      // Cerrar sidebar (quitar clase)
      this.renderer.removeClass(document.body, 'sidebar-hidden');
      this.isSidebarVisible = false;
    } else {
      // Abrir sidebar (agregar clase)
      this.renderer.addClass(document.body, 'sidebar-hidden');
      this.isSidebarVisible = true;
    }
  }

  toggleTheme(): void {
    this.toggleService.toggleTheme();
    this.updateDarkModeState();
  }

  toggleUserMenu(): void {
    this.showNotifications.set(false);
    this.showUserMenu.update(v => !v);
  }

  toggleNotifications(): void {
    this.showUserMenu.set(false);
    this.showNotifications.update(v => !v);
  }

  closeAllDropdowns(): void {
    this.showUserMenu.set(false);
    this.showNotifications.set(false);
  }

  goToMyAccount(): void {
    this.closeAllDropdowns();
    this.auth.goToMyAccount();
  }

  logout(): void {
    this.closeAllDropdowns();
    this.auth.logout();
  }

  // Método para agregar notificaciones (será usado por WebSocket)
  addNotification(notification: Omit<HeaderNotification, 'id'>): void {
    const newNotif: HeaderNotification = {
      ...notification,
      id: crypto.randomUUID(),
    };
    this.notifications.update(list => [newNotif, ...list]);
  }

  markNotificationAsRead(id: string): void {
    this.notifications.update(list =>
      list.map(n => n.id === id ? { ...n, read: true } : n)
    );
  }

  clearNotifications(): void {
    this.notifications.set([]);
  }
}
