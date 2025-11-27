import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { API_BASE_URL } from '../../core/config/api.config';

export interface LlTireDTO {
  id: string;
  sku: string;
  marcaId: number;
  modelo: string;
  ancho: number;
  perfil?: number | null;
  rin: number;
  construccion: string;
  tipoTubo: string;
  calificacionCapas: string;
  indiceCarga: string;
  indiceVelocidad: string;
  tipoNormalizadoId?: number | null;
  abreviaturaUso: string;
  descripcion: string;
  precioPublico: number;
  urlImagen: string;
  medidaOriginal: string;
  creadoEn: string;
  actualizadoEn: string;
}

export interface LlTireInventoryDTO {
  id: string;
  llantaId: string;
  cantidad: number;
  stockMinimo: number;
  creadoEn: string;
  actualizadoEn: string;
}

export interface LlTireAdminDTO {
  tire: LlTireDTO;
  inventory?: LlTireInventoryDTO | null;
  prices?: { [code: string]: number } | null;
}

export interface LlTireAdminUpdatePayload {
  cantidad?: number | null;
  precios?: { [code: string]: number | null };
}

// Payload para crear/actualizar una llanta completa desde el formulario admin.
// Se alinea con el tirePayload del backend (UpsertFromMeasurement).
export interface LlTireUpsertPayload {
  sku: string;
  marcaNombre: string;
  aliasMarca?: string;
  modelo: string;
  ancho: number;
  perfil?: number | null;
  rin: number;
  construccion: string;
  tipoTubo: string;
  calificacionCapas?: string;
  indiceCarga?: string;
  indiceVelocidad?: string;
  tipoNormalizado?: string;
  abreviaturaUso?: string;
  descripcion?: string;
  precioPublico?: number;
  urlImagen?: string;
  medidaOriginal?: string;
}

export interface LlTireCatalogItemDTO {
  tire: LlTireDTO;
  price: number;
  referencePrice?: number | null;
  priceCode: string;
  referenceCode?: string | null;
  stock?: number | null;
}

export interface ApiListResponse<T> {
  data: T[];
  meta: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface LlTireCatalogFilter {
  search?: string;
  limit?: number;
  offset?: number;
  level?: string;
  marcaId?: number;
  tipoId?: number;
  abreviatura?: string;
  ancho?: number;
  perfil?: number;
  rin?: number;
  construccion?: string;
  capas?: string;
  indiceCarga?: string;
  indiceVelocidad?: string;
  inStock?: boolean;
}

@Injectable({ providedIn: 'root' })
export class LlLlantasCatalogService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(filter: LlTireCatalogFilter = {}): Observable<LlTireCatalogItemDTO[]> {
    let params = new HttpParams();

    const limit = filter.limit ?? 24;
    const offset = filter.offset ?? 0;
    const level = filter.level ?? 'public';

    params = params.set('limit', String(limit));
    params = params.set('offset', String(offset));
    params = params.set('level', level);

    if (filter.search && filter.search.trim()) {
      params = params.set('search', filter.search.trim());
    }
    if (typeof filter.marcaId === 'number') {
      params = params.set('marcaId', String(filter.marcaId));
    }
    if (typeof filter.tipoId === 'number') {
      params = params.set('tipoId', String(filter.tipoId));
    }
    if (filter.abreviatura && filter.abreviatura.trim()) {
      params = params.set('abreviatura', filter.abreviatura.trim());
    }

    if (typeof filter.ancho === 'number') {
      params = params.set('ancho', String(filter.ancho));
    }
    if (typeof filter.perfil === 'number') {
      params = params.set('perfil', String(filter.perfil));
    }
    if (typeof filter.rin === 'number') {
      params = params.set('rin', String(filter.rin));
    }
    if (filter.construccion && filter.construccion.trim()) {
      params = params.set('construccion', filter.construccion.trim());
    }
    if (filter.capas && filter.capas.trim()) {
      params = params.set('capas', filter.capas.trim());
    }
    if (filter.indiceCarga && filter.indiceCarga.trim()) {
      params = params.set('indiceCarga', filter.indiceCarga.trim());
    }
    if (filter.indiceVelocidad && filter.indiceVelocidad.trim()) {
      params = params.set('indiceVelocidad', filter.indiceVelocidad.trim());
    }
    if (filter.inStock) {
      params = params.set('inStock', '1');
    }

    return this.http
      .get<ApiListResponse<LlTireCatalogItemDTO>>(`${this.baseUrl}/catalog/tires/`, { params })
      .pipe(map((response) => response.data ?? []));
  }

  listTires(filter: LlTireCatalogFilter = {}): Observable<LlTireDTO[]> {
    let params = new HttpParams();

    const limit = filter.limit ?? 50;
    const offset = filter.offset ?? 0;

    params = params.set('limit', String(limit));
    params = params.set('offset', String(offset));

    if (filter.search && filter.search.trim()) {
      params = params.set('search', filter.search.trim());
    }
    if (typeof filter.marcaId === 'number') {
      params = params.set('marcaId', String(filter.marcaId));
    }
    if (typeof filter.tipoId === 'number') {
      params = params.set('tipoId', String(filter.tipoId));
    }
    if (filter.abreviatura && filter.abreviatura.trim()) {
      params = params.set('abreviatura', filter.abreviatura.trim());
    }

    return this.http
      .get<ApiListResponse<LlTireDTO>>(`${this.baseUrl}/tires/`, { params })
      .pipe(map((response) => response.data ?? []));
  }

  listAdmin(filter: LlTireCatalogFilter = {}): Observable<ApiListResponse<LlTireAdminDTO>> {
    let params = new HttpParams();

    const limit = filter.limit ?? 50;
    const offset = filter.offset ?? 0;

    params = params.set('limit', String(limit));
    params = params.set('offset', String(offset));

    if (filter.search && filter.search.trim()) {
      params = params.set('search', filter.search.trim());
    }
    if (typeof filter.marcaId === 'number') {
      params = params.set('marcaId', String(filter.marcaId));
    }
    if (typeof filter.tipoId === 'number') {
      params = params.set('tipoId', String(filter.tipoId));
    }
    if (filter.abreviatura && filter.abreviatura.trim()) {
      params = params.set('abreviatura', filter.abreviatura.trim());
    }

    return this.http.get<ApiListResponse<LlTireAdminDTO>>(`${this.baseUrl}/tires/admin/`, { params });
  }

  updateAdmin(sku: string, payload: LlTireAdminUpdatePayload): Observable<LlTireAdminDTO> {
    return this.http.put<LlTireAdminDTO>(`${this.baseUrl}/tires/admin/${encodeURIComponent(sku)}`, payload);
  }

  getAdminBySku(sku: string): Observable<LlTireAdminDTO | null> {
    const trimmed = sku.trim();
    if (!trimmed) {
      return new Observable<LlTireAdminDTO | null>((subscriber) => {
        subscriber.next(null);
        subscriber.complete();
      });
    }

    const params = new HttpParams()
      .set('search', trimmed)
      .set('limit', '1')
      .set('offset', '0');

    return this.http
      .get<ApiListResponse<LlTireAdminDTO>>(`${this.baseUrl}/tires/admin/`, { params })
      .pipe(map((response) => (response.data && response.data.length ? response.data[0] : null)));
  }

  upsertTire(payload: LlTireUpsertPayload): Observable<LlTireDTO> {
    const sku = payload.sku.trim();
    return this.http.put<LlTireDTO>(`${this.baseUrl}/tires/${encodeURIComponent(sku)}`, payload);
  }

  exportAdminXlsx(filter: LlTireCatalogFilter = {}): Observable<Blob> {
    let params = new HttpParams();

    if (filter.search && filter.search.trim()) {
      params = params.set('search', filter.search.trim());
    }
    if (typeof filter.marcaId === 'number') {
      params = params.set('marcaId', String(filter.marcaId));
    }
    if (typeof filter.tipoId === 'number') {
      params = params.set('tipoId', String(filter.tipoId));
    }
    if (filter.abreviatura && filter.abreviatura.trim()) {
      params = params.set('abreviatura', filter.abreviatura.trim());
    }

    return this.http.get(`${this.baseUrl}/tires/admin/export`, {
      params,
      responseType: 'blob',
    });
  }

  importAdminXlsx(file: File): Observable<{ processed: number }> {
    const formData = new FormData();
    formData.append('file', file);

    return this.http.post<{ processed: number }>(`${this.baseUrl}/tires/admin/import`, formData);
  }
}
