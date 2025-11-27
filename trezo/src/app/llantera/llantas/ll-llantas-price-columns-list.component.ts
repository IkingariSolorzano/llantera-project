import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, inject, signal } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs/operators';
import { firstValueFrom } from 'rxjs';

import {
  LlLlantasPriceColumnsService,
  LlPriceColumnDTO,
  LlPriceColumnUpsertPayload,
} from './ll-llantas-price-columns.service';

type LlDeleteDependentAction = 'fixed' | 'changeBase';

interface LlDeleteDependentState {
  columnId: number;
  label: string;
  code: string;
  action: LlDeleteDependentAction;
  newBaseCode: string | null;
}

@Component({
  selector: 'app-ll-llantas-price-columns-list',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './ll-llantas-price-columns-list.component.html',
})
export class LlLlantasPriceColumnsListComponent {
  private readonly service = inject(LlLlantasPriceColumnsService);
  private readonly fb = inject(FormBuilder);
  private readonly destroyRef = inject(DestroyRef);

  readonly columns = signal<LlPriceColumnDTO[]>([]);
  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal<string | null>(null);
  readonly editingId = signal<number | null>(null);

  readonly showDeleteAssist = signal(false);
  readonly deleteBaseColumn = signal<LlPriceColumnDTO | null>(null);
  readonly deleteDependentsState = signal<LlDeleteDependentState[]>([]);

  readonly form = this.fb.group({
    code: ['', [Validators.required, Validators.maxLength(64), Validators.pattern(/^[A-Za-z0-9_]+$/)]],
    label: ['', [Validators.required, Validators.maxLength(120)]],
    description: [''],
    visualOrder: [0, [Validators.min(0)]],
    active: [true],
    isPublicPrice: [false],
    mode: ['fixed'],
    baseCode: [''],
    operation: ['percent'],
    amount: [null as number | null],
  });

  readonly pageTitle = computed(() => 'Columnas de precio de llantas');
  readonly submitLabel = computed(() => (this.editingId() ? 'Guardar cambios' : 'Crear columna'));

  constructor() {
    this.loadColumns();
  }

  isInvalid(controlName: string): boolean {
    const control = this.form.get(controlName);
    return !!control && control.invalid && (control.dirty || control.touched);
  }

  onCodeInput(event: Event): void {
    const input = event.target as HTMLInputElement | null;
    if (!input) {
      return;
    }
    const original = input.value;
    const transformed = original.replace(/\s+/g, '_');
    if (transformed !== original) {
      this.form.controls.code.setValue(transformed, { emitEvent: false });
      input.value = transformed;
    }
  }

  onCreateNew(): void {
    this.editingId.set(null);
    this.error.set(null);
    this.success.set(null);
    this.form.reset({
      code: '',
      label: '',
      description: '',
      visualOrder: 0,
      active: true,
      isPublicPrice: false,
      mode: 'fixed',
      baseCode: '',
      operation: 'percent',
      amount: null,
    });
  }

  onEdit(column: LlPriceColumnDTO): void {
    const codeLower = (column.code ?? '').toLowerCase();
    if (codeLower === 'lista') {
      this.error.set('La columna "lista" no se puede editar.');
      return;
    }
    this.editingId.set(column.id);
    const calc = column.calculation;
    this.error.set(null);
    this.success.set(null);
    this.form.reset({
      code: column.code,
      label: column.label,
      description: column.description ?? '',
      visualOrder: column.visualOrder ?? 0,
      active: column.active,
      isPublicPrice: column.isPublicPrice ?? false,
      mode: calc?.mode ?? 'fixed',
      baseCode: calc?.baseCode ?? '',
      operation: calc?.operation ?? 'percent',
      amount: calc?.amount ?? null,
    });
  }

  onDelete(column: LlPriceColumnDTO): void {
    if (!column.id || this.saving()) {
      return;
    }
    const codeLower = (column.code ?? '').toLowerCase();
    if (codeLower === 'lista') {
      this.error.set('La columna "lista" no se puede eliminar.');
      return;
    }

    // Verificar si esta columna es base de otras columnas derivadas.
    const dependents = this.columns().filter((c) => {
      if (!c.calculation || c.calculation.mode !== 'derived') {
        return false;
      }
      const baseCode = (c.calculation.baseCode ?? '').toLowerCase();
      return c.id !== column.id && baseCode === codeLower;
    });

    if (dependents.length > 0) {
      // Abrir flujo asistido para reconfigurar derivadas antes de eliminar la columna base.
      const initialState: LlDeleteDependentState[] = dependents.map((d) => ({
        columnId: d.id,
        label: d.label,
        code: d.code,
        action: 'fixed',
        newBaseCode: null,
      }));
      this.deleteBaseColumn.set(column);
      this.deleteDependentsState.set(initialState);
      this.showDeleteAssist.set(true);
      this.error.set(null);
      this.success.set(null);
      return;
    }

    const confirmed = window.confirm(
      `¿Estás seguro de que deseas eliminar la columna de precio "${column.label || column.code}"? Esto eliminará sus precios asociados en todas las llantas.`,
    );
    if (!confirmed) {
      return;
    }

    this.error.set(null);
    this.success.set(null);
    this.saving.set(true);

    this.service
      .delete(column.id)
      .pipe(
        finalize(() => this.saving.set(false)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: () => {
          this.success.set('Columna eliminada correctamente.');
          this.loadColumns();
          if (this.editingId() === column.id) {
            this.onCreateNew();
          }
        },
        error: (err) => {
          console.error('Error al eliminar columna de precio', err);
          this.error.set('No se pudo eliminar la columna de precio.');
        },
      });
  }

  onChangeDeleteDependentAction(index: number, raw: string): void {
    const action: LlDeleteDependentAction = raw === 'changeBase' ? 'changeBase' : 'fixed';
    const current = [...this.deleteDependentsState()];
    if (!current[index]) {
      return;
    }
    current[index] = {
      ...current[index],
      action,
    };
    this.deleteDependentsState.set(current);
  }

  onChangeDeleteDependentBase(index: number, baseCode: string): void {
    const value = (baseCode ?? '').toString().trim();
    const current = [...this.deleteDependentsState()];
    if (!current[index]) {
      return;
    }
    current[index] = {
      ...current[index],
      newBaseCode: value || null,
    };
    this.deleteDependentsState.set(current);
  }

  onCancelDeleteAssist(): void {
    this.showDeleteAssist.set(false);
    this.deleteBaseColumn.set(null);
    this.deleteDependentsState.set([]);
  }

  async onConfirmDeleteAssist(): Promise<void> {
    const baseColumn = this.deleteBaseColumn();
    const states = this.deleteDependentsState();
    if (!baseColumn || states.length === 0) {
      this.showDeleteAssist.set(false);
      return;
    }

    this.error.set(null);
    this.success.set(null);
    this.saving.set(true);

    try {
      const codeRegex = /^[A-Za-z0-9_]+$/;

      for (const state of states) {
        const dep = this.columns().find((c) => c.id === state.columnId);
        if (!dep || !dep.id) {
          continue;
        }

        if (state.action === 'fixed') {
          const payload: LlPriceColumnUpsertPayload = {
            code: dep.code,
            label: dep.label,
            description: dep.description ?? '',
            visualOrder: dep.visualOrder ?? 0,
            active: dep.active,
            isPublicPrice: dep.isPublicPrice ?? false,
            calculation: null,
          };
          await firstValueFrom(this.service.update(dep.id, payload));
        } else {
          const newBase = (state.newBaseCode ?? '').toString().trim().toLowerCase();
          if (!newBase || !codeRegex.test(newBase)) {
            this.error.set('Debes seleccionar una columna base válida para todas las columnas derivadas.');
            this.saving.set(false);
            return;
          }

          const calc = dep.calculation;
          const payload: LlPriceColumnUpsertPayload = {
            code: dep.code,
            label: dep.label,
            description: dep.description ?? '',
            visualOrder: dep.visualOrder ?? 0,
            active: dep.active,
            isPublicPrice: dep.isPublicPrice ?? false,
            calculation: {
              mode: 'derived',
              baseCode: newBase,
              operation: calc?.operation ?? 'percent',
              amount: calc?.amount ?? 0,
            },
          };
          await firstValueFrom(this.service.update(dep.id, payload));
        }
      }

      // Una vez actualizadas las columnas derivadas, eliminamos la columna base.
      await firstValueFrom(this.service.delete(baseColumn.id));

      this.success.set(
        `Columna "${baseColumn.label || baseColumn.code}" y columnas derivadas actualizadas correctamente.`,
      );
      this.showDeleteAssist.set(false);
      this.deleteBaseColumn.set(null);
      this.deleteDependentsState.set([]);
      this.loadColumns();
      if (this.editingId() === baseColumn.id) {
        this.onCreateNew();
      }
    } catch (err) {
      console.error('Error en el flujo asistido de eliminación de columna base', err);
      this.error.set('Ocurrió un error al actualizar columnas derivadas o eliminar la columna base.');
    } finally {
      this.saving.set(false);
    }
  }

  onSubmit(): void {
    this.error.set(null);
    this.success.set(null);
    this.form.markAllAsTouched();

    if (this.form.invalid) {
      this.error.set('Por favor revisa los campos marcados en rojo.');
      return;
    }

    const raw = this.form.value;
    const payload: LlPriceColumnUpsertPayload = {
      code: (raw.code ?? '').toString().trim(),
      label: (raw.label ?? '').toString().trim(),
      description: (raw.description ?? '').toString(),
      visualOrder: typeof raw.visualOrder === 'number' ? raw.visualOrder : Number(raw.visualOrder ?? 0) || 0,
      active: !!raw.active,
      isPublicPrice: !!raw.isPublicPrice,
      calculation: null,
    };

    const codeRegex = /^[A-Za-z0-9_]+$/;
    if (!payload.code || !codeRegex.test(payload.code)) {
      this.error.set('El código solo puede contener letras, números y guiones bajos, sin espacios.');
      return;
    }
    payload.code = payload.code.toLowerCase();

    if (!payload.label) {
      this.error.set('El nombre es obligatorio.');
      return;
    }

    const modeRaw = (raw.mode ?? 'fixed').toString().trim().toLowerCase();
    const mode = modeRaw === 'derived' ? 'derived' : 'fixed';

    if (mode === 'derived') {
      const baseCodeRaw = (raw.baseCode ?? '').toString().trim();
      if (!baseCodeRaw || !codeRegex.test(baseCodeRaw)) {
        this.error.set('La columna base es obligatoria y solo puede contener letras, números y guiones bajos.');
        return;
      }

      const operationRaw = (raw.operation ?? '').toString().trim().toLowerCase();
      const allowedOperations = ['add', 'subtract', 'multiply', 'percent'];
      if (!allowedOperations.includes(operationRaw)) {
        this.error.set('La operación de cálculo no es válida.');
        return;
      }

      const amountRaw = raw.amount;
      const amount =
        typeof amountRaw === 'number' ? amountRaw : amountRaw !== null && amountRaw !== undefined && amountRaw !== ''
          ? Number(amountRaw)
          : NaN;
      if (!Number.isFinite(amount)) {
        this.error.set('La cantidad de cálculo debe ser un número válido.');
        return;
      }

      payload.calculation = {
        mode: 'derived',
        baseCode: baseCodeRaw.toLowerCase(),
        operation: operationRaw as any,
        amount,
      };
    } else {
      payload.calculation = null;
    }

    this.saving.set(true);

    const id = this.editingId();
    const request$ = id ? this.service.update(id, payload) : this.service.create(payload);

    request$
      .pipe(
        finalize(() => this.saving.set(false)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: () => {
          this.success.set(id ? 'Columna actualizada correctamente.' : 'Columna creada correctamente.');
          if (!id) {
            this.onCreateNew();
          }
          this.loadColumns();
        },
        error: (err) => {
          console.error('Error al guardar columna de precio', err);
          this.error.set('No se pudo guardar la columna de precio.');
        },
      });
  }

  private loadColumns(): void {
    this.loading.set(true);
    this.error.set(null);

    this.service
      .list()
      .pipe(
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe({
        next: (items) => {
          this.columns.set(items ?? []);
        },
        error: (err) => {
          console.error('Error al cargar columnas de precio', err);
          this.error.set('No se pudieron cargar las columnas de precio.');
          this.columns.set([]);
        },
      });
  }
}
