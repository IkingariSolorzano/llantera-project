import { ApplicationConfig, provideZoneChangeDetection } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withFetch, withInterceptors } from '@angular/common/http';

import { routes } from './app.routes';
import { provideClientHydration } from '@angular/platform-browser';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { API_BASE_URL, DEFAULT_API_BASE_URL } from './core/config/api.config';
import { llAuthInterceptor } from './llantera/auth/ll-auth.interceptor';

export const appConfig: ApplicationConfig = {
    providers: [
        provideZoneChangeDetection({ eventCoalescing: true }),
        provideRouter(routes),
        provideClientHydration(),
        provideAnimationsAsync(),
        provideHttpClient(
            withFetch(),
            withInterceptors([llAuthInterceptor]),
        ),
        { provide: API_BASE_URL, useValue: DEFAULT_API_BASE_URL },
    ]
};