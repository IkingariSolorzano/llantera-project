import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { API_BASE_URL } from '../../core/config/api.config';
import { ApiListResponse } from './ll-llantas-catalog.service';

export type LlPriceCalculationOperation = 'add' | 'subtract' | 'multiply' | 'percent';

export type LlPriceColumnMode = 'fixed' | 'derived';

export interface LlPriceCalculationConfig {
  mode: LlPriceColumnMode;
  baseCode?: string | null;
  operation?: LlPriceCalculationOperation | null;
  amount?: number | null;
}

export interface LlPriceColumnDTO {
  id: number;
  code: string;
  label: string;
  description?: string;
  visualOrder?: number;
  active: boolean;
  isPublicPrice?: boolean;
  calculation?: LlPriceCalculationConfig | null;
  createdAt?: string;
  updatedAt?: string;
}

export interface LlPriceColumnUpsertPayload {
  code: string;
  label: string;
  description?: string;
  visualOrder?: number;
  active: boolean;
  isPublicPrice?: boolean;
  calculation?: LlPriceCalculationConfig | null;
}

interface LlPriceColumnApiResponse {
  id: number;
  code: string;
  name: string;
  description: string;
  visualOrder: number;
  active: boolean;
  isPublic: boolean;
  mode?: string;
  baseCode?: string | null;
  operation?: string;
  amount?: number | null;
}

@Injectable({ providedIn: 'root' })
export class LlLlantasPriceColumnsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(): Observable<LlPriceColumnDTO[]> {
    return this.http
      .get<ApiListResponse<LlPriceColumnApiResponse>>(`${this.baseUrl}/price-columns/`)
      .pipe(
        map((response) => {
          const items = response.data ?? [];
          return items.map((item) => this.mapFromApi(item));
        })
      );
  }

  getById(id: number): Observable<LlPriceColumnDTO> {
    return this.http
      .get<LlPriceColumnApiResponse>(`${this.baseUrl}/price-columns/${id}`)
      .pipe(map((item) => this.mapFromApi(item)));
  }

  create(payload: LlPriceColumnUpsertPayload): Observable<LlPriceColumnDTO> {
    const body = this.mapToApi(payload);
    return this.http
      .post<LlPriceColumnApiResponse>(`${this.baseUrl}/price-columns/`, body)
      .pipe(map((item) => this.mapFromApi(item)));
  }

  update(id: number, payload: LlPriceColumnUpsertPayload): Observable<LlPriceColumnDTO> {
    const body = this.mapToApi(payload);
    return this.http
      .put<LlPriceColumnApiResponse>(`${this.baseUrl}/price-columns/${id}`, body)
      .pipe(map((item) => this.mapFromApi(item)));
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/price-columns/${id}`);
  }

  private mapFromApi(item: LlPriceColumnApiResponse): LlPriceColumnDTO {
    const mode = (item.mode ?? '').toLowerCase();
    let calculation: LlPriceCalculationConfig | null = null;

    if (mode === 'derived') {
      calculation = {
        mode: 'derived',
        baseCode: item.baseCode ?? null,
        operation: (item.operation as LlPriceCalculationOperation | undefined) ?? 'percent',
        amount: typeof item.amount === 'number' ? item.amount : null,
      };
    }

    return {
      id: item.id,
      code: item.code,
      label: item.name,
      description: item.description,
      visualOrder: item.visualOrder,
      active: item.active,
      isPublicPrice: item.isPublic,
      calculation,
    };
  }

  private mapToApi(payload: LlPriceColumnUpsertPayload): {
    code?: string;
    name: string;
    description: string;
    visualOrder: number;
    active: boolean;
    isPublic: boolean;
    mode?: string;
    baseCode?: string;
    operation?: string;
    amount?: number | null;
  } {
    const calc = payload.calculation;
    const mode = calc?.mode ?? 'fixed';
    const baseCode = calc?.baseCode?.trim() || '';
    const operation = calc?.operation ?? '';
    const amount = typeof calc?.amount === 'number' ? calc.amount : null;

    return {
      code: payload.code?.trim() || undefined,
      name: payload.label?.trim() || '',
      description: payload.description?.trim() || '',
      visualOrder: typeof payload.visualOrder === 'number' ? payload.visualOrder : 0,
      active: payload.active,
      isPublic: payload.isPublicPrice ?? false,
      mode,
      baseCode,
      operation,
      amount,
    };
  }
}
