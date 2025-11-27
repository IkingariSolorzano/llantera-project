import { Component, Inject, PLATFORM_ID, Renderer2 } from '@angular/core';
import { RouterLink } from '@angular/router';
import { isPlatformBrowser } from '@angular/common';
import { ToggleService } from './toggle.service';

@Component({
    selector: 'app-header',
    standalone: true,
    imports: [RouterLink],
    templateUrl: './header.component.html',
    styleUrl: './header.component.scss'
})
export class HeaderComponent {

    isSidebarVisible = true;  // Track sidebar visibility

    constructor(
        @Inject(PLATFORM_ID) private platformId: Object,
        public toggleService: ToggleService,
        private renderer: Renderer2
    ) {}

    ngOnInit(): void {
        // Initialize theme and direction on component load
        this.toggleService.initializeTheme();
    }

    // Toggle theme between light and dark
    toggleTheme() {
        this.toggleService.toggleTheme();
    }

    // Toggle direction between LTR and RTL
    toggleDirection() {
        this.toggleService.toggleDirection();
    }

    // Toggle boyd class between sidebar-hidden and sidebar-show
    toggleSidebar() {
        this.isSidebarVisible = !this.isSidebarVisible;
        // Add or remove the 'sidebar-hidden' class on the <body> element
        if (this.isSidebarVisible) {
            this.renderer.removeClass(document.body, 'sidebar-hidden');
        } else {
            this.renderer.addClass(document.body, 'sidebar-hidden');
        }
    }

    // Toogle Class
    buttonStates: { [key: string]: boolean } = {
        connectedAppsMenuBtn: false,
        languageMenuButton: false,
        notificationsMenuBtn: false,
        profileMenuBtn: false,
        settingsMenuBtn: false
    };
    toggleClass(buttonId: string) {
        // Check if the clicked button is already active
        const isCurrentlyActive = this.buttonStates[buttonId];
        // Set all buttons to inactive
        for (const key in this.buttonStates) {
            this.buttonStates[key] = false;
        }
        // Toggle the clicked button based on its previous state
        this.buttonStates[buttonId] = !isCurrentlyActive;
    }

    // Fullscreen
    isFullscreen: boolean = false;
    ngAfterViewInit() {
        if (isPlatformBrowser(this.platformId)) {
            // Only add event listeners if the platform is the browser
            document.addEventListener('fullscreenchange', this.onFullscreenChange.bind(this));
            document.addEventListener('webkitfullscreenchange', this.onFullscreenChange.bind(this));
            document.addEventListener('mozfullscreenchange', this.onFullscreenChange.bind(this));
            document.addEventListener('MSFullscreenChange', this.onFullscreenChange.bind(this));
        }
    }
    toggleFullscreen() {
        if (this.isFullscreen) {
            this.closeFullscreen();
        } else {
            this.openFullscreen();
        }
    }
    openFullscreen() {
        if (isPlatformBrowser(this.platformId)) {
            const element = document.documentElement as HTMLElement & {
                mozRequestFullScreen?: () => Promise<void>;
                webkitRequestFullscreen?: () => Promise<void>;
                msRequestFullscreen?: () => Promise<void>;
            };
            if (element.requestFullscreen) {
                element.requestFullscreen();
            } else if (element.mozRequestFullScreen) { // Firefox
                element.mozRequestFullScreen();
            } else if (element.webkitRequestFullscreen) { // Chrome, Safari, and Opera
                element.webkitRequestFullscreen();
            } else if (element.msRequestFullscreen) { // IE/Edge
                element.msRequestFullscreen();
            }
        }
    }
    closeFullscreen() {
        if (isPlatformBrowser(this.platformId)) {
            const doc = document as Document & {
                mozCancelFullScreen?: () => Promise<void>;
                webkitExitFullscreen?: () => Promise<void>;
                msExitFullscreen?: () => Promise<void>;
            };
            if (document.exitFullscreen) {
                document.exitFullscreen();
            } else if (doc.mozCancelFullScreen) { // Firefox
                doc.mozCancelFullScreen();
            } else if (doc.webkitExitFullscreen) { // Chrome, Safari, and Opera
                doc.webkitExitFullscreen();
            } else if (doc.msExitFullscreen) { // IE/Edge
                doc.msExitFullscreen();
            }
        }
    }
    onFullscreenChange() {
        if (isPlatformBrowser(this.platformId)) {
            const doc = document as Document & {
                webkitFullscreenElement?: Element;
                mozFullScreenElement?: Element;
                msFullscreenElement?: Element;
            };
            this.isFullscreen = !!(document.fullscreenElement || doc.webkitFullscreenElement || doc.mozFullScreenElement || doc.msFullscreenElement);
        }
    }

}