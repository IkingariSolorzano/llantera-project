import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { API_BASE_URL } from '../../core/config/api.config';

export interface LlCustomerRequestDTO {
  id: string;
  fullName: string;
  requestType: string;
  message: string;
  phone: string;
  contactPreference: string;
  email: string;
  status: string;
  employeeId?: string | null;
  agreement: string;
  createdAt: string;
  updatedAt: string;
  attendedAt?: string | null;
}

interface ApiListResponse<T> {
  data: T[];
  meta: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface LlCustomerRequestCreatePayload {
  fullName: string;
  requestType: string;
  message: string;
  phone: string;
  contactPreference: string;
  email: string;
}

export interface LlCustomerRequestUpdatePayload {
  message?: string;
  status?: string;
  employeeId?: string | null;
  agreement?: string;
}

@Injectable({ providedIn: 'root' })
export class LlSolicitudesClientesService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = inject(API_BASE_URL);

  list(
    search = '',
    status?: string,
    employeeId?: string,
    limit = 20,
    offset = 0
  ): Observable<LlCustomerRequestDTO[]> {
    let params = new HttpParams()
      .set('limit', limit.toString())
      .set('offset', offset.toString());

    const sanitizedSearch = search.trim();
    if (sanitizedSearch) {
      params = params.set('search', sanitizedSearch);
    }

    const sanitizedStatus = status?.trim();
    if (sanitizedStatus) {
      params = params.set('status', sanitizedStatus);
    }

    const sanitizedEmployeeId = employeeId?.trim();
    if (sanitizedEmployeeId) {
      params = params.set('employeeId', sanitizedEmployeeId);
    }

    return this.http
      .get<ApiListResponse<LlCustomerRequestDTO>>(
        `${this.baseUrl}/customer-requests/`,
        { params }
      )
      .pipe(map((response) => response.data ?? []));
  }

  getById(id: string): Observable<LlCustomerRequestDTO> {
    return this.http.get<LlCustomerRequestDTO>(
      `${this.baseUrl}/customer-requests/${id}`
    );
  }

  create(
    payload: LlCustomerRequestCreatePayload
  ): Observable<LlCustomerRequestDTO> {
    return this.http.post<LlCustomerRequestDTO>(
      `${this.baseUrl}/customer-requests/`,
      payload
    );
  }

  update(
    id: string,
    payload: LlCustomerRequestUpdatePayload
  ): Observable<LlCustomerRequestDTO> {
    return this.http.put<LlCustomerRequestDTO>(
      `${this.baseUrl}/customer-requests/${id}`,
      payload
    );
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(
      `${this.baseUrl}/customer-requests/${id}`
    );
  }
}
