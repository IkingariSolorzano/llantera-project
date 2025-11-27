import { CommonModule } from '@angular/common';
import { Component, inject, signal, OnInit } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { LlAddressService, Address, CreateAddressRequest, UpdateAddressRequest } from '../services/ll-address.service';

@Component({
  selector: 'app-ll-cliente-direcciones',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './ll-cliente-direcciones.component.html',
})
export class LlClienteDireccionesComponent implements OnInit {
  private readonly addressService = inject(LlAddressService);

  readonly direcciones = signal<Address[]>([]);
  readonly showForm = signal(false);
  readonly editingId = signal<number | null>(null);
  readonly loading = signal(false);
  readonly loadingList = signal(false);
  readonly success = signal<string | null>(null);
  readonly error = signal<string | null>(null);

  ngOnInit(): void {
    this.loadAddresses();
  }

  private loadAddresses(): void {
    this.loadingList.set(true);
    this.addressService.list().subscribe({
      next: (addresses) => {
        this.direcciones.set(addresses);
        this.loadingList.set(false);
      },
      error: (err) => {
        console.error('Error cargando direcciones:', err);
        this.error.set('Error al cargar las direcciones');
        this.loadingList.set(false);
      }
    });
  }

  readonly estadosMexico = [
    'Aguascalientes', 'Baja California', 'Baja California Sur', 'Campeche',
    'Chiapas', 'Chihuahua', 'Ciudad de México', 'Coahuila', 'Colima',
    'Durango', 'Estado de México', 'Guanajuato', 'Guerrero', 'Hidalgo',
    'Jalisco', 'Michoacán', 'Morelos', 'Nayarit', 'Nuevo León', 'Oaxaca',
    'Puebla', 'Querétaro', 'Quintana Roo', 'San Luis Potosí', 'Sinaloa',
    'Sonora', 'Tabasco', 'Tamaulipas', 'Tlaxcala', 'Veracruz', 'Yucatán', 'Zacatecas'
  ];

  private readonly fb = new FormBuilder();

  readonly addressForm = this.fb.group({
    alias: ['', [Validators.required]],
    street: ['', [Validators.required]],
    exteriorNumber: ['', [Validators.required]],
    interiorNumber: [''],
    neighborhood: ['', [Validators.required]],
    postalCode: ['', [Validators.required, Validators.pattern(/^\d{5}$/)]],
    city: ['', [Validators.required]],
    state: ['', [Validators.required]],
    reference: [''],
    phone: ['', [Validators.required, Validators.pattern(/^\d{10}$/)]],
    isDefault: [false],
  });

  openNewForm(): void {
    this.editingId.set(null);
    this.addressForm.reset({ isDefault: false });
    this.showForm.set(true);
    this.clearMessages();
  }

  openEditForm(direccion: Address): void {
    this.editingId.set(direccion.id);
    this.addressForm.patchValue(direccion);
    this.showForm.set(true);
    this.clearMessages();
  }

  closeForm(): void {
    this.showForm.set(false);
    this.editingId.set(null);
    this.addressForm.reset();
    this.clearMessages();
  }

  clearMessages(): void {
    this.success.set(null);
    this.error.set(null);
  }

  onSubmit(): void {
    this.clearMessages();
    this.addressForm.markAllAsTouched();

    if (this.addressForm.invalid) {
      this.error.set('Por favor completa todos los campos requeridos.');
      return;
    }

    this.loading.set(true);
    const formValue = this.addressForm.value;

    if (this.editingId()) {
      // Actualizar dirección existente
      const request: UpdateAddressRequest = {
        alias: formValue.alias || undefined,
        street: formValue.street || undefined,
        exteriorNumber: formValue.exteriorNumber || undefined,
        interiorNumber: formValue.interiorNumber || undefined,
        neighborhood: formValue.neighborhood || undefined,
        postalCode: formValue.postalCode || undefined,
        city: formValue.city || undefined,
        state: formValue.state || undefined,
        reference: formValue.reference || undefined,
        phone: formValue.phone || undefined,
        isDefault: formValue.isDefault || undefined,
      };

      this.addressService.update(this.editingId()!, request).subscribe({
        next: (updated) => {
          this.direcciones.update(list =>
            list.map(d => d.id === updated.id ? updated : d)
          );
          this.success.set('Dirección actualizada correctamente.');
          this.loading.set(false);
          this.closeForm();
        },
        error: (err) => {
          console.error('Error actualizando dirección:', err);
          this.error.set('Error al actualizar la dirección');
          this.loading.set(false);
        }
      });
    } else {
      // Crear nueva dirección
      const request: CreateAddressRequest = {
        alias: formValue.alias || 'Nueva dirección',
        street: formValue.street || '',
        exteriorNumber: formValue.exteriorNumber || '',
        interiorNumber: formValue.interiorNumber || undefined,
        neighborhood: formValue.neighborhood || '',
        postalCode: formValue.postalCode || '',
        city: formValue.city || '',
        state: formValue.state || '',
        reference: formValue.reference || undefined,
        phone: formValue.phone || '',
        isDefault: formValue.isDefault || false,
      };

      this.addressService.create(request).subscribe({
        next: (created) => {
          this.direcciones.update(list => [...list, created]);
          this.success.set('Dirección agregada correctamente.');
          this.loading.set(false);
          this.closeForm();
        },
        error: (err) => {
          console.error('Error creando dirección:', err);
          this.error.set('Error al crear la dirección');
          this.loading.set(false);
        }
      });
    }
  }

  setAsDefault(id: number): void {
    this.addressService.setDefault(id).subscribe({
      next: () => {
        this.direcciones.update(list =>
          list.map(d => ({ ...d, isDefault: d.id === id }))
        );
        this.success.set('Dirección establecida como predeterminada.');
      },
      error: (err) => {
        console.error('Error estableciendo dirección predeterminada:', err);
        this.error.set('Error al establecer dirección predeterminada');
      }
    });
  }

  deleteDireccion(id: number): void {
    if (confirm('¿Estás seguro de eliminar esta dirección?')) {
      this.addressService.delete(id).subscribe({
        next: () => {
          this.direcciones.update(list => list.filter(d => d.id !== id));
          this.success.set('Dirección eliminada.');
        },
        error: (err) => {
          console.error('Error eliminando dirección:', err);
          this.error.set('Error al eliminar la dirección');
        }
      });
    }
  }

  isInvalid(controlName: string): boolean {
    const control = this.addressForm.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  formatDireccion(d: Address): string {
    let dir = `${d.street} #${d.exteriorNumber}`;
    if (d.interiorNumber) dir += ` Int. ${d.interiorNumber}`;
    dir += `, ${d.neighborhood}, C.P. ${d.postalCode}, ${d.city}, ${d.state}`;
    return dir;
  }
}
