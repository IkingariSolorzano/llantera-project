import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { ToggleService } from '../common/header/toggle.service';

@Component({
    selector: 'app-not-found',
    standalone: true,
    imports: [RouterLink],
    templateUrl: './not-found.component.html',
    styleUrl: './not-found.component.scss'
})
export class NotFoundComponent {

    constructor(
        public toggleService: ToggleService
    ) {}

    ngOnInit(): void {
        // Initialize theme and direction on component load
        this.toggleService.initializeTheme();
    }

    // Toggle theme between light and dark
    toggleTheme() {
        this.toggleService.toggleTheme();
    }

    // Toggle direction between LTR and RTL
    toggleDirection() {
        this.toggleService.toggleDirection();
    }

}