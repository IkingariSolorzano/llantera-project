import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { API_BASE_URL } from '../../core/config/api.config';

export interface LlUsuarioDTO {
  id: string;
  email: string;
  name: string;
  firstName: string;
  firstLastName: string;
  secondLastName: string;
  phone: string;
  addressStreet: string;
  addressNumber: string;
  addressNeighborhood: string;
  addressPostalCode: string;
  jobTitle: string;
  role: string;
  level: string;
  active: boolean;
  companyId?: number;
  profileImageUrl?: string;
  priceLevelId?: number;
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

export interface LlUsuarioBasePayload {
  email: string;
  firstName: string;
  firstLastName: string;
  secondLastName?: string;
  phone?: string;
  addressStreet?: string;
  addressNumber?: string;
  addressNeighborhood?: string;
  addressPostalCode?: string;
  jobTitle?: string;
  active: boolean;
  companyId?: number | null;
  profileImageUrl?: string;
  role: string;
  priceLevelId?: number | null;
  password?: string;
}

export interface LlUsuarioCreatePayload extends LlUsuarioBasePayload {
  password: string;
}

export interface LlUsuarioUpdatePayload extends LlUsuarioBasePayload {}

@Injectable({ providedIn: 'root' })
export class LlUsuariosService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(
    search = '',
    role?: string,
    limit = 20,
    offset = 0,
  ): Observable<LlUsuarioDTO[]> {
    let params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    const sanitizedSearch = search.trim();
    if (sanitizedSearch) {
      params = params.set('search', sanitizedSearch);
    }

    const sanitizedRole = role?.trim();
    if (sanitizedRole) {
      params = params.set('role', sanitizedRole);
    }

    return this.http
      .get<ApiListResponse<LlUsuarioDTO>>(`${this.baseUrl}/users/`, { params })
      .pipe(map((response) => response.data ?? []));
  }

  getById(id: string): Observable<LlUsuarioDTO> {
    return this.http.get<LlUsuarioDTO>(`${this.baseUrl}/users/${id}`);
  }

  create(payload: LlUsuarioCreatePayload): Observable<LlUsuarioDTO> {
    return this.http.post<LlUsuarioDTO>(`${this.baseUrl}/users/`, payload);
  }

  update(id: string, payload: LlUsuarioUpdatePayload): Observable<LlUsuarioDTO> {
    return this.http.put<LlUsuarioDTO>(`${this.baseUrl}/users/${id}`, payload);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/users/${id}`);
  }
}
