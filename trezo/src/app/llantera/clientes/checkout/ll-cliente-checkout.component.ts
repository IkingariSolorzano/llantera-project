import { CommonModule, isPlatformBrowser } from '@angular/common';
import { Component, inject, signal, computed, OnInit, PLATFORM_ID } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { LlCartService } from '../cart/ll-cart.service';
import { LlAuthService } from '../../auth/ll-auth.service';
import { LlAddressService, Address, CreateAddressRequest } from '../services/ll-address.service';
import { LlOrderService, CreateOrderRequest, PaymentMethod, PaymentMode } from '../services/ll-order.service';
import { LlBillingService, BillingInfo, CreateBillingRequest } from '../services/ll-billing.service';

export type CheckoutStep = 'direccion' | 'facturacion' | 'pago' | 'confirmacion';

@Component({
  selector: 'app-ll-cliente-checkout',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './ll-cliente-checkout.component.html',
})
export class LlClienteCheckoutComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly router = inject(Router);
  private readonly platformId = inject(PLATFORM_ID);
  readonly cart = inject(LlCartService);
  private readonly auth = inject(LlAuthService);
  private readonly addressService = inject(LlAddressService);
  private readonly orderService = inject(LlOrderService);
  private readonly billingService = inject(LlBillingService);

  readonly currentStep = signal<CheckoutStep>('direccion');
  readonly loading = signal(false);
  readonly loadingAddresses = signal(false);
  readonly error = signal<string | null>(null);
  readonly successMessage = signal<string | null>(null);
  
  readonly showProcessingModal = signal(false);
  readonly showSuccessModal = signal(false);
  readonly showErrorModal = signal(false);
  readonly completedOrderNumber = signal<string | null>(null);

  readonly direcciones = signal<Address[]>([]);
  readonly selectedDireccionId = signal<number | null>(null);
  readonly showNewAddressForm = signal(false);

  readonly selectedMetodoPago = signal<PaymentMethod | null>(null);
  readonly selectedModalidadPago = signal<PaymentMode>('contado');
  readonly selectedParcialidades = signal<number>(1);

  readonly regimenesFiscales = [
    { codigo: '601', nombre: 'General de Ley Personas Morales' },
    { codigo: '603', nombre: 'Personas Morales con Fines no Lucrativos' },
    { codigo: '605', nombre: 'Sueldos y Salarios e Ingresos Asimilados a Salarios' },
    { codigo: '606', nombre: 'Arrendamiento' },
    { codigo: '612', nombre: 'Personas Físicas con Actividades Empresariales y Profesionales' },
    { codigo: '616', nombre: 'Sin obligaciones fiscales' },
    { codigo: '621', nombre: 'Incorporación Fiscal' },
    { codigo: '626', nombre: 'Régimen Simplificado de Confianza' },
  ];

  readonly usosCfdi = [
    { codigo: 'G01', nombre: 'Adquisición de mercancías' },
    { codigo: 'G03', nombre: 'Gastos en general' },
    { codigo: 'I03', nombre: 'Equipo de transporte' },
    { codigo: 'S01', nombre: 'Sin efectos fiscales' },
  ];

  readonly estadosMexico = [
    'Aguascalientes', 'Baja California', 'Baja California Sur', 'Campeche',
    'Chiapas', 'Chihuahua', 'Ciudad de México', 'Coahuila', 'Colima',
    'Durango', 'Estado de México', 'Guanajuato', 'Guerrero', 'Hidalgo',
    'Jalisco', 'Michoacán', 'Morelos', 'Nayarit', 'Nuevo León', 'Oaxaca',
    'Puebla', 'Querétaro', 'Quintana Roo', 'San Luis Potosí', 'Sinaloa',
    'Sonora', 'Tabasco', 'Tamaulipas', 'Tlaxcala', 'Veracruz', 'Yucatán', 'Zacatecas'
  ];

  readonly addressForm = this.fb.group({
    alias: ['', [Validators.required]],
    calle: ['', [Validators.required]],
    numeroExterior: ['', [Validators.required]],
    numeroInterior: [''],
    colonia: ['', [Validators.required]],
    codigoPostal: ['', [Validators.required, Validators.pattern(/^\d{5}$/)]],
    ciudad: ['', [Validators.required]],
    estado: ['', [Validators.required]],
    referencias: [''],
    telefono: ['', [Validators.required, Validators.pattern(/^\d{10}$/)]],
  });

  readonly invoiceForm = this.fb.group({
    requiereFactura: [false],
    rfc: ['', [Validators.pattern(/^[A-ZÑ&]{3,4}\d{6}[A-Z0-9]{3}$/i)]],
    razonSocial: [''],
    regimenFiscal: [''],
    usoCfdi: [''],
    codigoPostalFiscal: ['', [Validators.pattern(/^\d{5}$/)]],
    email: ['', [Validators.email]],
  });

  readonly selectedDireccion = computed(() => {
    const id = this.selectedDireccionId();
    return this.direcciones().find(d => d.id === id) || null;
  });

  ngOnInit(): void {
    this.loadAddresses();
    this.loadDefaultBilling();
  }

  private loadDefaultBilling(): void {
    this.billingService.getDefault().subscribe({
      next: (billing) => {
        if (billing) {
          this.invoiceForm.patchValue({
            requiereFactura: true,
            rfc: billing.rfc,
            razonSocial: billing.razonSocial,
            regimenFiscal: billing.regimenFiscal,
            usoCfdi: billing.usoCfdi,
            codigoPostalFiscal: billing.postalCode,
            email: billing.email || '',
          });
        }
      },
      error: () => {}
    });
  }

  private loadAddresses(): void {
    this.loadingAddresses.set(true);
    this.addressService.list().subscribe({
      next: (addresses) => {
        this.direcciones.set(addresses);
        const defaultDir = addresses.find(d => d.isDefault);
        if (defaultDir) {
          this.selectedDireccionId.set(defaultDir.id);
        } else if (addresses.length > 0) {
          this.selectedDireccionId.set(addresses[0].id);
        }
        this.loadingAddresses.set(false);
      },
      error: () => {
        this.loadingAddresses.set(false);
      }
    });
  }

  goToStep(step: CheckoutStep): void {
    this.currentStep.set(step);
    this.error.set(null);
  }

  nextStep(): void {
    const current = this.currentStep();
    switch (current) {
      case 'direccion':
        if (this.showNewAddressForm()) {
          this.saveNewAddressAndContinue();
          return;
        }
        if (!this.selectedDireccionId()) {
          this.error.set('Selecciona una dirección de envío');
          return;
        }
        this.goToStep('facturacion');
        break;
      case 'facturacion':
        if (this.invoiceForm.value.requiereFactura && this.invoiceForm.invalid) {
          this.invoiceForm.markAllAsTouched();
          this.error.set('Completa los datos de facturación');
          return;
        }
        if (this.invoiceForm.value.requiereFactura) {
          this.saveBillingInfo();
        }
        this.goToStep('pago');
        break;
      case 'pago':
        if (!this.selectedMetodoPago()) {
          this.error.set('Selecciona un método de pago');
          return;
        }
        this.goToStep('confirmacion');
        break;
    }
  }

  prevStep(): void {
    const current = this.currentStep();
    switch (current) {
      case 'facturacion': this.goToStep('direccion'); break;
      case 'pago': this.goToStep('facturacion'); break;
      case 'confirmacion': this.goToStep('pago'); break;
    }
  }

  selectDireccion(id: number): void {
    this.selectedDireccionId.set(id);
    this.showNewAddressForm.set(false);
  }

  toggleNewAddressForm(): void {
    this.showNewAddressForm.update(v => !v);
    if (this.showNewAddressForm()) {
      this.selectedDireccionId.set(null);
    }
  }

  saveNewAddressAndContinue(): void {
    this.addressForm.markAllAsTouched();
    if (this.addressForm.invalid) {
      this.error.set('Completa todos los campos de la dirección');
      return;
    }

    const formValue = this.addressForm.value;
    const request: CreateAddressRequest = {
      alias: formValue.alias || 'Nueva dirección',
      street: formValue.calle || '',
      exteriorNumber: formValue.numeroExterior || '',
      interiorNumber: formValue.numeroInterior || undefined,
      neighborhood: formValue.colonia || '',
      postalCode: formValue.codigoPostal || '',
      city: formValue.ciudad || '',
      state: formValue.estado || '',
      reference: formValue.referencias || undefined,
      phone: formValue.telefono || '',
      isDefault: this.direcciones().length === 0,
    };

    this.loading.set(true);
    this.error.set(null);
    
    this.addressService.create(request).subscribe({
      next: (newAddress) => {
        this.direcciones.update(list => [...list, newAddress]);
        this.selectedDireccionId.set(newAddress.id);
        this.showNewAddressForm.set(false);
        this.addressForm.reset();
        this.loading.set(false);
        this.goToStep('facturacion');
      },
      error: (err) => {
        this.error.set(err.error?.message || 'Error al guardar la dirección.');
        this.loading.set(false);
      }
    });
  }

  selectMetodoPago(metodo: PaymentMethod): void {
    this.selectedMetodoPago.set(metodo);
  }

  selectModalidadPago(modalidad: PaymentMode): void {
    this.selectedModalidadPago.set(modalidad);
    if (modalidad !== 'parcialidades') {
      this.selectedParcialidades.set(1);
    }
  }

  selectParcialidades(num: number): void {
    this.selectedParcialidades.set(num);
  }

  getMetodoPagoLabel(metodo: PaymentMethod): string {
    return this.orderService.getPaymentMethodLabel(metodo);
  }

  getModalidadPagoLabel(modalidad: PaymentMode): string {
    return this.orderService.getPaymentModeLabel(modalidad);
  }

  private saveBillingInfo(): void {
    const formValue = this.invoiceForm.value;
    if (!formValue.rfc || !formValue.razonSocial) return;

    const request: CreateBillingRequest = {
      rfc: formValue.rfc.toUpperCase(),
      razonSocial: formValue.razonSocial,
      regimenFiscal: formValue.regimenFiscal || '',
      usoCfdi: formValue.usoCfdi || '',
      postalCode: formValue.codigoPostalFiscal || '',
      email: formValue.email || undefined,
      isDefault: true,
    };

    this.billingService.create(request).subscribe({ next: () => {}, error: () => {} });
  }

  confirmarPedido(): void {
    const direccion = this.selectedDireccion();
    const metodoPago = this.selectedMetodoPago();

    if (!direccion || !metodoPago) {
      this.error.set('Faltan datos para completar el pedido');
      return;
    }

    this.loading.set(true);
    this.error.set(null);

    const cartSummary = this.cart.getCartSummary();

    const request: CreateOrderRequest = {
      items: this.cart.items().map(ci => ({
        tireSku: ci.item.tire.sku,
        tireMeasure: ci.item.tire.medidaOriginal || ci.item.tire.sku,
        tireBrand: ci.item.tire.modelo || undefined,
        tireModel: ci.item.tire.modelo || undefined,
        quantity: ci.quantity,
        unitPrice: ci.item.price,
      })),
      shippingAddress: {
        street: direccion.street,
        exteriorNumber: direccion.exteriorNumber,
        interiorNumber: direccion.interiorNumber,
        neighborhood: direccion.neighborhood,
        postalCode: direccion.postalCode,
        city: direccion.city,
        state: direccion.state,
        reference: direccion.reference,
        phone: direccion.phone,
      },
      paymentMethod: metodoPago,
      paymentMode: this.selectedModalidadPago(),
      paymentInstallments: this.selectedModalidadPago() === 'parcialidades' ? this.selectedParcialidades() : undefined,
      requiresInvoice: this.invoiceForm.value.requiereFactura || false,
      billingInfo: this.invoiceForm.value.requiereFactura ? {
        rfc: this.invoiceForm.value.rfc || '',
        razonSocial: this.invoiceForm.value.razonSocial || '',
        regimenFiscal: this.invoiceForm.value.regimenFiscal || '',
        usoCfdi: this.invoiceForm.value.usoCfdi || '',
        postalCode: this.invoiceForm.value.codigoPostalFiscal || '',
        email: this.invoiceForm.value.email || undefined,
      } : undefined,
      customerNotes: undefined,
      subtotal: cartSummary.subtotal,
      iva: cartSummary.iva,
      total: cartSummary.total,
    };

    this.showProcessingModal.set(true);

    this.orderService.createOrder(request).subscribe({
      next: (order) => {
        this.loading.set(false);
        this.showProcessingModal.set(false);
        this.cart.clearCart();
        this.completedOrderNumber.set(order.orderNumber);
        this.showSuccessModal.set(true);
        
        if (isPlatformBrowser(this.platformId)) {
          window.dispatchEvent(new CustomEvent('ll-reload-inventory'));
        }
      },
      error: (err) => {
        this.loading.set(false);
        this.showProcessingModal.set(false);
        this.error.set(err.error?.message || 'Error al procesar el pedido.');
        this.showErrorModal.set(true);
      }
    });
  }

  closeSuccessModal(): void {
    this.showSuccessModal.set(false);
    this.router.navigate(['/cliente']).then(() => {
      setTimeout(() => {
        window.dispatchEvent(new CustomEvent('ll-change-tab', { detail: 'pedidos' }));
      }, 100);
    });
  }

  continueShopping(): void {
    this.showSuccessModal.set(false);
    this.router.navigate(['/cliente']).then(() => {
      setTimeout(() => {
        window.dispatchEvent(new CustomEvent('ll-change-tab', { detail: 'catalogo' }));
      }, 100);
    });
  }

  closeErrorModal(): void {
    this.showErrorModal.set(false);
  }

  cancelarCheckout(): void {
    this.router.navigate(['/cliente']);
  }

  formatPrice(price: number): string {
    return price.toLocaleString('es-MX', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  formatDireccion(d: Address): string {
    let dir = `${d.street} #${d.exteriorNumber}`;
    if (d.interiorNumber) dir += ` Int. ${d.interiorNumber}`;
    dir += `, ${d.neighborhood}, C.P. ${d.postalCode}, ${d.city}, ${d.state}`;
    return dir;
  }

  isAddressInvalid(controlName: string): boolean {
    const control = this.addressForm.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  isInvoiceInvalid(controlName: string): boolean {
    const control = this.invoiceForm.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }
}
