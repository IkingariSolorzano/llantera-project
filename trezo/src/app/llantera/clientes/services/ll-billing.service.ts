import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DEFAULT_API_BASE_URL } from '../../../core/config/api.config';

const API_BASE_URL = DEFAULT_API_BASE_URL;

export interface BillingInfo {
  id: number;
  userId: string;
  rfc: string;
  razonSocial: string;
  regimenFiscal: string;
  usoCfdi: string;
  postalCode: string;
  email?: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateBillingRequest {
  rfc: string;
  razonSocial: string;
  regimenFiscal: string;
  usoCfdi: string;
  postalCode: string;
  email?: string;
  isDefault?: boolean;
}

export interface UpdateBillingRequest {
  rfc?: string;
  razonSocial?: string;
  regimenFiscal?: string;
  usoCfdi?: string;
  postalCode?: string;
  email?: string;
  isDefault?: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class LlBillingService {
  private readonly http = inject(HttpClient);
  // Nota: El backend espera trailing slash
  private readonly baseUrl = `${API_BASE_URL}/billing/`;

  list(): Observable<BillingInfo[]> {
    return this.http.get<BillingInfo[]>(this.baseUrl);
  }

  getById(id: number): Observable<BillingInfo> {
    return this.http.get<BillingInfo>(`${this.baseUrl}${id}`);
  }

  getDefault(): Observable<BillingInfo | null> {
    return this.http.get<BillingInfo | null>(`${this.baseUrl}default`);
  }

  create(request: CreateBillingRequest): Observable<BillingInfo> {
    return this.http.post<BillingInfo>(this.baseUrl, request);
  }

  update(id: number, request: UpdateBillingRequest): Observable<BillingInfo> {
    return this.http.patch<BillingInfo>(`${this.baseUrl}${id}`, request);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}${id}`);
  }

  setDefault(id: number): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}${id}/set-default`, {});
  }
}
