import { Routes } from '@angular/router';
import { DashboardComponent } from './dashboard/dashboard.component';
// Landing page desactivada temporalmente
// import { LlLandingPageComponent } from './llantera/landing/ll-landing-page.component';
import { LlLoginPageComponent } from './llantera/auth/ll-login-page.component';
import { UsersPageComponent } from './pages/users-page/users-page.component';
import { LlUsuariosListComponent } from './llantera/usuarios/ll-usuarios-list.component';
import { LlUsuarioFormComponent } from './llantera/usuarios/ll-usuario-form.component';
import { LlUsuarioDetailComponent } from './llantera/usuarios/ll-usuario-detail.component';
import { LlPriceLevelsComponent } from './llantera/usuarios/ll-price-levels.component';
// Empresas desactivadas temporalmente
// import { CrmPageComponent } from './pages/crm-page/crm-page.component';
// import { CContactsComponent } from './pages/crm-page/c-contacts/c-contacts.component';
// import { CLeadsComponent } from './pages/crm-page/c-leads/c-leads.component';
// import { CDealsComponent } from './pages/crm-page/c-deals/c-deals.component';
import { EcommercePageComponent } from './pages/ecommerce-page/ecommerce-page.component';
import { EProductsGridComponent } from './pages/ecommerce-page/e-products-grid/e-products-grid.component';
import { EProductsListComponent } from './pages/ecommerce-page/e-products-list/e-products-list.component';
import { EProductDetailsComponent } from './pages/ecommerce-page/e-product-details/e-product-details.component';
import { LlClientesCatalogPageComponent } from './llantera/clientes/catalogo/ll-clientes-catalog-page.component';
import { LlClientesTireDetailComponent } from './llantera/clientes/catalogo/ll-clientes-tire-detail.component';
import { LlClientePanelComponent } from './llantera/clientes/ll-cliente-panel.component';
import { LlClienteCheckoutComponent } from './llantera/clientes/checkout/ll-cliente-checkout.component';
import { ECategoriesComponent } from './pages/ecommerce-page/e-categories/e-categories.component';
import { ESellersComponent } from './pages/ecommerce-page/e-sellers/e-sellers.component';
import { ESellerDetailsComponent } from './pages/ecommerce-page/e-seller-details/e-seller-details.component';
// Empresas desactivadas temporalmente
// import { LlEmpresasListComponent } from './llantera/empresas/ll-empresas-list.component';
// import { LlEmpresaFormComponent } from './llantera/empresas/ll-empresa-form.component';
import { LlLlantasListComponent } from './llantera/llantas/ll-llantas-list.component';
import { LlLlantasFormComponent } from './llantera/llantas/ll-llantas-form.component';
import { LlLlantasPriceColumnsListComponent } from './llantera/llantas/ll-llantas-price-columns-list.component';
import { LlMarcasListComponent } from './llantera/marcas/ll-marcas-list.component';
import { LlAdminPedidosComponent } from './llantera/pedidos/ll-admin-pedidos.component';
import { LlAdminVentasComponent } from './llantera/ventas/ll-admin-ventas.component';
// Solicitudes desactivadas temporalmente
// import { LlSolicitudesListComponent } from './llantera/solicitudes/ll-solicitudes-list.component';
// import { RequestsPageComponent } from './pages/requests-page/requests-page.component';
import {
    adminEmployeeGuard,
    customerGuard,
    guestGuard,
} from './llantera/auth/ll-auth.guards';

export const routes: Routes = [
    // Landing page desactivada - redirige a login (guestGuard redirigirá al panel si ya está logueado)
    {path: '', pathMatch: 'full', redirectTo: 'login'},
    {path: 'login', component: LlLoginPageComponent, canActivate: [guestGuard]},
    {path: 'cliente', component: LlClientePanelComponent, canActivate: [customerGuard]},
    {path: 'cliente/checkout', component: LlClienteCheckoutComponent, canActivate: [customerGuard]},
    {
        path: 'dashboard',
        component: DashboardComponent,
        canActivate: [adminEmployeeGuard],
        children: [
            {path: '', redirectTo: 'usuarios', pathMatch: 'full'},
            {
                path: 'usuarios',
                component: UsersPageComponent,
                children: [
                    {path: '', component: LlUsuariosListComponent},
                    {path: 'lista', component: LlUsuariosListComponent},
                    {path: 'nuevo', component: LlUsuarioFormComponent},
                    {path: ':id/detalle', component: LlUsuarioDetailComponent},
                    {path: ':id/editar', component: LlUsuarioFormComponent},
                    {path: 'niveles-precio', component: LlPriceLevelsComponent}
                ]
            },
            // Empresas desactivadas temporalmente
            // {
            //     path: 'empresas',
            //     component: CrmPageComponent,
            //     children: [
            //         {path: '', component: LlEmpresasListComponent},
            //         {path: 'lista', component: LlEmpresasListComponent},
            //         {path: 'nueva', component: LlEmpresaFormComponent},
            //         {path: ':id/editar', component: LlEmpresaFormComponent},
            //         {path: 'contactos', component: CContactsComponent},
            //         {path: 'leads', component: CLeadsComponent},
            //         {path: 'negocios', component: CDealsComponent}
            //     ]
            // },
            {
                path: 'llantas',
                component: EcommercePageComponent,
                children: [
                    {path: '', component: LlLlantasListComponent},
                    {path: 'nueva', component: LlLlantasFormComponent},
                    {path: 'editar/:sku', component: LlLlantasFormComponent},
                    {path: 'precios', component: LlLlantasPriceColumnsListComponent},
                    {path: 'catalogo-publico', component: LlClientesCatalogPageComponent},
                    {path: 'catalogo-publico/:sku/ver', component: LlClientesTireDetailComponent},
                    {path: 'lista', component: EProductsListComponent},
                    {path: 'detalle', component: EProductDetailsComponent}
                ]
            },
            {
                path: 'marcas',
                component: EcommercePageComponent,
                children: [
                    {path: '', component: LlMarcasListComponent},
                    {path: 'proveedores', component: ESellersComponent},
                    {path: 'proveedores/:id', component: ESellerDetailsComponent}
                ]
            },
            {
                path: 'pedidos',
                component: EcommercePageComponent,
                children: [
                    {path: '', component: LlAdminPedidosComponent}
                ]
            },
            {
                path: 'ventas',
                component: EcommercePageComponent,
                children: [
                    {path: '', component: LlAdminVentasComponent}
                ]
            }
            // Solicitudes desactivadas temporalmente
            // {
            //     path: 'solicitudes',
            //     component: RequestsPageComponent,
            //     children: [
            //         {path: '', component: LlSolicitudesListComponent}
            //     ]
            // }
        ]
    },
    {path: '**', redirectTo: 'login'}
];
