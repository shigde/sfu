import {Component} from '@angular/core';
import {Router} from '@angular/router';
import {Observable} from 'rxjs';
import {SessionService} from '@shigde/core';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent {
  user$: Observable<string>;
  title = 'shig-web-client';

  constructor(private router: Router, private session: SessionService) {
    this.user$ = session.getUserName();
  }

  logout() {
    this.session.clearData()
    this.router.navigate(['login']);
  }
}
