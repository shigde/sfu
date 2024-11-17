import {Component} from '@angular/core';
import {SessionService} from '../../../../../../shig-js-sdk/dist/core';
import {Observable} from 'rxjs';
import {AsyncPipe, NgIf} from '@angular/common';
import {Router} from '@angular/router';
import {map} from 'rxjs/operators';
import {AvatarComponent} from '../avatar/avatar.component';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    AsyncPipe,
    NgIf,
    AvatarComponent
  ],
  templateUrl: './header.component.html',
  styleUrl: './header.component.scss'
})
export class HeaderComponent {
  public readonly isUserLogin$: Observable<boolean>;

  constructor(private session: SessionService, private router: Router) {
    this.isUserLogin$ = session.getUserName().pipe(
      map((name) => name !== 'anonymous')
    );
  }

  public logout(): void {
    this.session.clearData()
    this.router.navigate(['login']);
  }
}
