import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
    selector: 'app-published',
    standalone: true,
    imports: [RouterLink],
    templateUrl: './published.component.html',
    styleUrl: './published.component.scss'
})
export class PublishedComponent {}