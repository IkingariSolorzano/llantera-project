import { Component, Renderer2, PLATFORM_ID, inject } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { FooterComponent } from '../common/footer/footer.component';
import { LlanteraSidebarComponent } from '../llantera/sidebar/sidebar.component';
import { LlHeaderComponent } from '../llantera/header/ll-header.component';

@Component({
    selector: 'app-dashboard',
    standalone: true,
    imports: [RouterOutlet, LlHeaderComponent, LlanteraSidebarComponent, FooterComponent],
    templateUrl: './dashboard.component.html',
    styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent {
    private readonly renderer = inject(Renderer2);
    private readonly platformId = inject(PLATFORM_ID);

    closeSidebar(): void {
        if (!isPlatformBrowser(this.platformId)) return;
        this.renderer.removeClass(document.body, 'sidebar-hidden');
    }
}