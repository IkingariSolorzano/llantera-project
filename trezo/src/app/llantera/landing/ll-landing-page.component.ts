import { CommonModule, NgClass } from '@angular/common';
import { Component, HostListener, inject, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { LlSolicitudesClientesService } from '../solicitudes/ll-solicitudes-clientes.service';
import {
  LlLlantasCatalogService,
  LlTireCatalogItemDTO,
  LlTireCatalogFilter,
} from '../llantas/ll-llantas-catalog.service';
import { LlAuthService } from '../auth/ll-auth.service';
import { ToggleService } from '../../common/header/toggle.service';

interface LandingSlide {
  id: number;
  title: string;
  subtitle: string;
}

interface TopTire {
  id: number;
  name: string;
  description: string;
  image: string;
}

@Component({
  selector: 'app-ll-landing-page',
  standalone: true,
  imports: [CommonModule, NgClass, RouterLink],
  templateUrl: './ll-landing-page.component.html',
  styleUrl: './ll-landing-page.component.scss',
})
export class LlLandingPageComponent {
  private readonly customerRequestsService = inject(LlSolicitudesClientesService);
  private readonly catalogService = inject(LlLlantasCatalogService);
  readonly auth = inject(LlAuthService);
  private readonly toggleService = inject(ToggleService);
  private readonly sectionIds = [
    'hero',
    'servicios',
    'quienes-somos',
    'sucursales',
    'productos-destacados',
  ];

  readonly slides: LandingSlide[] = [
    {
      id: 0,
      title: 'Expertos en llantas y servicios automotrices en Morelia.',
      subtitle:
        'Más de 20 años de experiencia cuidando tu seguridad en el camino.',
    },
    {
      id: 1,
      title: 'Tu camino, nuestra seguridad.',
      subtitle:
        'Calidad y servicio que ruedan contigo en cada kilómetro.',
    },
    {
      id: 2,
      title: 'Todo para tu vehículo en un solo lugar.',
      subtitle:
        'Llantas, alineación, balanceo y mantenimiento preventivo para tu auto, SUV o camioneta.',
    },
  ];

  readonly currentSlideIndex = signal(0);
  readonly isCotizadorOpen = signal(false);
  readonly isClienteModalOpen = signal(false);
  readonly isSubmittingClienteRequest = signal(false);
  readonly showClienteSuccessToast = signal(false);
  readonly currentYear = new Date().getFullYear();
  readonly activeSection = signal<string>('hero');
  readonly selectedBranchId = signal<'periodismo' | 'camelinas'>('periodismo');
  readonly topTiresCenterIndex = signal(1);

  readonly cotizadorResults = signal<LlTireCatalogItemDTO[]>([]);
  readonly cotizadorLoading = signal(false);
  readonly cotizadorError = signal<string | null>(null);
  readonly cotizadorCantidad = signal(1);

  readonly cotizadorAnchos = signal<number[]>([]);
  readonly cotizadorPerfiles = signal<number[]>([]);
  readonly cotizadorRines = signal<number[]>([]);
  readonly cotizadorConstrucciones = signal<string[]>([]);
  readonly cotizadorUsos = signal<string[]>([]);
  readonly cotizadorIndicesCarga = signal<string[]>([]);
  readonly cotizadorAllAnchos = signal<number[]>([]);

  readonly cotizadorOnlyInStock = signal(false);

  readonly cotizadorSelectedAncho = signal<number | null>(null);
  readonly cotizadorSelectedPerfil = signal<number | null>(null);
  readonly cotizadorSelectedRin = signal<number | null>(null);
  readonly cotizadorSelectedConstruccion = signal<string | null>(null);
  readonly cotizadorSelectedUso = signal<string | null>(null);
  readonly cotizadorSelectedIndiceCarga = signal<string | null>(null);

  readonly topTires: TopTire[] = [
    {
      id: 0,
      name: 'Bridgestone Turanza ER300',
      description:
        'Llanta para automóvil con excelente rendimiento en mojado, gran durabilidad e ideal para uso diario y carretera.',
      image: '/images/llantera/llantas/llanta1.jpg',
    },
    {
      id: 1,
      name: 'Michelin Defender',
      description:
        'Ideal para SUVs y camionetas, reconocida por su alta resistencia, confort y contribución al ahorro de combustible.',
      image: '/images/llantera/llantas/llanta2.jpg',
    },
    {
      id: 2,
      name: 'Pirelli Cinturato P7',
      description:
        'Llanta de alto desempeño con agarre superior en carretera, bajo consumo de combustible y gran confort de manejo.',
      image: '/images/llantera/llantas/llanta3.jpg',
    },
    {
      id: 3,
      name: 'Firestone Destination LE2',
      description:
        'Diseñada para camionetas y SUVs, ofrece tracción confiable en ciudad y carretera con excelente relación valor-precio.',
      image: '/images/llantera/llantas/llanta4.jpg',
    },
    {
      id: 4,
      name: 'Goodyear Wrangler All-Terrain',
      description:
        'Llantas todo terreno para quienes combinan ciudad y caminos difíciles, con buena tracción y resistencia.',
      image: '/images/llantera/llantas/llanta2.jpg',
    },
    {
      id: 5,
      name: 'Bridgestone Dueler H/T',
      description:
        'Llanta para camioneta orientada a autopista, con enfoque en confort, estabilidad y larga vida útil.',
      image: '/images/llantera/llantas/llanta1.jpg',
    },
    {
      id: 6,
      name: 'Kit básico de emergencias',
      description:
        'Kit de parches, sellador y herramientas básicas para resolver pinchaduras y emergencias ligeras en el camino.',
        image: '/images/llantera/llantas/llanta4.jpg'
    },
  ];

  private clienteToastTimeoutId: number | null = null;

  constructor() {
    this.toggleService.initializeTheme();
  }

  nextSlide(): void {
    const next = (this.currentSlideIndex() + 1) % this.slides.length;
    this.currentSlideIndex.set(next);
  }

  prevSlide(): void {
    const prev =
      (this.currentSlideIndex() - 1 + this.slides.length) % this.slides.length;
    this.currentSlideIndex.set(prev);
  }

  goToSlide(index: number): void {
    if (index < 0 || index >= this.slides.length) {
      return;
    }
    this.currentSlideIndex.set(index);
  }

  nextTopTire(): void {
    const length = this.topTires.length;
    if (!length) {
      return;
    }
    const next = (this.topTiresCenterIndex() + 1) % length;
    this.topTiresCenterIndex.set(next);
  }

  toggleTheme(): void {
    this.toggleService.toggleTheme();
  }

  prevTopTire(): void {
    const length = this.topTires.length;
    if (!length) {
      return;
    }
    const prev = (this.topTiresCenterIndex() - 1 + length) % length;
    this.topTiresCenterIndex.set(prev);
  }

  isTopTireCenter(index: number): boolean {
    return index === this.topTiresCenterIndex();
  }

  isTopTireSide(index: number): boolean {
    const length = this.topTires.length;
    if (!length) {
      return false;
    }
    const center = this.topTiresCenterIndex();
    const left = (center - 1 + length) % length;
    const right = (center + 1) % length;
    return index === left || index === right;
  }

  isTopTireVisible(index: number): boolean {
    return this.isTopTireCenter(index) || this.isTopTireSide(index);
  }

  getVisibleTopTires(): { tire: TopTire; position: 'left' | 'center' | 'right' }[] {
    const length = this.topTires.length;
    if (!length) {
      return [];
    }

    const center = this.topTiresCenterIndex();
    const left = (center - 1 + length) % length;
    const right = (center + 1) % length;

    return [
      { tire: this.topTires[left], position: 'left' },
      { tire: this.topTires[center], position: 'center' },
      { tire: this.topTires[right], position: 'right' },
    ];
  }

  openClienteModal(): void {
    this.isClienteModalOpen.set(true);
  }

  closeClienteModal(): void {
    this.isClienteModalOpen.set(false);
  }

  onClienteFormSubmit(event: Event): void {
    event.preventDefault();
    const form = event.target as HTMLFormElement | null;
    if (!form) {
      return;
    }

    const phoneInput = form.querySelector<HTMLInputElement>('input[name="phone"]');
    const emailInput = form.querySelector<HTMLInputElement>('input[name="email"]');

    if (phoneInput) {
      phoneInput.setCustomValidity('');
    }
    if (emailInput) {
      emailInput.setCustomValidity('');
    }

    const phone = (phoneInput?.value ?? '').trim();
    const email = (emailInput?.value ?? '').trim();

    if (!phone && !email) {
      const message =
        'Ingresa al menos un medio de contacto: teléfono o correo electrónico.';
      if (phoneInput) {
        phoneInput.setCustomValidity(message);
      }
      if (emailInput) {
        emailInput.setCustomValidity(message);
      }
    }

    if (!form.checkValidity()) {
      form.reportValidity();
      return;
    }

    const formData = new FormData(form);
    const fullName = (formData.get('fullName') ?? '').toString().trim();
    const requestType = (formData.get('requestType') ?? '').toString().trim();
    const message = (formData.get('message') ?? '').toString().trim();
    const contactPreference = (
      formData.get('contactPreference') ?? 'whatsapp'
    )
      .toString()
      .trim();

    this.isSubmittingClienteRequest.set(true);

    this.customerRequestsService
      .create({
        fullName,
        requestType,
        message,
        phone,
        contactPreference,
        email,
      })
      .subscribe({
        next: () => {
          this.isSubmittingClienteRequest.set(false);
          form.reset();
          this.closeClienteModal();

          this.showClienteSuccessToast.set(true);
          if (this.clienteToastTimeoutId !== null) {
            clearTimeout(this.clienteToastTimeoutId);
          }
          this.clienteToastTimeoutId = window.setTimeout(() => {
            this.showClienteSuccessToast.set(false);
            this.clienteToastTimeoutId = null;
          }, 4000);
        },
        error: (err) => {
          console.error('Error al enviar la solicitud de cliente', err);
          this.isSubmittingClienteRequest.set(false);
        },
      });
  }

  onPhoneInput(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input) {
      return;
    }

    const digitsOnly = input.value.replace(/\D/g, '').slice(0, 10);
    input.value = digitsOnly;
  }

  openCotizador(): void {
    this.isCotizadorOpen.set(true);
    this.cotizadorError.set(null);

    this.cotizadorSelectedAncho.set(null);
    this.cotizadorSelectedPerfil.set(null);
    this.cotizadorSelectedRin.set(null);
    this.cotizadorSelectedConstruccion.set(null);
    this.cotizadorSelectedUso.set(null);
    this.cotizadorSelectedIndiceCarga.set(null);
    this.cotizadorOnlyInStock.set(false);

    this.loadCotizadorFromSelections();
  }

  onCotizadorInStockToggle(checked: boolean): void {
    this.cotizadorOnlyInStock.set(checked);
    this.loadCotizadorFromSelections();
  }

  closeCotizador(): void {
    this.isCotizadorOpen.set(false);
  }

  onCotizadorSubmit(event: Event): void {
    event.preventDefault();
    this.loadCotizadorFromSelections();
  }

  onCotizadorSelectChange(
    field: 'ancho' | 'perfil' | 'rin' | 'construccion' | 'uso' | 'indiceCarga',
    rawValue: string,
  ): void {
    const value = rawValue.trim();

    switch (field) {
      case 'ancho': {
        if (!value) {
          this.cotizadorSelectedAncho.set(null);
          break;
        }
        const n = Number(value);
        this.cotizadorSelectedAncho.set(Number.isNaN(n) ? null : n);
        break;
      }
      case 'perfil': {
        if (!value) {
          this.cotizadorSelectedPerfil.set(null);
          break;
        }
        const n = Number(value);
        this.cotizadorSelectedPerfil.set(Number.isNaN(n) ? null : n);
        break;
      }
      case 'rin': {
        if (!value) {
          this.cotizadorSelectedRin.set(null);
          break;
        }
        const n = Number(value);
        this.cotizadorSelectedRin.set(Number.isNaN(n) ? null : n);
        break;
      }
      case 'construccion': {
        this.cotizadorSelectedConstruccion.set(value || null);
        break;
      }
      case 'uso': {
        this.cotizadorSelectedUso.set(value || null);
        break;
      }
      case 'indiceCarga': {
        this.cotizadorSelectedIndiceCarga.set(value || null);
        break;
      }
    }

    this.loadCotizadorFromSelections();
  }

  onCotizadorCantidadInput(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input) {
      return;
    }

    let cantidad = Number(input.value || '1');
    if (!Number.isFinite(cantidad) || cantidad <= 0) {
      cantidad = 1;
    }
    this.cotizadorCantidad.set(Math.floor(cantidad));
  }

  private loadCotizadorFromSelections(): void {
    const filter: LlTireCatalogFilter = {
      limit: 10000,
      offset: 0,
      level: 'public',
    };

    const ancho = this.cotizadorSelectedAncho();
    const perfil = this.cotizadorSelectedPerfil();
    const rin = this.cotizadorSelectedRin();
    const construccion = this.cotizadorSelectedConstruccion();
    const uso = this.cotizadorSelectedUso();
    const indiceCarga = this.cotizadorSelectedIndiceCarga();
    const onlyInStock = this.cotizadorOnlyInStock();

    if (ancho !== null) {
      filter.ancho = ancho;
    }
    if (perfil !== null) {
      filter.perfil = perfil;
    }
    if (rin !== null) {
      filter.rin = rin;
    }
    if (construccion) {
      filter.construccion = construccion;
    }
    if (uso) {
      filter.abreviatura = uso;
    }
    if (indiceCarga) {
      filter.indiceCarga = indiceCarga;
    }
    if (onlyInStock) {
      filter.inStock = true;
    }

    this.cotizadorLoading.set(true);
    this.cotizadorError.set(null);

    this.catalogService.list(filter).subscribe({
      next: (items) => {
        const list = items ?? [];
        this.cotizadorResults.set(list);
        this.buildCotizadorOptions(list);
        this.normalizeCotizadorSelections();
        this.cotizadorLoading.set(false);
      },
      error: (err) => {
        console.error('Error al cargar resultados del cotizador', err);
        this.cotizadorError.set('No se pudieron cargar llantas para esos filtros.');
        this.cotizadorLoading.set(false);
      },
    });
  }

  private buildCotizadorOptions(items: LlTireCatalogItemDTO[]): void {
    const anchos = new Set<number>();
    const perfiles = new Set<number>();
    const rines = new Set<number>();
    const construcciones = new Set<string>();
    const usos = new Set<string>();
    const indicesCarga = new Set<string>();

    for (const item of items) {
      const t = item.tire;
      if (typeof t.ancho === 'number' && !Number.isNaN(t.ancho)) {
        anchos.add(t.ancho);
      }
      if (typeof t.perfil === 'number' && !Number.isNaN(t.perfil)) {
        perfiles.add(t.perfil);
      }
      if (typeof t.rin === 'number' && !Number.isNaN(t.rin)) {
        rines.add(t.rin);
      }
      const construccion = (t.construccion ?? '').trim();
      if (construccion) {
        construcciones.add(construccion);
      }
      const uso = (t.abreviaturaUso ?? '').trim();
      if (uso) {
        usos.add(uso);
      }
      const indiceCarga = (t.indiceCarga ?? '').trim();
      if (indiceCarga) {
        indicesCarga.add(indiceCarga);
      }
    }

    const anchosArray = Array.from(anchos).sort((a, b) => a - b);
    const perfilesArray = Array.from(perfiles).sort((a, b) => a - b);
    const rinesArray = Array.from(rines).sort((a, b) => a - b);
    const construccionesArray = Array.from(construcciones).sort();
    const usosArray = Array.from(usos).sort();
    const indicesCargaArray = Array.from(indicesCarga).sort();

    const selectedAncho = this.cotizadorSelectedAncho();
    const hasOtherFilters =
      this.cotizadorSelectedPerfil() !== null ||
      this.cotizadorSelectedRin() !== null ||
      !!this.cotizadorSelectedConstruccion() ||
      !!this.cotizadorSelectedUso() ||
      !!this.cotizadorSelectedIndiceCarga();

    // Si no hay ningún filtro activo, guardamos el catálogo base de anchos.
    if (selectedAncho === null && !hasOtherFilters) {
      this.cotizadorAllAnchos.set(anchosArray);
    }

    // Si solo hay ancho seleccionado (ningún otro filtro), mostramos
    // siempre el catálogo completo de anchos para que el select no se
    // reduzca a una sola opción.
    if (selectedAncho !== null && !hasOtherFilters) {
      const baseAnchos = this.cotizadorAllAnchos();
      this.cotizadorAnchos.set(baseAnchos.length ? baseAnchos : anchosArray);
    } else {
      this.cotizadorAnchos.set(anchosArray);
    }

    this.cotizadorPerfiles.set(perfilesArray);
    this.cotizadorRines.set(rinesArray);
    this.cotizadorConstrucciones.set(construccionesArray);
    this.cotizadorUsos.set(usosArray);
    this.cotizadorIndicesCarga.set(indicesCargaArray);
  }

  private normalizeCotizadorSelections(): void {
    const anchos = new Set(this.cotizadorAnchos());
    const perfiles = new Set(this.cotizadorPerfiles());
    const rines = new Set(this.cotizadorRines());
    const construcciones = new Set(this.cotizadorConstrucciones());
    const usos = new Set(this.cotizadorUsos());
    const indicesCarga = new Set(this.cotizadorIndicesCarga());

    const anchoSel = this.cotizadorSelectedAncho();
    if (anchoSel !== null && !anchos.has(anchoSel)) {
      this.cotizadorSelectedAncho.set(null);
    }

    const perfilSel = this.cotizadorSelectedPerfil();
    if (perfilSel !== null && !perfiles.has(perfilSel)) {
      this.cotizadorSelectedPerfil.set(null);
    }

    const rinSel = this.cotizadorSelectedRin();
    if (rinSel !== null && !rines.has(rinSel)) {
      this.cotizadorSelectedRin.set(null);
    }

    const construccionSel = this.cotizadorSelectedConstruccion();
    if (construccionSel && !construcciones.has(construccionSel)) {
      this.cotizadorSelectedConstruccion.set(null);
    }

    const usoSel = this.cotizadorSelectedUso();
    if (usoSel && !usos.has(usoSel)) {
      this.cotizadorSelectedUso.set(null);
    }

    const indiceCargaSel = this.cotizadorSelectedIndiceCarga();
    if (indiceCargaSel && !indicesCarga.has(indiceCargaSel)) {
      this.cotizadorSelectedIndiceCarga.set(null);
    }
  }

  trackByIndex(index: number): number {
    return index;
  }

  selectBranch(id: 'periodismo' | 'camelinas'): void {
    this.selectedBranchId.set(id);
  }

  onMenuClick(event: Event, sectionId: string): void {
    event.preventDefault();
    this.scrollToSection(sectionId);
  }

  private scrollToSection(sectionId: string): void {
    const element = document.getElementById(sectionId);
    if (!element) {
      return;
    }

    const header = document.querySelector('header') as HTMLElement | null;
    const baseHeaderOffset = header?.offsetHeight ?? 40; // altura del header sticky
    const headerOffset = baseHeaderOffset + 75; // pequeño margen extra para que no se tape el título
    const rect = element.getBoundingClientRect();
    const scrollTop = window.scrollY || window.pageYOffset;
    const targetY = rect.top + scrollTop - headerOffset;

    window.scrollTo({ top: targetY, behavior: 'smooth' });
  }

  @HostListener('window:scroll')
  onWindowScroll(): void {
    const viewportHeight =
      window.innerHeight || document.documentElement.clientHeight || 0;
    const probeOffset = viewportHeight * 0.45; // ~35% de la altura de la pantalla
    const scrollPos = (window.scrollY || window.pageYOffset) + probeOffset;

    let current: string | null = null;

    for (const id of this.sectionIds) {
      const el = document.getElementById(id);
      if (!el) {
        continue;
      }

      const offsetTop = el.offsetTop;
      const offsetHeight = el.offsetHeight;

      if (scrollPos >= offsetTop && scrollPos < offsetTop + offsetHeight) {
        current = id;
        break;
      }
    }

    if (current && this.activeSection() !== current) {
      this.activeSection.set(current);
    }
  }
}
