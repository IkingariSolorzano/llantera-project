import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { API_BASE_URL } from '../../core/config/api.config';

export interface LlTireBrandDTO {
  id: number;
  nombre: string;
  creadoEn: string;
  actualizadoEn: string;
  aliases?: string[];
}

interface ApiListResponse<T> {
  data: T[];
  meta: {
    total: number;
  };
}

@Injectable({ providedIn: 'root' })
export class LlLlantasBrandsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(): Observable<LlTireBrandDTO[]> {
    return this.http
      .get<ApiListResponse<LlTireBrandDTO>>(`${this.baseUrl}/brands/`)
      .pipe(map((response) => response.data ?? []));
  }

  create(payload: { nombre: string; aliases: string[] }): Observable<LlTireBrandDTO> {
    return this.http.post<LlTireBrandDTO>(`${this.baseUrl}/brands/`, payload);
  }

  update(id: number, payload: { nombre: string; aliases: string[] }): Observable<LlTireBrandDTO> {
    return this.http.put<LlTireBrandDTO>(`${this.baseUrl}/brands/${id}`, payload);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/brands/${id}`);
  }
}
