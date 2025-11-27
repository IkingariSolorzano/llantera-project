import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DEFAULT_API_BASE_URL } from '../../../core/config/api.config';

const API_BASE_URL = DEFAULT_API_BASE_URL;

export interface Address {
  id: number;
  userId: number;
  alias: string;
  street: string;
  exteriorNumber: string;
  interiorNumber?: string;
  neighborhood: string;
  postalCode: string;
  city: string;
  state: string;
  reference?: string;
  phone: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateAddressRequest {
  alias: string;
  street: string;
  exteriorNumber: string;
  interiorNumber?: string;
  neighborhood: string;
  postalCode: string;
  city: string;
  state: string;
  reference?: string;
  phone: string;
  isDefault?: boolean;
}

export interface UpdateAddressRequest {
  alias?: string;
  street?: string;
  exteriorNumber?: string;
  interiorNumber?: string;
  neighborhood?: string;
  postalCode?: string;
  city?: string;
  state?: string;
  reference?: string;
  phone?: string;
  isDefault?: boolean;
}

@Injectable({
  providedIn: 'root'
})
export class LlAddressService {
  private readonly http = inject(HttpClient);
  // Nota: El backend espera trailing slash en /api/addresses/
  private readonly baseUrl = `${API_BASE_URL}/addresses/`;

  list(): Observable<Address[]> {
    return this.http.get<Address[]>(this.baseUrl);
  }

  getById(id: number): Observable<Address> {
    return this.http.get<Address>(`${this.baseUrl}${id}`);
  }

  create(request: CreateAddressRequest): Observable<Address> {
    return this.http.post<Address>(this.baseUrl, request);
  }

  update(id: number, request: UpdateAddressRequest): Observable<Address> {
    return this.http.patch<Address>(`${this.baseUrl}${id}`, request);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}${id}`);
  }

  setDefault(id: number): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}${id}/set-default`, {});
  }
}
