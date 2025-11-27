import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { NgClass, NgIf } from '@angular/common';
import { DraftComponent } from './draft/draft.component';
import { PublishedComponent } from './published/published.component';

@Component({
    selector: 'app-e-products-list',
    standalone: true,
    imports: [RouterLink, NgIf, NgClass, PublishedComponent, DraftComponent],
    templateUrl: './e-products-list.component.html',
    styleUrl: './e-products-list.component.scss'
})
export class EProductsListComponent {

    // Tabs
    currentTab = 'tab1';
    switchTab(event: MouseEvent, tab: string) {
        event.preventDefault();
        this.currentTab = tab;
    }

}