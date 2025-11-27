import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { API_BASE_URL } from '../../core/config/api.config';

export interface LlEmpresaDTO {
  id: number;
  keyName: string;
  socialReason: string;
  rfc: string;
  address: string;
  emails: string[];
  phones: string[];
  mainContactId?: string | null;
  createdAt: string;
  updatedAt: string;
}

interface ApiListResponse<T> {
  data: T[];
  meta: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface LlEmpresaBasePayload {
  keyName: string;
  socialReason: string;
  rfc?: string;
  address?: string;
  emails: string[];
  phones: string[];
  mainContactId?: string | null;
}

@Injectable({ providedIn: 'root' })
export class LlEmpresasService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(search = '', limit = 20, offset = 0): Observable<LlEmpresaDTO[]> {
    let params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    const sanitizedSearch = search.trim();
    if (sanitizedSearch) {
      params = params.set('search', sanitizedSearch);
    }

    return this.http
      .get<ApiListResponse<LlEmpresaDTO>>(`${this.baseUrl}/companies/`, { params })
      .pipe(map((response) => response.data ?? []));
  }

  getById(id: number): Observable<LlEmpresaDTO> {
    return this.http.get<LlEmpresaDTO>(`${this.baseUrl}/companies/${id}`);
  }

  create(payload: LlEmpresaBasePayload): Observable<LlEmpresaDTO> {
    return this.http.post<LlEmpresaDTO>(`${this.baseUrl}/companies/`, payload);
  }

  update(id: number, payload: LlEmpresaBasePayload): Observable<LlEmpresaDTO> {
    return this.http.put<LlEmpresaDTO>(`${this.baseUrl}/companies/${id}`, payload);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/companies/${id}`);
  }
}
