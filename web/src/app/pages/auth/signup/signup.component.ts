import {Component} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {AuthService} from '@shigde/core';
import {catchError, of, take, tap} from 'rxjs';
import {NgIf} from '@angular/common';
import {PASSWORD_CONSTRAIN, PasswordValidator} from '../../../validators/password.validator';

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
    user: new FormControl('', [Validators.required, Validators.min(4)]),
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [
        Validators.required,
        Validators.pattern(PASSWORD_CONSTRAIN),
        PasswordValidator.uppercaseLetter,
        PasswordValidator.lowercaseLetter,
        PasswordValidator.digit,
        PasswordValidator.specialCharacter,
        PasswordValidator.minLength
      ]
    ),
    confirmPassword: new FormControl('', [Validators.required]),
    terms: new FormControl('', [Validators.requiredTrue]),
  }, {validators: [PasswordValidator.confirm('confirmPassword', 'password')]});

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
    } else {
      this.signupForm.get('user')?.markAsTouched({onlySelf: true});
      this.signupForm.get('email')?.markAsTouched({onlySelf: true});
      this.signupForm.get('password')?.markAsTouched({onlySelf: true});
      this.signupForm.get('terms')?.markAsTouched({onlySelf: true});
      this.signupForm.get('confirmPassword')?.markAsTouched({onlySelf: true});
    }
  }

  private handleError() {
    this.fail = true;
    return of('');
  }
}
