import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import {Observable} from 'rxjs';
import {map} from 'rxjs/operators';
import {SessionService} from '@shigde/core';

@Injectable({providedIn: 'root'})
export class AdminAccessGuard implements CanActivate {
  constructor(private router: Router, private session: SessionService) {
  }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> {
    return this.session.isActive().pipe(
      map(hasSession => {
        if (hasSession) {
          return true;
        }
        this.router.navigate(['/']);
        return false;
      })
    );
  }
}
