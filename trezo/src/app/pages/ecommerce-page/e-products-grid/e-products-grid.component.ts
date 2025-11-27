import { Component, HostListener, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgIf, CommonModule } from '@angular/common';

import {
  LlLlantasCatalogService,
  LlTireCatalogItemDTO,
} from '../../../llantera/llantas/ll-llantas-catalog.service';

@Component({
    selector: 'app-e-products-grid',
    standalone: true,
    imports: [RouterLink, NgIf, CommonModule],
    templateUrl: './e-products-grid.component.html',
    styleUrl: './e-products-grid.component.scss'
})
export class EProductsGridComponent {

    private readonly catalogService = inject(LlLlantasCatalogService);

    readonly items = signal<LlTireCatalogItemDTO[]>([]);
    readonly loading = signal(false);
    readonly error = signal<string | null>(null);
    readonly genericImage = 'images/llantera/generic-tyre.png';

    constructor() {
        this.loadCatalog();
    }

    // Card Header Menu
    isCardHeaderOpen = false;
    toggleCardHeaderMenu() {
        this.isCardHeaderOpen = !this.isCardHeaderOpen;
    }
    @HostListener('document:click', ['$event'])
    handleClickOutside(event: Event) {
        const target = event.target as HTMLElement;
        if (!target.closest('.trezo-card-dropdown')) {
            this.isCardHeaderOpen = false;
        }
    }

    private loadCatalog(): void {
        this.loading.set(true);
        this.error.set(null);

        this.catalogService
            .list({ limit: 24, offset: 0, level: 'public' })
            .subscribe({
                next: (items) => {
                    this.items.set(items);
                },
                error: (err) => {
                    console.error('Error al cargar catálogo de llantas', err);
                    this.error.set('No se pudo cargar el catálogo de llantas.');
                    this.loading.set(false);
                },
                complete: () => {
                    this.loading.set(false);
                },
            });
    }

    getTireImage(item: LlTireCatalogItemDTO): string {
        const url = item.tire.urlImagen?.trim();
        return url ? url : this.genericImage;
    }

}