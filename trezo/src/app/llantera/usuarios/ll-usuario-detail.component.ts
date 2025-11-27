import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { finalize } from 'rxjs/operators';

import { LlUsuariosService, LlUsuarioDTO } from './ll-usuarios.service';
import { LlEmpresasService, LlEmpresaDTO } from '../empresas/ll-empresas.service';

@Component({
  selector: 'app-ll-usuario-detail',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './ll-usuario-detail.component.html',
  styleUrl: './ll-usuario-detail.component.scss',
})
export class LlUsuarioDetailComponent {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly usuariosService = inject(LlUsuariosService);
  private readonly empresasService = inject(LlEmpresasService);

  readonly loading = signal(true);
  readonly error = signal<string | null>(null);
  readonly usuario = signal<LlUsuarioDTO | null>(null);
  readonly company = signal<LlEmpresaDTO | null>(null);
  readonly pageTitle = computed(() =>
    this.usuario() ? this.usuario()!.name || 'Detalle de usuario' : 'Detalle de usuario'
  );

  constructor() {
    this.route.paramMap.pipe(takeUntilDestroyed()).subscribe((params) => {
      const id = params.get('id');
      if (!id) {
        this.error.set('Identificador inválido.');
        this.loading.set(false);
        return;
      }
      this.loadUsuario(id);
    });
  }

  goBack(): void {
    this.router.navigate(['/dashboard/usuarios']);
  }

  goToEdit(): void {
    const usuario = this.usuario();
    if (!usuario) {
      return;
    }
    this.router.navigate(['/dashboard/usuarios', usuario.id, 'editar']);
  }

  private loadUsuario(id: string): void {
    this.loading.set(true);
    this.error.set(null);

    this.usuariosService
      .getById(id)
      .pipe(
        finalize(() => this.loading.set(false)),
        takeUntilDestroyed()
      )
      .subscribe({
        next: (usuario) => {
          this.usuario.set(usuario);
          this.loadCompanyForUsuario(usuario);
        },
        error: () => this.error.set('No se pudo cargar la información del usuario.'),
      });
  }

  formatRole(usuario: LlUsuarioDTO | null | undefined): string {
    const role = (usuario?.role || 'customer').toLowerCase();
    return role === 'admin' ? 'Administrador' : 'Cliente';
  }

  formatNivel(usuario: LlUsuarioDTO | null | undefined): string {
    const level = (usuario?.level || 'public').toLowerCase();
    switch (level) {
      case 'public':
        return 'Público';
      case 'empresa':
        return 'Empresa';
      case 'distribuidor':
        return 'Distribuidor';
      case 'mayorista':
        return 'Mayorista';
      case 'silver':
        return 'Plata';
      case 'gold':
        return 'Oro';
      case 'platinum':
        return 'Platino';
      default:
        return usuario?.level || 'public';
    }
  }

  private loadCompanyForUsuario(usuario: LlUsuarioDTO): void {
    if (!usuario.companyId) {
      this.company.set(null);
      return;
    }

    this.empresasService
      .getById(usuario.companyId)
      .pipe(takeUntilDestroyed())
      .subscribe({
        next: (empresa) => this.company.set(empresa),
        error: (err) => {
          console.error('Error al cargar empresa del usuario', err);
          this.company.set(null);
        },
      });
  }
}
