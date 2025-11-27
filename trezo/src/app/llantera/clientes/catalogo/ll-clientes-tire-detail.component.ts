import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';

import {
  LlLlantasCatalogService,
  LlTireCatalogItemDTO,
  LlTireDTO,
} from '../../llantas/ll-llantas-catalog.service';

@Component({
  selector: 'app-ll-clientes-tire-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './ll-clientes-tire-detail.component.html',
  styleUrl: './ll-clientes-tire-detail.component.scss',
})
export class LlClientesTireDetailComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly catalogService = inject(LlLlantasCatalogService);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly item = signal<LlTireCatalogItemDTO | null>(null);
  readonly genericImage = 'images/llantera/generic-tyre.png';

  constructor() {
    const fromState = (history.state as { item?: LlTireCatalogItemDTO }).item;
    if (fromState) {
      this.item.set(fromState);
    }

    const sku = this.route.snapshot.paramMap.get('sku');
    if (!fromState && sku) {
      this.loadFromApi(sku);
    }
  }

  private loadFromApi(sku: string): void {
    this.loading.set(true);
    this.error.set(null);

    this.catalogService
      .list({ limit: 1, offset: 0, level: 'public', search: sku })
      .subscribe({
        next: (items) => {
          const found = items && items.length ? items[0] : null;
          if (!found) {
            this.error.set('No se encontró la llanta solicitada.');
          }
          this.item.set(found);
        },
        error: (err) => {
          console.error('Error al cargar detalle de llanta', err);
          this.error.set('No se pudo cargar el detalle de la llanta.');
          this.loading.set(false);
        },
        complete: () => {
          this.loading.set(false);
        },
      });
  }

  getTireImage(item: LlTireCatalogItemDTO | null): string {
    if (!item) {
      return this.genericImage;
    }
    const url = item.tire.urlImagen?.trim();
    return url ? url : this.genericImage;
  }

  formatMedida(tire: LlTireDTO | null | undefined): string {
    if (!tire) {
      return 'N/D';
    }
    if (tire.medidaOriginal && tire.medidaOriginal.trim()) {
      return tire.medidaOriginal;
    }

    const ancho = tire.ancho ? String(tire.ancho) : '';
    const perfil = typeof tire.perfil === 'number' ? `/${tire.perfil}` : '';
    const rinValue = typeof tire.rin === 'number' ? tire.rin : null;
    const rin =
      rinValue !== null
        ? ` R${Number.isInteger(rinValue) ? rinValue.toFixed(0) : rinValue}`
        : '';

    const value = `${ancho}${perfil}${rin}`.trim();
    return value || 'N/D';
  }

  formatCargaVelocidad(tire: LlTireDTO | null | undefined): string {
    if (!tire) {
      return 'N/D';
    }
    const carga = (tire.indiceCarga || '').trim();
    const vel = (tire.indiceVelocidad || '').trim();

    if (!carga && !vel) {
      return 'N/D';
    }
    if (!carga) {
      return vel;
    }
    if (!vel) {
      return carga;
    }
    return `${carga}/${vel}`;
  }

  formatUso(tire: LlTireDTO | null | undefined): string {
    if (!tire) {
      return 'N/D';
    }
    return tire.abreviaturaUso || 'N/D';
  }
}
