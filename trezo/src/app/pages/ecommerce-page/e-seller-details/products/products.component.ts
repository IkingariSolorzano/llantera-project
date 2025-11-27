import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgClass, NgIf } from '@angular/common';
import { DraftComponent } from './draft/draft.component';
import { PublishedComponent } from './published/published.component';

@Component({
    selector: 'app-products',
    standalone: true,
    imports: [RouterLink, NgIf, NgClass, PublishedComponent, DraftComponent],
    templateUrl: './products.component.html',
    styleUrl: './products.component.scss'
})
export class ProductsComponent {

    // Tabs
    currentTab = 'tab1';
    switchTab(event: MouseEvent, tab: string) {
        event.preventDefault();
        this.currentTab = tab;
    }

}