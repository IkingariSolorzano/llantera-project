import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { QuantityCounterComponent } from './quantity-counter/quantity-counter.component';
import { PdManageReviewsComponent } from './pd-manage-reviews/pd-manage-reviews.component';
import { CarouselModule } from 'ngx-owl-carousel-o';
import { NgClass, NgFor, NgIf } from '@angular/common';

@Component({
    selector: 'app-e-product-details',
    standalone: true,
    imports: [RouterLink, QuantityCounterComponent, CarouselModule, NgIf, NgFor, NgClass, PdManageReviewsComponent],
    templateUrl: './e-product-details.component.html',
    styleUrl: './e-product-details.component.scss'
})
export class EProductDetailsComponent {

    // Product Images
    productImages = [
        {
            url: 'images/products/product-details1.jpg'
        },
        {
            url: 'images/products/product-details2.jpg'
        },
        {
            url: 'images/products/product-details3.jpg'
        }
    ]
    selectedImage: string = this.productImages[0].url;
    changeImage(image: string) {
        this.selectedImage = image;
    }

    // Tabs
    currentTab = 'tab1';
    switchTab(event: MouseEvent, tab: string) {
        event.preventDefault();
        this.currentTab = tab;
    }

}