import {Component} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {AuthService} from '@shigde/core';
import {NgIf} from '@angular/common';
import {catchError, of, take, tap} from 'rxjs';

@Component({
  selector: 'app-password-forgotten-mail',
  standalone: true,
    imports: [
        FormsModule,
        ReactiveFormsModule,
        NgIf
    ],
  templateUrl: './forgot-password-mail.component.html',
  styleUrl: './forgot-password-mail.component.scss'
})
export class ForgotPasswordMailComponent {

  public success = false;

  protected passForgetForm = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
  });

  constructor(private router: Router, private authService: AuthService) {
  }

  onSubmit() {
    this.success = false;
    if (this.passForgetForm.valid) {
      const email = `${this.passForgetForm.value.email}`;
      this.authService.sendForgotPasswordMail(email).pipe(
        take(1),
        tap(a => this.success = true),
        catchError((_) => this.handleError())
      ).subscribe();
    } else  {
      this.passForgetForm.get('email')?.markAsTouched({ onlySelf: true });
    }
  }

  private handleError() {
    this.success = true;
    return of('');
  }
}
