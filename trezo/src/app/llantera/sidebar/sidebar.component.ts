import { Component, inject, Renderer2, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { NgScrollbarModule } from 'ngx-scrollbar';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { LlAuthService } from '../auth/ll-auth.service';

@Component({
    selector: 'app-llantera-sidebar',
    standalone: true,
    imports: [NgScrollbarModule, RouterLink, RouterLinkActive],
    templateUrl: './sidebar.component.html',
    styleUrls: ['./sidebar.component.scss']
})
export class LlanteraSidebarComponent {
    private readonly auth = inject(LlAuthService);
    private readonly renderer = inject(Renderer2);
    private readonly platformId = inject(PLATFORM_ID);

    get role(): string {
        const user = this.auth.user();
        return (user?.role ?? '').toString().toLowerCase();
    }

    get isAdmin(): boolean {
        return this.role === 'admin';
    }

    get isEmployee(): boolean {
        return this.role === 'employee';
    }

    closeSidebar(): void {
        if (!isPlatformBrowser(this.platformId)) return;
        this.renderer.removeClass(document.body, 'sidebar-hidden');
    }
}
