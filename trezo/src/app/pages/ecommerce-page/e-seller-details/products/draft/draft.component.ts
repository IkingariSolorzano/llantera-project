import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
    selector: 'app-draft',
    standalone: true,
    imports: [RouterLink],
    templateUrl: './draft.component.html',
    styleUrl: './draft.component.scss'
})
export class DraftComponent {}