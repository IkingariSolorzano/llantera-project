import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators, FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs/operators';
import { of } from 'rxjs';

import {
  LlLlantasCatalogService,
  LlTireAdminDTO,
  LlTireUpsertPayload,
  ApiListResponse,
} from './ll-llantas-catalog.service';
import { LlLlantasBrandsService, LlTireBrandDTO } from './ll-llantas-brands.service';
import {
  LlLlantasPriceColumnsService,
  LlPriceColumnDTO,
} from './ll-llantas-price-columns.service';

@Component({
  selector: 'app-ll-llantas-form',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, FormsModule, RouterLink],
  templateUrl: './ll-llantas-form.component.html',
  styleUrl: './ll-llantas-form.component.scss',
})
export class LlLlantasFormComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly fb = inject(FormBuilder);
  private readonly catalogService = inject(LlLlantasCatalogService);
  private readonly brandsService = inject(LlLlantasBrandsService);
  private readonly priceColumnsService = inject(LlLlantasPriceColumnsService);
  private readonly destroyRef = inject(DestroyRef);

  sku = '';

  readonly isEditMode = signal(false);

  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal<string | null>(null);

  readonly brands = signal<LlTireBrandDTO[]>([]);
  readonly priceCodes = signal<string[]>([]);
  readonly prices = signal<Record<string, number | null>>({});
  readonly priceColumnsDef = signal<LlPriceColumnDTO[]>([]);

  // Catálogos locales para selects
  readonly construcciones = ['RADIAL', 'DIAGONAL'];

  readonly tiposTubo = ['TL', 'TT']; // Tubeless / Tube Type

  readonly capas = [
    '4PR', '6PR', '8PR', '10PR', '12PR', '14PR', '16PR',
    '18PR', '20PR', '22PR', '24PR' // adicionales comunes
  ];

  // Abreviaturas de uso normalizadas
  readonly usos = [
    { code: 'PS', label: 'Pasajero' },
    { code: 'PSR', label: 'Pasajero Radial' },
    { code: 'LT', label: 'Camioneta (Light Truck)' },
    { code: 'LTR', label: 'Camioneta Radial (Light Truck Radial)' },
    { code: 'LTS', label: 'Camioneta Convencional (Light Truck Standard)' },
    { code: 'ST', label: 'Special Trailer' },
    { code: 'TBR', label: 'Truck & Bus Radial' },
    { code: 'IND', label: 'Industrial' },
    { code: 'AG', label: 'Agrícola' },
    { code: 'OTR', label: 'Fuera de carretera (Off The Road)' },
    { code: 'MC', label: 'Motocicleta' },
    { code: 'ATV', label: 'ATV / UTV' }
  ];

  // Índices de velocidad normalizados UNECE
  readonly indicesVelocidad = [
    'E','F','G',
    'J','K','L','M','N',
    'P','Q','R','S','T',
    'U','H','V','W','Y','Z'
  ];

  // Índices de carga comunes (pueden llegar hasta 200+)
  readonly indicesCarga = [
    70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
    80, 81, 82, 83, 84, 85, 86, 87, 88, 89,
    90, 91, 92, 93, 94, 95, 96, 97, 98, 99,
    100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
    110, 111, 112, 113, 114, 115, 116, 117, 118, 119
  ];

  // Tipos de terreno
  readonly tiposTerreno = [
    { code: 'HT', label: 'Highway Terrain' },
    { code: 'AT', label: 'All Terrain' },
    { code: 'MT', label: 'Mud Terrain' },
    { code: 'RT', label: 'Rugged Terrain' },
    { code: 'AS', label: 'All Season' },
    { code: 'WS', label: 'Winter/Snow' },
  ];

  // Tipos de uso comercial específico
  readonly tiposUso = [
    { code: 'URB', label: 'Urbano' },
    { code: 'REG', label: 'Regional' },
    { code: 'LONG', label: 'Larga distancia' },
    { code: 'MIX', label: 'Mixto carretera/camino' },
  ];

  // Categorías de llanta para clasificar catálogo
  readonly categorias = [
    'Automóvil',
    'Camioneta',
    'SUV',
    'Camión ligero',
    'Camión pesado',
    'Agrícola',
    'Industrial',
    'OTR',
    'Motocicleta',
    'ATV / UTV',
  ];

  private listenersInitialized = false;

  private mapConstructionFromDto(value: string | null | undefined): string {
    const v = (value ?? '').toString().toUpperCase().trim();
    if (v === 'R') {
      return 'RADIAL';
    }
    if (v === 'D') {
      return 'DIAGONAL';
    }
    if (v === 'RADIAL' || v === 'DIAGONAL') {
      return v;
    }
    return v || 'RADIAL';
  }

  private mapConstructionToDto(value: unknown): string {
    const v = value === null || value === undefined ? '' : String(value).toUpperCase().trim();
    if (v === 'RADIAL') {
      return 'R';
    }
    if (v === 'DIAGONAL') {
      return 'D';
    }
    return v;
  }

  readonly pageTitle = computed(() =>
    this.isEditMode() ? `Editar llanta ${this.sku}` : 'Nueva llanta'
  );

  readonly submitLabel = computed(() =>
    this.isEditMode() ? 'Guardar cambios' : 'Crear llanta'
  );

  readonly form = this.fb.group({
    sku: ['', Validators.required],
    marcaId: [null as number | null, Validators.required],
    modelo: ['', Validators.required],
    ancho: [0, [Validators.required, Validators.min(1)]],
    perfil: [null as number | null],
    rin: [0, [Validators.required, Validators.min(1)]],
    construccion: ['RADIAL'],
    tipoTubo: ['TL'],
    calificacionCapas: [''],
    indiceCarga: [''],
    indiceVelocidad: [''],
    tipoNormalizado: [''],
    abreviaturaUso: [''],
    descripcion: [''],
    medidaOriginal: [''],
    precioPublico: [0, [Validators.min(0)]],
    urlImagen: [''],
    cantidad: [0, [Validators.min(0)]],
    stockMinimo: [0, [Validators.min(0)]],
  });

  constructor() {
    this.loadBrands();
    this.loadPriceColumnsDefinition();

    this.route.paramMap
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((params) => {
        const paramSku = params.get('sku');
        if (paramSku) {
          this.isEditMode.set(true);
          this.sku = paramSku.trim();
          this.loadTire();
        } else {
          this.isEditMode.set(false);
          this.sku = '';
          this.loading.set(false);
          this.error.set(null);
          this.priceCodes.set([]);
          this.prices.set({});
          this.loadPriceColumnsForNewTire();
          this.setupFormListeners();
        }
      });
  }

  private loadBrands(): void {
    this.brandsService
      .list()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (items) => this.brands.set(items),
        error: () => this.brands.set([]),
      });
  }

  private loadPriceColumnsDefinition(): void {
    this.priceColumnsService
      .list()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (items) => this.priceColumnsDef.set(items ?? []),
        error: () => this.priceColumnsDef.set([]),
      });
  }

  private loadPriceColumnsForNewTire(
    basePrices?: Record<string, number | null>
  ): void {
    this.catalogService
      .listAdmin({ limit: 50, offset: 0 })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (response: ApiListResponse<LlTireAdminDTO>) => {
          const codes = new Set<string>();
          for (const item of response.data ?? []) {
            const prices = item.prices || {};
            for (const key of Object.keys(prices)) {
              const trimmed = key && key.trim();
              if (trimmed) {
                codes.add(trimmed);
              }
            }
          }

          if (basePrices) {
            for (const key of Object.keys(basePrices)) {
              const trimmed = key && key.trim();
              if (trimmed) {
                codes.add(trimmed);
              }
            }
          }

          const sorted = Array.from(codes).sort();
          this.priceCodes.set(sorted);

          const initial: Record<string, number | null> = {};
          for (const code of sorted) {
            if (basePrices && Object.prototype.hasOwnProperty.call(basePrices, code)) {
              initial[code] = basePrices[code] ?? null;
            } else {
              initial[code] = null;
            }
          }
          this.prices.set(initial);
          this.syncPublicPriceFromPrices();
        },
        error: () => {
          this.priceCodes.set([]);
          this.prices.set({});
        },
      });
  }

  private loadTire(): void {
    const sku = this.sku.trim();
    if (!sku) {
      this.error.set('SKU inválido.');
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    this.catalogService
      .getAdminBySku(sku)
      .pipe(
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: (admin) => this.populateForm(admin),
        error: () => this.error.set('No se pudo cargar la llanta solicitada.'),
      });
  }

  private populateForm(admin: LlTireAdminDTO | null): void {
    if (!admin) {
      this.error.set('No se encontró la llanta con el SKU proporcionado.');
      return;
    }

    const t = admin.tire;
    const inv = admin.inventory;

    this.form.patchValue({
      sku: t.sku,
      marcaId: t.marcaId ?? null,
      modelo: t.modelo,
      ancho: t.ancho,
      perfil: t.perfil ?? null,
      rin: t.rin,
      construccion: this.mapConstructionFromDto(t.construccion),
      tipoTubo: t.tipoTubo || 'TL',
      calificacionCapas: t.calificacionCapas || '',
      indiceCarga: t.indiceCarga || '',
      indiceVelocidad: t.indiceVelocidad || '',
      tipoNormalizado: '',
      abreviaturaUso: t.abreviaturaUso || '',
      descripcion: t.descripcion || '',
      medidaOriginal: t.medidaOriginal || '',
      precioPublico: t.precioPublico ?? 0,
      urlImagen: t.urlImagen || '',
      cantidad: inv?.cantidad ?? 0,
      stockMinimo: inv?.stockMinimo ?? 0,
    });

    const prices = admin.prices || {};
    const mapped: Record<string, number | null> = {};
    for (const code of Object.keys(prices)) {
      const value = prices[code];
      mapped[code] = typeof value === 'number' ? value : null;
    }

    this.loadPriceColumnsForNewTire(mapped);
    this.setupFormListeners();
  }

  onSubmit(): void {
    this.error.set(null);
    this.success.set(null);
    this.form.markAllAsTouched();

    if (this.form.invalid) {
      this.error.set('Por favor revisa los campos marcados en rojo.');
      return;
    }

    const brandId = this.form.value.marcaId as number | null;
    const brand = this.brands().find((b) => b.id === brandId);
    if (!brand) {
      this.error.set('Selecciona una marca válida.');
      return;
    }

    const raw = this.form.value;
    const sku = (raw.sku ?? '').toString().trim();
    if (!sku) {
      this.error.set('El SKU es obligatorio.');
      return;
    }
    this.sku = sku;

    const payload: LlTireUpsertPayload = {
      sku,
      marcaNombre: brand.nombre,
      modelo: raw.modelo ?? '',
      ancho: Number(raw.ancho) || 0,
      perfil: raw.perfil !== null && raw.perfil !== undefined ? Number(raw.perfil) : null,
      rin: Number(raw.rin) || 0,
      construccion: this.mapConstructionToDto(raw.construccion),
      tipoTubo: (raw.tipoTubo ?? '').toString().toUpperCase(),
      calificacionCapas: raw.calificacionCapas ?? '',
      indiceCarga: raw.indiceCarga ?? '',
      indiceVelocidad: raw.indiceVelocidad ?? '',
      tipoNormalizado: raw.tipoNormalizado ?? '',
      abreviaturaUso: raw.abreviaturaUso ?? '',
      descripcion: raw.descripcion ?? '',
      precioPublico: raw.precioPublico ?? 0,
      urlImagen: raw.urlImagen ?? '',
      medidaOriginal: raw.medidaOriginal ?? '',
    };

    // Log de depuración para revisar qué se envía al backend
    console.log('Llanta upsert payload', payload);

    const cantidad = Number(raw.cantidad ?? 0);
    const stockMinimo = Number(raw.stockMinimo ?? 0);

    const preciosPayload: { [code: string]: number | null } = {};
    const currentPrices = this.prices();
    for (const code of this.priceCodes()) {
      const value = currentPrices[code];
      if (value === null || value === undefined) {
        preciosPayload[code] = null;
      } else {
        const parsed = Number(value);
        preciosPayload[code] = Number.isNaN(parsed) ? null : parsed;
      }
    }

    this.saving.set(true);

    this.catalogService
      .upsertTire(payload)
      .pipe(
        finalize(() => this.saving.set(false)),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe({
        next: () => {
          // Después de actualizar la llanta, actualizamos inventario y precios.
          this.saveAdminData(cantidad, stockMinimo, preciosPayload);
        },
        error: (err) => {
          console.error('Error al guardar llanta', err);
          this.error.set('No se pudo guardar la información de la llanta.');
        },
      });
  }

  private saveAdminData(
    cantidad: number,
    stockMinimo: number,
    precios: { [code: string]: number | null }
  ): void {
    // stockMinimo todavía no se utiliza en UpdateAdmin; se mantiene para futuras extensiones.
    const payload = {
      cantidad,
      precios,
    };

    this.catalogService
      .updateAdmin(this.sku, payload)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.success.set('Llanta guardada correctamente.');
          this.navigateToList();
        },
        error: (err) => {
          console.error('Error al guardar inventario/precios de llanta', err);
          this.error.set('La llanta se guardó parcialmente, pero ocurrió un error al actualizar inventario o precios.');
        },
      });
  }

  onCancel(): void {
    this.navigateToList();
  }

  isInvalid(controlName: string): boolean {
    const control = this.form.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  onPriceChange(code: string, value: string | number): void {
    const current = { ...this.prices() };
    if (value === null || value === undefined || value === '') {
      current[code] = null;
    } else {
      const parsed = typeof value === 'number' ? value : Number(value);
      current[code] = Number.isNaN(parsed) ? null : parsed;
    }
    this.prices.set(current);

    if (code && code.trim().toLowerCase() === 'lista') {
      this.syncPublicPriceFromPrices();
    }
  }

  isDerivedPriceCode(code: string): boolean {
    const normalized = (code ?? '').toString().trim().toLowerCase();
    if (!normalized) {
      return false;
    }
    const columns = this.priceColumnsDef();
    for (const col of columns) {
      const colCode = (col.code ?? '').toString().trim().toLowerCase();
      if (!colCode || colCode !== normalized) {
        continue;
      }
      if (col.calculation && col.calculation.mode === 'derived') {
        return true;
      }
    }
    return false;
  }

  private setupFormListeners(): void {
    if (this.listenersInitialized) {
      return;
    }
    this.listenersInitialized = true;

    const fieldsToWatch = [
      'ancho',
      'perfil',
      'rin',
      'construccion',
      'calificacionCapas',
      'abreviaturaUso',
      'indiceCarga',
      'indiceVelocidad',
      'modelo',
    ];

    for (const name of fieldsToWatch) {
      const control = this.form.get(name);
      if (!control) {
        continue;
      }
      control.valueChanges
        .pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe(() => {
          this.updateMedidaOriginalFromForm();
        });
    }
  }

  private buildMedidaOriginalFromForm(): string {
    const get = (name: string): unknown => this.form.get(name)?.value;

    const anchoRaw = get('ancho');
    const perfilRaw = get('perfil');
    const rinRaw = get('rin');
    const construccionRaw = get('construccion');
    const capasRaw = get('calificacionCapas');
    const usoRaw = get('abreviaturaUso');
    const indiceCargaRaw = get('indiceCarga');
    const indiceVelocidadRaw = get('indiceVelocidad');
    const modeloRaw = get('modelo');

    const ancho =
      anchoRaw !== null && anchoRaw !== undefined && anchoRaw !== ''
        ? Number(anchoRaw)
        : 0;
    const perfil =
      perfilRaw !== null && perfilRaw !== undefined && perfilRaw !== ''
        ? Number(perfilRaw)
        : null;
    const rin =
      rinRaw !== null && rinRaw !== undefined && rinRaw !== '' ? Number(rinRaw) : 0;

    const construccion = (construccionRaw ?? '').toString().toUpperCase().trim();
    const capas = (capasRaw ?? '').toString().trim();
    const uso = (usoRaw ?? '').toString().toUpperCase().trim();
    const indiceCarga = (indiceCargaRaw ?? '').toString().trim();
    const indiceVelocidad = (indiceVelocidadRaw ?? '').toString().toUpperCase().trim();
    const modelo = (modeloRaw ?? '').toString().trim();

    const isDiagonal =
      construccion === 'DIAGONAL' || construccion === 'D' || construccion === '-';
    const isRadial = construccion === 'RADIAL' || construccion === 'R';

    let medidaBase = '';

    if (ancho > 0) {
      const rinStr =
        rin > 0
          ? Number.isInteger(rin)
            ? rin.toFixed(0)
            : String(rin)
          : '';

      if (perfil !== null && perfil > 0) {
        const perfilStr = `/${perfil}`;
        const sep = isDiagonal ? '-' : 'R';
        const rinPart = rinStr ? `${sep}${rinStr}` : '';
        medidaBase = `${ancho}${perfilStr}${rinPart}`;
      } else if (isDiagonal) {
        const rinPart = rinStr ? `-${rinStr}` : '';
        medidaBase = `${ancho}${rinPart}`;
      } else {
        const rinPart = rinStr ? `X${rinStr}` : '';
        medidaBase = `${ancho}${rinPart}`;
      }
    }

    let usoCap = '';
    if (uso && capas) {
      usoCap = `${uso}-${capas}`;
    } else if (uso) {
      usoCap = uso;
    } else if (capas) {
      usoCap = capas;
    }

    let cargaVel = '';
    if (indiceCarga && indiceVelocidad) {
      cargaVel = `${indiceCarga}${indiceVelocidad}`;
    } else if (indiceCarga) {
      cargaVel = indiceCarga;
    } else if (indiceVelocidad) {
      cargaVel = indiceVelocidad;
    }

    const parts: string[] = [];
    if (medidaBase) {
      parts.push(medidaBase);
    }
    if (usoCap) {
      parts.push(usoCap);
    }
    if (cargaVel) {
      parts.push(cargaVel);
    }
    if (modelo) {
      parts.push(modelo);
    }

    return parts.join(' ').replace(/\s+/g, ' ').trim();
  }

  private updateMedidaOriginalFromForm(): void {
    const medida = this.buildMedidaOriginalFromForm();
    if (!medida) {
      return;
    }

    const current = (this.form.value.medidaOriginal ?? '').toString().trim();
    if (current === medida) {
      return;
    }

    this.form.patchValue(
      {
        medidaOriginal: medida,
      },
      { emitEvent: false }
    );
  }

  private syncPublicPriceFromPrices(): void {
    const codes = this.priceCodes();
    if (!codes || codes.length === 0) {
      return;
    }

    const listaCode = codes.find(
      (code) => code && code.trim().toLowerCase() === 'lista'
    );
    if (!listaCode) {
      return;
    }

    const currentPrices = this.prices();
    const rawValue = currentPrices[listaCode];

    if (rawValue === null || rawValue === undefined) {
      return;
    }

    const parsed = Number(rawValue);
    if (Number.isNaN(parsed)) {
      return;
    }

    const current = this.form.value.precioPublico ?? null;
    if (current !== null && Number(current) === parsed) {
      return;
    }

    this.form.patchValue(
      {
        precioPublico: parsed,
      },
      { emitEvent: false }
    );
  }

  private navigateToList(): void {
    this.router.navigate(['/dashboard/llantas']);
  }
}
