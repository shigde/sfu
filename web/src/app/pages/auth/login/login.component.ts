import {Component} from '@angular/core';
import {FormControl, FormGroup, NgForm, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {User, SessionService, AuthService} from '@shigde/core';
import {catchError, of, take, tap} from 'rxjs';
import {NgIf} from '@angular/common';
import {map} from 'rxjs/operators';

@Component({
  selector: 'app-login',
  standalone: true,
  templateUrl: './login.component.html',
  imports: [
    ReactiveFormsModule,
    NgIf
  ],
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {

  public fail: boolean = false;
  protected loginForm = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required])
  });

  constructor(
    private router: Router,
    private readonly authService: AuthService,
    private session: SessionService) {
  }

  onSubmit() {
    this.fail = false;
    if (this.loginForm.valid) {
      this.authService.login(`${this.loginForm.value.email}`, `${this.loginForm.value.password}`).pipe(
        (take(1)),
        map(_ => window.location.href = '/dashboard'),
        catchError(_ => this.handleError())
      ).subscribe();
    } else {
      this.loginForm.get('email')?.markAsTouched({onlySelf: true});
      this.loginForm.get('password')?.markAsTouched({onlySelf: true});
    }
  }

  private handleError() {
    this.fail = true;
    return of('');
  }
}
