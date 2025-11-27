import { Component, HostListener } from '@angular/core';
import { RouterLink } from '@angular/router';
import { SellerOverviewComponent } from './seller-overview/seller-overview.component';
import { ProductsComponent } from './products/products.component';
import { NgIf } from '@angular/common';

@Component({
    selector: 'app-e-seller-details',
    standalone: true,
    imports: [RouterLink, NgIf, SellerOverviewComponent, ProductsComponent],
    templateUrl: './e-seller-details.component.html',
    styleUrl: './e-seller-details.component.scss'
})
export class ESellerDetailsComponent {

    // Card Header Menu
    isCardHeaderOpen = false;
    toggleCardHeaderMenu() {
        this.isCardHeaderOpen = !this.isCardHeaderOpen;
    }
    @HostListener('document:click', ['$event'])
    handleClickOutside(event: Event) {
        const target = event.target as HTMLElement;
        if (!target.closest('.trezo-card-dropdown')) {
            this.isCardHeaderOpen = false;
        }
    }

}