import { Injectable } from '@angular/core';
import { jsPDF } from 'jspdf';
import autoTable from 'jspdf-autotable';
import { Order } from '../clientes/services/ll-order.service';

// Datos de la empresa
const EMPRESA = {
  nombre: 'LLANTERA DE OCCIDENTE',
  rfc: 'LDO980101A01',
  direccion: 'Av. Periodismo José Tocaven Lavín 2649, Morelia, Mich.',
  telefono: '(443) 326 1312',
  email: 'ventas@llanteradeoccidente.com'
};

@Injectable({
  providedIn: 'root'
})
export class LlPdfGeneratorService {
  private logoBase64: string | null = null;
  private logoLoaded = false;

  constructor() {
    this.loadLogo();
  }

  private loadLogo(): void {
    const img = new Image();
    img.crossOrigin = 'Anonymous';
    img.onload = () => {
      const canvas = document.createElement('canvas');
      canvas.width = img.width;
      canvas.height = img.height;
      const ctx = canvas.getContext('2d');
      if (ctx) {
        ctx.drawImage(img, 0, 0);
        this.logoBase64 = canvas.toDataURL('image/png');
        this.logoLoaded = true;
      }
    };
    img.src = '/images/llantera/Logo_Llantera de Occidente.png';
  }

  /**
   * Genera un PDF de comprobante/nota de pedido
   */
  generateOrderReceipt(order: Order): void {
    const doc = new jsPDF();
    const pageWidth = doc.internal.pageSize.getWidth();
    
    // Colores corporativos
    const primaryColor: [number, number, number] = [230, 0, 18]; // #E60012
    const darkColor: [number, number, number] = [51, 51, 51];
    const grayColor: [number, number, number] = [128, 128, 128];

    let yPos = 10;

    // === ENCABEZADO ===
    // Línea superior decorativa
    doc.setFillColor(...primaryColor);
    doc.rect(0, 0, pageWidth, 5, 'F');

    // Logo a la izquierda
    const logoX = 15;
    const logoWidth = 45;
    const logoHeight = 18;
    
    if (this.logoBase64) {
      try {
        doc.addImage(this.logoBase64, 'PNG', logoX, yPos, logoWidth, logoHeight);
      } catch {
        // Si falla el logo, mostrar texto
        doc.setFontSize(14);
        doc.setFont('helvetica', 'bold');
        doc.setTextColor(...primaryColor);
        doc.text(EMPRESA.nombre, logoX, yPos + 10);
      }
    } else {
      doc.setFontSize(14);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(...primaryColor);
      doc.text(EMPRESA.nombre, logoX, yPos + 10);
    }

    // Datos de contacto a la derecha
    doc.setFontSize(8);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(...grayColor);
    doc.text(EMPRESA.direccion, pageWidth - 15, yPos + 5, { align: 'right' });
    doc.text(`Tel: ${EMPRESA.telefono}`, pageWidth - 15, yPos + 10, { align: 'right' });
    doc.text(`RFC: ${EMPRESA.rfc}`, pageWidth - 15, yPos + 15, { align: 'right' });

    yPos = 32;

    // Línea separadora
    doc.setDrawColor(...primaryColor);
    doc.setLineWidth(0.5);
    doc.line(15, yPos, pageWidth - 15, yPos);

    yPos += 8;

    // === TÍTULO DEL DOCUMENTO ===
    doc.setFontSize(14);
    doc.setFont('helvetica', 'bold');
    doc.setTextColor(...darkColor);
    doc.text('COMPROBANTE DE PEDIDO', pageWidth / 2, yPos, { align: 'center' });

    yPos += 10;

    // === INFORMACIÓN DEL PEDIDO (compacta) ===
    const leftCol = 15;
    const midCol = pageWidth / 2 + 5;

    // Cuadro de información más compacto
    doc.setFillColor(248, 248, 248);
    doc.roundedRect(leftCol, yPos - 3, pageWidth - 30, 22, 2, 2, 'F');

    doc.setFontSize(9);
    doc.setFont('helvetica', 'bold');
    doc.setTextColor(...darkColor);
    
    // Columna izquierda
    doc.text('Pedido:', leftCol + 3, yPos + 4);
    doc.text('Fecha:', leftCol + 3, yPos + 11);
    doc.text('Estado:', leftCol + 3, yPos + 18);

    doc.setFont('helvetica', 'normal');
    doc.text(order.orderNumber, leftCol + 25, yPos + 4);
    doc.text(this.formatDate(order.createdAt), leftCol + 25, yPos + 11);
    doc.text(this.getStatusLabel(order.status), leftCol + 25, yPos + 18);

    // Columna derecha
    doc.setFont('helvetica', 'bold');
    doc.text('Método de Pago:', midCol, yPos + 4);
    doc.text('Modalidad:', midCol, yPos + 11);
    doc.text('Requiere Factura:', midCol, yPos + 18);

    doc.setFont('helvetica', 'normal');
    doc.text(this.getPaymentMethodLabel(order.paymentMethod), midCol + 38, yPos + 4);
    let modalidadText = this.getPaymentModeLabel(order.paymentMode || 'contado');
    if (order.paymentMode === 'parcialidades' && order.paymentInstallments) {
      modalidadText += ` (${order.paymentInstallments} meses)`;
    }
    doc.text(modalidadText, midCol + 38, yPos + 11);
    doc.text(order.requiresInvoice ? 'Sí' : 'No', midCol + 38, yPos + 18);

    yPos += 28;

    // === TABLA DE PRODUCTOS ===
    doc.setFontSize(10);
    doc.setFont('helvetica', 'bold');
    doc.setTextColor(...primaryColor);
    doc.text('DETALLE DEL PEDIDO', leftCol, yPos);

    yPos += 3;

    const tableData = (order.items || []).map(item => [
      item.quantity.toString(),
      `${item.tireMeasure}\n${item.tireBrand || ''} ${item.tireModel || ''}\nSKU: ${item.tireSku}`,
      this.formatCurrency(item.unitPrice),
      this.formatCurrency(item.subtotal)
    ]);

    autoTable(doc, {
      startY: yPos,
      head: [['Cant.', 'Descripción', 'Precio Unit.', 'Subtotal']],
      body: tableData,
      theme: 'striped',
      headStyles: {
        fillColor: primaryColor,
        textColor: [255, 255, 255],
        fontStyle: 'bold',
        fontSize: 9
      },
      bodyStyles: {
        fontSize: 8,
        textColor: darkColor
      },
      columnStyles: {
        0: { halign: 'center', cellWidth: 15 },
        1: { cellWidth: 'auto' },
        2: { halign: 'right', cellWidth: 28 },
        3: { halign: 'right', cellWidth: 28 }
      },
      margin: { left: leftCol, right: 15 }
    });

    // @ts-ignore - autoTable adds this property
    yPos = doc.lastAutoTable.finalY + 8;

    // === TOTALES (a la derecha) ===
    const totalsX = pageWidth - 65;
    const totalsWidth = 50;
    
    doc.setFillColor(248, 248, 248);
    doc.roundedRect(totalsX - 5, yPos - 3, totalsWidth + 10, 32, 2, 2, 'F');

    doc.setFontSize(9);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(...darkColor);
    
    doc.text('Subtotal:', totalsX, yPos + 4);
    doc.text(this.formatCurrency(order.subtotal), pageWidth - 18, yPos + 4, { align: 'right' });

    doc.text('IVA (16%):', totalsX, yPos + 11);
    doc.text(this.formatCurrency(order.iva || 0), pageWidth - 18, yPos + 11, { align: 'right' });

    doc.text('Envío:', totalsX, yPos + 18);
    doc.text(order.shippingCost > 0 ? this.formatCurrency(order.shippingCost) : 'Gratis', pageWidth - 18, yPos + 18, { align: 'right' });

    doc.setFont('helvetica', 'bold');
    doc.setFontSize(10);
    doc.setTextColor(...primaryColor);
    doc.text('TOTAL:', totalsX, yPos + 27);
    doc.text(this.formatCurrency(order.total), pageWidth - 18, yPos + 27, { align: 'right' });

    yPos += 40;

    // === DIRECCIÓN DE ENVÍO Y FACTURACIÓN (lado a lado) ===
    const addr = order.shippingAddress;
    const colWidth = (pageWidth - 35) / 2;
    
    // Dirección de envío (izquierda)
    doc.setFontSize(9);
    doc.setFont('helvetica', 'bold');
    doc.setTextColor(...primaryColor);
    doc.text('DIRECCIÓN DE ENVÍO', leftCol, yPos);

    yPos += 5;
    doc.setFontSize(8);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(...darkColor);
    
    if (addr) {
      let direccion = `${addr.street} #${addr.exteriorNumber}`;
      if (addr.interiorNumber) direccion += ` Int. ${addr.interiorNumber}`;
      doc.text(direccion, leftCol, yPos);
      doc.text(`Col. ${addr.neighborhood}, C.P. ${addr.postalCode}`, leftCol, yPos + 4);
      doc.text(`${addr.city}, ${addr.state}`, leftCol, yPos + 8);
      doc.text(`Tel: ${addr.phone}`, leftCol, yPos + 12);
    }

    // Datos de facturación (derecha) - si aplica
    if (order.requiresInvoice && order.billingInfo) {
      const rightColX = leftCol + colWidth + 5;
      
      doc.setFontSize(9);
      doc.setFont('helvetica', 'bold');
      doc.setTextColor(...primaryColor);
      doc.text('DATOS DE FACTURACIÓN', rightColX, yPos - 5);

      doc.setFontSize(8);
      doc.setFont('helvetica', 'normal');
      doc.setTextColor(...darkColor);
      doc.text(`RFC: ${order.billingInfo.rfc}`, rightColX, yPos);
      doc.text(`Razón Social: ${order.billingInfo.razonSocial}`, rightColX, yPos + 4);
      doc.text(`Uso CFDI: ${order.billingInfo.usoCfdi}`, rightColX, yPos + 8);
    }

    // === PIE DE PÁGINA ===
    const footerY = doc.internal.pageSize.getHeight() - 20;
    
    doc.setDrawColor(...grayColor);
    doc.setLineWidth(0.3);
    doc.line(15, footerY - 5, pageWidth - 15, footerY - 5);

    doc.setFontSize(7);
    doc.setFont('helvetica', 'normal');
    doc.setTextColor(...grayColor);
    doc.text('Este documento es un comprobante de pedido y no tiene validez fiscal.', pageWidth / 2, footerY, { align: 'center' });
    doc.text(`Generado el ${this.formatDate(new Date().toISOString())} | ${EMPRESA.nombre}`, pageWidth / 2, footerY + 4, { align: 'center' });

    // Línea inferior decorativa
    doc.setFillColor(...primaryColor);
    doc.rect(0, doc.internal.pageSize.getHeight() - 5, pageWidth, 5, 'F');

    // Guardar PDF
    doc.save(`Comprobante_${order.orderNumber}.pdf`);
  }

  private formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return new Intl.DateTimeFormat('es-MX', {
      day: '2-digit',
      month: 'long',
      year: 'numeric'
    }).format(date);
  }

  private formatCurrency(amount: number): string {
    return new Intl.NumberFormat('es-MX', {
      style: 'currency',
      currency: 'MXN'
    }).format(amount);
  }

  private getStatusLabel(status: string): string {
    const labels: Record<string, string> = {
      solicitado: 'Solicitado',
      preparando: 'En preparación',
      enviado: 'Enviado',
      entregado: 'Entregado',
      cancelado: 'Cancelado'
    };
    return labels[status] || status;
  }

  private getPaymentMethodLabel(method: string): string {
    const labels: Record<string, string> = {
      transferencia: 'Transferencia bancaria',
      tarjeta: 'Tarjeta de crédito/débito',
      efectivo: 'Pago en efectivo'
    };
    return labels[method] || method;
  }

  private getPaymentModeLabel(mode: string): string {
    const labels: Record<string, string> = {
      contado: 'Pago de contado',
      credito: 'Crédito',
      parcialidades: 'Parcialidades',
      anticipo: 'Anticipo'
    };
    return labels[mode] || mode;
  }
}
