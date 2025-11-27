import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
    selector: 'app-published:not(2)',
    standalone: true,
    imports: [RouterLink],
    templateUrl: './published.component.html',
    styleUrl: './published.component.scss'
})
export class PublishedComponent {}