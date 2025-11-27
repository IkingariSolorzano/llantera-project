import { CommonModule } from '@angular/common';
import { Component, inject, signal, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { 
  LlPriceLevelsService, 
  LlPriceLevelDTO, 
  CreatePriceLevelRequest 
} from './ll-price-levels.service';
import { LlLlantasPriceColumnsService, LlPriceColumnDTO } from '../llantas/ll-llantas-price-columns.service';

@Component({
  selector: 'app-ll-price-levels',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './ll-price-levels.component.html',
})
export class LlPriceLevelsComponent implements OnInit {
  private readonly priceLevelsService = inject(LlPriceLevelsService);
  private readonly priceColumnsService = inject(LlLlantasPriceColumnsService);

  readonly levels = signal<LlPriceLevelDTO[]>([]);
  readonly priceColumns = signal<LlPriceColumnDTO[]>([]);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly showModal = signal(false);
  readonly editingLevel = signal<LlPriceLevelDTO | null>(null);

  // Form fields
  readonly formData = signal<CreatePriceLevelRequest>({
    code: '',
    name: '',
    description: null,
    discountPercentage: 0,
    priceColumn: 'public',
    referenceColumn: null,
    canViewOffers: false,
  });

  ngOnInit(): void {
    this.loadLevels();
    this.loadPriceColumns();
  }

  loadLevels(): void {
    this.loading.set(true);
    this.error.set(null);

    this.priceLevelsService.list().subscribe({
      next: (response) => {
        this.levels.set(response.items);
        this.loading.set(false);
      },
      error: (err) => {
        console.error('Error cargando niveles de precio:', err);
        this.error.set('Error al cargar los niveles de precio');
        this.loading.set(false);
      }
    });
  }

  loadPriceColumns(): void {
    this.priceColumnsService.list().subscribe({
      next: (columns: LlPriceColumnDTO[]) => {
        this.priceColumns.set(columns);
      },
      error: (err: unknown) => {
        console.error('Error cargando columnas de precio:', err);
      }
    });
  }

  openCreateModal(): void {
    this.editingLevel.set(null);
    this.formData.set({
      code: '',
      name: '',
      description: null,
      discountPercentage: 0,
      priceColumn: 'public',
      referenceColumn: null,
      canViewOffers: false,
    });
    this.showModal.set(true);
  }

  openEditModal(level: LlPriceLevelDTO): void {
    this.editingLevel.set(level);
    this.formData.set({
      code: level.code,
      name: level.name,
      description: level.description ?? null,
      discountPercentage: level.discountPercentage,
      priceColumn: level.priceColumn,
      referenceColumn: level.referenceColumn ?? null,
      canViewOffers: level.canViewOffers,
    });
    this.showModal.set(true);
  }

  closeModal(): void {
    this.showModal.set(false);
    this.editingLevel.set(null);
  }

  updateField<K extends keyof CreatePriceLevelRequest>(field: K, value: CreatePriceLevelRequest[K]): void {
    this.formData.update(data => ({ ...data, [field]: value }));
  }

  saveLevel(): void {
    this.loading.set(true);
    const data = this.formData();
    const editing = this.editingLevel();

    const request$ = editing 
      ? this.priceLevelsService.update(editing.id, data)
      : this.priceLevelsService.create(data);

    request$.subscribe({
      next: () => {
        this.loading.set(false);
        this.closeModal();
        this.loadLevels();
      },
      error: (err) => {
        console.error('Error guardando nivel:', err);
        this.error.set(err.error?.message || 'Error al guardar el nivel de precio');
        this.loading.set(false);
      }
    });
  }

  deleteLevel(level: LlPriceLevelDTO): void {
    if (!confirm(`¿Estás seguro de eliminar el nivel "${level.name}"?`)) {
      return;
    }

    this.loading.set(true);
    this.priceLevelsService.delete(level.id).subscribe({
      next: () => {
        this.loading.set(false);
        this.loadLevels();
      },
      error: (err) => {
        console.error('Error eliminando nivel:', err);
        this.error.set(err.error?.message || 'Error al eliminar el nivel de precio');
        this.loading.set(false);
      }
    });
  }

  getColumnName(code: string): string {
    const column = this.priceColumns().find(c => c.code === code);
    return column?.label ?? code;
  }
}
