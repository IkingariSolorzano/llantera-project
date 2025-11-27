import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { API_BASE_URL } from '../../core/config/api.config';

export interface LlPriceLevelDTO {
  id: number;
  code: string;
  name: string;
  description?: string | null;
  discountPercentage: number;
  priceColumn: string;
  referenceColumn?: string | null;
  canViewOffers: boolean;
}

export interface CreatePriceLevelRequest {
  code: string;
  name: string;
  description?: string | null;
  discountPercentage: number;
  priceColumn: string;
  referenceColumn?: string | null;
  canViewOffers: boolean;
}

export interface UpdatePriceLevelRequest {
  code?: string;
  name?: string;
  description?: string | null;
  discountPercentage?: number;
  priceColumn?: string;
  referenceColumn?: string | null;
  canViewOffers?: boolean;
}

interface ApiListResponse<T> {
  data: T[];
  meta: {
    total: number;
    limit: number;
    offset: number;
  };
}

@Injectable({ providedIn: 'root' })
export class LlPriceLevelsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(limit = 100, offset = 0, code?: string): Observable<{ items: LlPriceLevelDTO[]; total: number }> {
    let params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    const trimmedCode = code?.trim();
    if (trimmedCode) {
      params = params.set('code', trimmedCode);
    }

    return this.http
      .get<ApiListResponse<LlPriceLevelDTO>>(`${this.baseUrl}/price-levels/`, { params })
      .pipe(map((response) => ({
        items: response.data ?? [],
        total: response.meta?.total ?? 0
      })));
  }

  getById(id: number): Observable<LlPriceLevelDTO> {
    return this.http.get<LlPriceLevelDTO>(`${this.baseUrl}/price-levels/${id}`);
  }

  create(data: CreatePriceLevelRequest): Observable<LlPriceLevelDTO> {
    return this.http.post<LlPriceLevelDTO>(`${this.baseUrl}/price-levels/`, data);
  }

  update(id: number, data: UpdatePriceLevelRequest): Observable<LlPriceLevelDTO> {
    return this.http.put<LlPriceLevelDTO>(`${this.baseUrl}/price-levels/${id}`, data);
  }

  delete(id: number, transferToId?: number): Observable<void> {
    let params = new HttpParams();
    if (transferToId) {
      params = params.set('transferToId', transferToId.toString());
    }
    return this.http.delete<void>(`${this.baseUrl}/price-levels/${id}`, { params });
  }
}
