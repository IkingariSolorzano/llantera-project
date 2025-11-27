import { Injectable, Inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';

@Injectable({
  providedIn: 'root'
})
export class ToggleService {
    private readonly themeKey = 'trezo_theme';
    private readonly directionKey = 'trezo_direction';
    private isBrowser: boolean;

    constructor(@Inject(PLATFORM_ID) private platformId: Object) {
        this.isBrowser = isPlatformBrowser(this.platformId);  // Check if in browser
    }

    // Initialize theme and direction based on saved preferences
    initializeTheme() {
        if (this.isBrowser) {
            const savedTheme = localStorage.getItem(this.themeKey) || 'light';
            const savedDirection = localStorage.getItem(this.directionKey) || 'ltr';

            this.applyTheme(savedTheme);
            this.applyDirection(savedDirection);
        }
    }

    toggleTheme() {
        if (!this.isBrowser) return;

        const currentTheme = localStorage.getItem(this.themeKey) === 'dark' ? 'light' : 'dark';
        this.applyTheme(currentTheme);
        localStorage.setItem(this.themeKey, currentTheme);
    }

    toggleDirection() {
        if (!this.isBrowser) return;

        const currentDirection = localStorage.getItem(this.directionKey) === 'rtl' ? 'ltr' : 'rtl';
        this.applyDirection(currentDirection);
        localStorage.setItem(this.directionKey, currentDirection);
    }

    private applyTheme(theme: string) {
        if (this.isBrowser) {
            const htmlElement = document.documentElement;
            if (theme === 'dark') {
                htmlElement.classList.add('dark');
                htmlElement.classList.remove('light');
            } else {
                htmlElement.classList.add('light');
                htmlElement.classList.remove('dark');
            }
        }
    }

    private applyDirection(direction: string) {
        if (this.isBrowser) {
            document.documentElement.setAttribute('dir', direction);
        }
    }
}