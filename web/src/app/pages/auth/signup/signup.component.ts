import {Component} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {AuthService} from '@shigde/core';
import {catchError, of, take, tap} from 'rxjs';
import {NgIf} from '@angular/common';

// https://medium.com/@ojiofor/angular-reactive-forms-strong-password-validation-8dbcce92eb6c

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    NgIf
  ],
  templateUrl: './signup.component.html',
  styleUrl: './signup.component.scss'
})
export class SignupComponent {

  public success = false;
  public fail = false;

  public signupForm = new FormGroup({
    user: new FormControl('', [Validators.required]),
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required])
  });

  constructor(private readonly authService: AuthService, private readonly router: Router) {
  }

  public onSubmit() {
    if (this.signupForm.valid) {

      const user = `${this.signupForm.value.user}`;
      const email = `${this.signupForm.value.email}`;
      const password = `${this.signupForm.value.password}`;
      const account = {user, email, password};

      this.authService.registerAccount(account).pipe(
        take(1),
        tap(a => this.success = true),
        catchError((_) => this.handleError())
      ).subscribe();
    }
  }

  private handleError() {
    this.fail = true;
    return of('');
  }
}
